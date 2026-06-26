// Command sign-release は Release 用 SHA256SUMS を生成し、任意で Ed25519 署名する。
package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: sign-release <bin-dir>")
		os.Exit(2)
	}
	loadDotEnv()
	dir := os.Args[1]
	files := []string{
		"meguri-amd64_arm64-installer.exe",
		"meguri-windows-amd64.zip",
		"meguri-windows-arm64.zip",
	}
	var lines []string
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
			os.Exit(1)
		}
		sum := sha256.Sum256(data)
		lines = append(lines, fmt.Sprintf("%x  %s", sum, name))
	}
	if len(lines) == 0 {
		fmt.Fprintf(os.Stderr, "no release artifacts found in %s\n", dir)
		os.Exit(1)
	}
	sums := strings.Join(lines, "\n") + "\n"
	sumsPath := filepath.Join(dir, "SHA256SUMS")
	if err := os.WriteFile(sumsPath, []byte(sums), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", sumsPath, err)
		os.Exit(1)
	}
	fmt.Println(sumsPath)

	pemKey := strings.TrimSpace(os.Getenv("UPDATER_PRIVATE_KEY"))
	if pemKey == "" {
		return
	}
	priv, err := parseEd25519Private(pemKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse UPDATER_PRIVATE_KEY: %v\n", err)
		os.Exit(1)
	}
	digest := sha256.Sum256([]byte(sums))
	sig := ed25519.Sign(priv, digest[:])
	sigPath := filepath.Join(dir, "SHA256SUMS.sig")
	encoded := base64.StdEncoding.EncodeToString(sig)
	if err := os.WriteFile(sigPath, []byte(encoded+"\n"), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", sigPath, err)
		os.Exit(1)
	}
	fmt.Println(sigPath)
}

func parseEd25519Private(raw string) (ed25519.PrivateKey, error) {
	if block, _ := pem.Decode([]byte(raw)); block != nil {
		if len(block.Bytes) == ed25519.PrivateKeySize {
			return ed25519.PrivateKey(block.Bytes), nil
		}
		if len(block.Bytes) == ed25519.SeedSize {
			return ed25519.NewKeyFromSeed(block.Bytes), nil
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err == nil && len(decoded) == ed25519.SeedSize {
		return ed25519.NewKeyFromSeed(decoded), nil
	}
	return nil, fmt.Errorf("expected PEM or base64 ed25519 seed")
}
