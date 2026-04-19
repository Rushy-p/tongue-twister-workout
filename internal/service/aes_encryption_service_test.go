package service

import (
	"bytes"
	"testing"
)

func newTestService(t *testing.T) *AESEncryptionService {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	svc, err := NewAESEncryptionService(key)
	if err != nil {
		t.Fatalf("NewAESEncryptionService: %v", err)
	}
	return svc
}

// TestEncryptDecryptRoundTrip verifies that decrypting an encrypted value returns the original plaintext.
func TestEncryptDecryptRoundTrip(t *testing.T) {
	svc := newTestService(t)
	plaintext := []byte("hello, speech practice app!")

	encrypted, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	decrypted, err := svc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

// TestEncryptNonceRandomness verifies that encrypting the same plaintext twice produces different ciphertext.
func TestEncryptNonceRandomness(t *testing.T) {
	svc := newTestService(t)
	plaintext := []byte("same plaintext")

	enc1, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt (1): %v", err)
	}
	enc2, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt (2): %v", err)
	}

	if bytes.Equal(enc1, enc2) {
		t.Error("expected different ciphertext for same plaintext (nonce should be random)")
	}
}

// TestWrongKeyLengthReturnsError verifies that a key that is not 32 bytes is rejected.
func TestWrongKeyLengthReturnsError(t *testing.T) {
	cases := []int{0, 16, 24, 31, 33, 64}
	for _, n := range cases {
		key := make([]byte, n)
		_, err := NewAESEncryptionService(key)
		if err == nil {
			t.Errorf("expected error for key length %d, got nil", n)
		}
	}
}

// TestPrepareAndDecryptFromSync verifies the PrepareForSync / DecryptFromSync helpers.
func TestPrepareAndDecryptFromSync(t *testing.T) {
	svc := newTestService(t)
	original := []byte(`{"session_id":"abc","data":"test"}`)

	payload, err := PrepareForSync(svc, original)
	if err != nil {
		t.Fatalf("PrepareForSync: %v", err)
	}
	if payload.Algorithm != "AES-256-GCM" {
		t.Errorf("unexpected algorithm: %s", payload.Algorithm)
	}

	recovered, err := DecryptFromSync(svc, payload)
	if err != nil {
		t.Fatalf("DecryptFromSync: %v", err)
	}
	if !bytes.Equal(original, recovered) {
		t.Errorf("sync round-trip mismatch: got %q, want %q", recovered, original)
	}
}
