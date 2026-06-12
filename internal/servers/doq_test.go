package servers

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/quic-go/quic-go"
)

// mockResolver records the device hash it was called with. Resolve runs on a
// server goroutine while the tests read the hash, hence the mutex.
type mockResolver struct {
	mu         sync.Mutex
	deviceHash string
}

func (m *mockResolver) Resolve(msg *dns.Msg, deviceHash string) (*dns.Msg, error) {
	m.mu.Lock()
	m.deviceHash = deviceHash
	m.mu.Unlock()
	resp := new(dns.Msg)
	resp.SetReply(msg)
	rr, _ := dns.NewRR(msg.Question[0].Name + " 300 IN A 127.0.0.1")
	resp.Answer = append(resp.Answer, rr)
	return resp, nil
}

func (m *mockResolver) DeviceHash() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deviceHash
}

func selfSignedTLSConfig(t *testing.T) *tls.Config {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.com"},
		DNSNames:     []string{"example.com", "*.example.com"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
	}
}

// startTestDoQ starts a DoQ server on an ephemeral UDP port and returns its address.
func startTestDoQ(t *testing.T, resolver *mockResolver) string {
	t.Helper()
	tlsConf := selfSignedTLSConfig(t)
	tlsConf.NextProtos = []string{"doq"}
	listener, err := quic.ListenAddr("127.0.0.1:0", tlsConf, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go serveDoQ(ctx, listener, resolver, "example.com")

	return listener.Addr().String()
}

func dialTestDoQ(t *testing.T, addr, serverName string) *quic.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := quic.DialAddr(ctx, addr, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         serverName,
		NextProtos:         []string{"doq"},
	}, nil)
	if err != nil {
		t.Fatalf("failed to dial doq server: %v", err)
	}
	t.Cleanup(func() { conn.CloseWithError(0, "") })
	return conn
}

// exchange sends a DoQ query on a new stream and returns the response.
func exchange(conn *quic.Conn, msg *dns.Msg) (*dns.Msg, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}

	query, err := msg.Pack()
	if err != nil {
		return nil, err
	}
	if err := binary.Write(stream, binary.BigEndian, uint16(len(query))); err != nil {
		return nil, err
	}
	if _, err := stream.Write(query); err != nil {
		return nil, err
	}
	// Signal end of query with FIN (RFC 9250 section 4.2)
	if err := stream.Close(); err != nil {
		return nil, err
	}

	stream.SetReadDeadline(time.Now().Add(5 * time.Second))
	var length uint16
	if err := binary.Read(stream, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return nil, err
	}

	resp := new(dns.Msg)
	if err := resp.Unpack(buf); err != nil {
		return nil, err
	}
	return resp, nil
}

func TestDoQResolvesQuery(t *testing.T) {
	resolver := &mockResolver{}
	addr := startTestDoQ(t, resolver)
	conn := dialTestDoQ(t, addr, "abc123.example.com")

	msg := new(dns.Msg)
	msg.SetQuestion("test.example.org.", dns.TypeA)
	msg.Id = 0

	resp, err := exchange(conn, msg)
	if err != nil {
		t.Fatalf("doq exchange failed: %v", err)
	}
	if resp.Rcode != dns.RcodeSuccess {
		t.Errorf("expected NOERROR, got %s", dns.RcodeToString[resp.Rcode])
	}
	if len(resp.Answer) != 1 {
		t.Errorf("expected 1 answer, got %d", len(resp.Answer))
	}
	if resp.Id != 0 {
		t.Errorf("expected response message id 0, got %d", resp.Id)
	}
	if hash := resolver.DeviceHash(); hash != "abc123" {
		t.Errorf("expected device hash 'abc123' from SNI, got %q", hash)
	}
}

func TestDoQBareDomainHasNoDeviceHash(t *testing.T) {
	resolver := &mockResolver{deviceHash: "sentinel"}
	addr := startTestDoQ(t, resolver)
	conn := dialTestDoQ(t, addr, "example.com")

	msg := new(dns.Msg)
	msg.SetQuestion("test.example.org.", dns.TypeA)
	msg.Id = 0

	if _, err := exchange(conn, msg); err != nil {
		t.Fatalf("doq exchange failed: %v", err)
	}
	if hash := resolver.DeviceHash(); hash != "" {
		t.Errorf("expected empty device hash for bare domain, got %q", hash)
	}
}

func TestDoQRejectsNonZeroMessageID(t *testing.T) {
	resolver := &mockResolver{}
	addr := startTestDoQ(t, resolver)
	conn := dialTestDoQ(t, addr, "example.com")

	msg := new(dns.Msg)
	msg.SetQuestion("test.example.org.", dns.TypeA)
	msg.Id = 42

	_, err := exchange(conn, msg)
	if err == nil {
		t.Fatal("expected error for non-zero message id, got response")
	}
	var appErr *quic.ApplicationError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected quic application error, got %v", err)
	}
	if appErr.ErrorCode != doqProtocolError {
		t.Errorf("expected DOQ_PROTOCOL_ERROR (2), got %d", appErr.ErrorCode)
	}
}
