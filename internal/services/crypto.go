package services

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2s"
)

// CryptoService defines the interface for cryptographic operations
type CryptoService interface {
	RandomHash() (string, error)
	HashPassword(password string) (string, error)
	VerifyPassword(password, hashedPassword string) (bool, error)
}

// cryptoService implements the CryptoService interface
type cryptoService struct {
	secret []byte
}

// NewCryptoService creates a new CryptoService instance
func NewCryptoService(secret []byte) CryptoService {
	return &cryptoService{secret: secret}
}

func (cs *cryptoService) RandomHash() (string, error) {
	// Create a random value
	r := make([]byte, 16) // 16 bytes of randomness
	_, err := rand.Read(r)
	if err != nil {
		return "", err
	}

	// Create a keyed BLAKE2s hash
	hasher, err := blake2s.New256(cs.secret)
	if err != nil {
		return "", err
	}

	// Write data to the hasher
	hasher.Write(cs.secret)

	// Write random value to the hasher
	hasher.Write(r)

	// Get the hash (binary)
	hash := hasher.Sum(nil)

	// Encode the hash in Base32
	base32EncodedHash := base32.StdEncoding.EncodeToString(hash)

	return base32EncodedHash, nil
}

// HashPassword hashes the given password using Argon2 and returns the hash encoded as a string with salt
func (cs *cryptoService) HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, 16) // Recommended size is 16 bytes
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Set Argon2 parameters
	time := uint32(1)           // Number of iterations
	memory := uint32(64 * 1024) // Memory in KiB
	threads := uint8(4)         // Number of threads
	keyLen := uint32(32)        // Desired length of the resulting key

	// Hash the password with Argon2id
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	// Encode the hash and salt as a single string (e.g., base64)
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	// Combine the salt and hash for storage
	fullHash := fmt.Sprintf("%s$%s", encodedSalt, encodedHash)
	return fullHash, nil
}

// VerifyPassword verifies a password against a hashed password created by HashPassword
func (cs *cryptoService) VerifyPassword(password, hashedPassword string) (bool, error) {
	// Split the stored hash into salt and hash parts
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid hash format")
	}

	// Decode the salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Use the same Argon2 parameters as in HashPassword
	time := uint32(1)           // Number of iterations
	memory := uint32(64 * 1024) // Memory in KiB
	threads := uint8(4)         // Number of threads
	keyLen := uint32(32)        // Desired length of the resulting key

	// Hash the provided password with the extracted salt
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	// Compare the computed hash with the expected hash
	if len(hash) != len(expectedHash) {
		return false, nil
	}

	// Constant-time comparison to prevent timing attacks
	for i := range hash {
		if hash[i] != expectedHash[i] {
			return false, nil
		}
	}

	return true, nil
}
