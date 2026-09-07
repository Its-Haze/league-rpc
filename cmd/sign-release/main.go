// Command sign-release is a release-pipeline tool, not a shipped artifact.
package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const signingKeyEnv = "UPDATE_SIGNING_KEY"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "sign-release:", err)
		os.Exit(1)
	}
}

func run() error {
	var artifacts artifactList
	flag.Var(&artifacts, "artifact", "path to a release file to hash and sign; repeat for several")
	sumsOut := flag.String("sums-out", "SHA256SUMS", "path to write the sha256sum-style digest listing")
	sigOut := flag.String("sig-out", "SHA256SUMS.sig", "path to write the detached ed25519 signature listing")
	flag.Parse()

	if len(artifacts) == 0 {
		return fmt.Errorf("-artifact is required")
	}

	priv, err := loadPrivateKey()
	if err != nil {
		return err
	}

	var sums, sigs []string
	for _, artifact := range artifacts {
		digest, err := hashFile(artifact)
		if err != nil {
			return fmt.Errorf("hash %s: %w", artifact, err)
		}
		name := filepath.Base(artifact)

		sums = append(sums, line(hex.EncodeToString(digest), name))
		sigs = append(sigs, line(hex.EncodeToString(ed25519.Sign(priv, digest)), name))

		fmt.Printf("signed %s (sha256 %s)\n", name, hex.EncodeToString(digest))
	}

	if err := writeLines(*sumsOut, sums); err != nil {
		return err
	}
	return writeLines(*sigOut, sigs)
}

// artifactList collects a repeated -artifact flag. The updater reads only the
// line for its own binary, so extra entries are invisible to it.
type artifactList []string

func (a *artifactList) String() string { return strings.Join(*a, ",") }

func (a *artifactList) Set(v string) error {
	if v == "" {
		return fmt.Errorf("artifact path is empty")
	}
	*a = append(*a, v)
	return nil
}

// loadPrivateKey reads and parses the PKCS#8 PEM ed25519 key from the
// signing-key environment variable. It never appears in an error message.
func loadPrivateKey() (ed25519.PrivateKey, error) {
	raw := os.Getenv(signingKeyEnv)
	if raw == "" {
		return nil, fmt.Errorf("%s is not set", signingKeyEnv)
	}
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, fmt.Errorf("%s does not contain a PEM block", signingKeyEnv)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is %T, want ed25519.PrivateKey", key)
	}
	return priv, nil
}

func hashFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// line formats one sha256sum-style entry: "<hex>  <filename>".
func line(hexValue, filename string) string {
	return hexValue + "  " + filename + "\n"
}

func writeLines(path string, lines []string) error {
	return os.WriteFile(path, []byte(strings.Join(lines, "")), 0o644)
}
