package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func generateKeyPEM(t *testing.T) (ed25519.PublicKey, string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	return pub, string(pem.EncodeToMemory(block))
}

func TestLoadPrivateKey_MissingEnv(t *testing.T) {
	t.Setenv(signingKeyEnv, "")
	if _, err := loadPrivateKey(); err == nil {
		t.Fatal("expected an error when the signing key env var is unset")
	}
}

func TestLoadPrivateKey_ParsesPKCS8PEM(t *testing.T) {
	pub, keyPEM := generateKeyPEM(t)
	t.Setenv(signingKeyEnv, keyPEM)

	priv, err := loadPrivateKey()
	if err != nil {
		t.Fatalf("loadPrivateKey: %v", err)
	}
	if !bytes.Equal(priv.Public().(ed25519.PublicKey), pub) {
		t.Fatal("parsed private key does not match the generated public key")
	}
}

func TestRun_WritesSumsAndSignature(t *testing.T) {
	pub, keyPEM := generateKeyPEM(t)
	t.Setenv(signingKeyEnv, keyPEM)

	dir := t.TempDir()
	artifact := filepath.Join(dir, "league-rpc-gui.exe")
	if err := os.WriteFile(artifact, []byte("pretend-exe-bytes"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	sumsOut := filepath.Join(dir, "SHA256SUMS")
	sigOut := filepath.Join(dir, "SHA256SUMS.sig")

	os.Args = []string{"sign-release", "-artifact", artifact, "-sums-out", sumsOut, "-sig-out", sigOut}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	if err := run(); err != nil {
		t.Fatalf("run: %v", err)
	}

	sums, err := os.ReadFile(sumsOut)
	if err != nil {
		t.Fatalf("read SHA256SUMS: %v", err)
	}
	wantDigest := sha256.Sum256([]byte("pretend-exe-bytes"))
	wantLine := hex.EncodeToString(wantDigest[:]) + "  league-rpc-gui.exe\n"
	if string(sums) != wantLine {
		t.Fatalf("SHA256SUMS = %q, want %q", sums, wantLine)
	}

	sig, err := os.ReadFile(sigOut)
	if err != nil {
		t.Fatalf("read SHA256SUMS.sig: %v", err)
	}
	fields := strings.Fields(string(sig))
	if len(fields) != 2 || fields[1] != "league-rpc-gui.exe" {
		t.Fatalf("SHA256SUMS.sig = %q, want '<hex>  league-rpc-gui.exe'", sig)
	}
	sigBytes, err := hex.DecodeString(fields[0])
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	if !ed25519.Verify(pub, wantDigest[:], sigBytes) {
		t.Fatal("signature does not verify against the digest")
	}
}
