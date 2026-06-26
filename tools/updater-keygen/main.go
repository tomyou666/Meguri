// Command updater-keygen は updater 用 Ed25519 鍵ペアを生成する。
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	pubPath := "front/updater-key.pub"
	if len(os.Args) > 1 {
		pubPath = os.Args[1]
	}

	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		fmt.Fprintf(os.Stderr, "generate seed: %v\n", err)
		os.Exit(1)
	}

	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	pkix, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal public key: %v\n", err)
		os.Exit(1)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pkix})

	if err := os.MkdirAll(filepath.Dir(pubPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(pubPath, pubPEM, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", pubPath, err)
		os.Exit(1)
	}

	secret := base64.StdEncoding.EncodeToString(seed)
	fmt.Printf("Wrote public key: %s\n\n", pubPath)
	fmt.Println("Register this value as GitHub Actions secret UPDATER_PRIVATE_KEY")
	fmt.Println("(before deleting any local copy). Do not commit the secret.")
	fmt.Println()
	fmt.Println(secret)
}
