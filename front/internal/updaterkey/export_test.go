package updaterkey

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

// TestExportDevPublicKey は開発用固定シードから公開鍵 PEM を出力する（手動実行用）。
func TestExportDevPublicKey(t *testing.T) {
	t.Skip("manual: go test ./internal/updaterkey -run TestExportDevPublicKey -v")
	seed := make([]byte, ed25519.SeedSize)
	seed[len(seed)-1] = 1
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	pkix, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: pkix}
	t.Logf("\n%s", pem.EncodeToMemory(block))
}
