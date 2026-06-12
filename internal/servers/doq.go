package servers

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"

	"github.com/blkhole-sh/blkhole/internal/services"
	"github.com/miekg/dns"
	"github.com/quic-go/quic-go"
)

// DoQ application error codes (RFC 9250 section 8.4).
const (
	doqInternalError quic.ApplicationErrorCode = 1
	doqProtocolError quic.ApplicationErrorCode = 2
)

func StartDoQ(ctx context.Context, resolver services.Resolver, domain string, tlsConfig *tls.Config) error {
	// RFC 9250 section 4.1: ALPN token "doq"
	tlsConf := tlsConfig.Clone()
	tlsConf.NextProtos = []string{"doq"}

	listener, err := quic.ListenAddr(":853", tlsConf, nil)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down doq server...")
		listener.Close()
	}()

	log.Println("starting doq server on udp :853")
	err = serveDoQ(ctx, listener, resolver, domain)
	if ctx.Err() != nil {
		return nil
	}
	return err
}

func serveDoQ(ctx context.Context, listener *quic.Listener, resolver services.Resolver, domain string) error {
	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			return err
		}
		go handleDoQConn(ctx, conn, resolver, domain)
	}
}

func handleDoQConn(ctx context.Context, conn *quic.Conn, resolver services.Resolver, domain string) {
	// Get device hash from the connection's TLS SNI
	deviceHash := deviceHashFromServerName(conn.ConnectionState().TLS.ServerName, domain)

	// RFC 9250 section 4.2: each query/response pair is one bidirectional stream
	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			return
		}
		go handleDoQStream(conn, stream, resolver, deviceHash)
	}
}

func handleDoQStream(conn *quic.Conn, stream *quic.Stream, resolver services.Resolver, deviceHash string) {
	defer stream.Close()

	// Messages are prefixed with a 2-byte length field (RFC 9250 section 4.2)
	var length uint16
	if err := binary.Read(stream, binary.BigEndian, &length); err != nil {
		return
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return
	}

	msg := new(dns.Msg)
	if err := msg.Unpack(buf); err != nil {
		conn.CloseWithError(doqProtocolError, "malformed dns message")
		return
	}

	// RFC 9250 section 4.2.1: the DNS message ID must be zero
	if msg.Id != 0 {
		conn.CloseWithError(doqProtocolError, "dns message id must be zero")
		return
	}

	// Resolve DNS query
	resp, err := resolver.Resolve(msg, deviceHash)
	if err != nil {
		log.Printf("error resolving dns query: %v", err)
	}

	// If no response is available, reply with SERVFAIL
	if resp == nil {
		resp = new(dns.Msg)
		resp.SetRcode(msg, dns.RcodeServerFailure)
	}
	resp.Id = 0

	respBytes, err := resp.Pack()
	if err != nil {
		log.Printf("failed to pack dns response: %v", err)
		conn.CloseWithError(doqInternalError, "failed to pack dns response")
		return
	}
	if len(respBytes) > 0xFFFF {
		log.Printf("dns response too large for doq frame: %d bytes", len(respBytes))
		conn.CloseWithError(doqInternalError, "dns response too large")
		return
	}

	if err := binary.Write(stream, binary.BigEndian, uint16(len(respBytes))); err != nil {
		return
	}
	if _, err := stream.Write(respBytes); err != nil {
		log.Printf("failed to write dns response: %v", err)
	}
}
