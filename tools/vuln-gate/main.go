// Command vuln-gate runs npm audit and govulncheck, then matches findings
// against vuln-allowlist.yaml.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getwd: %v\n", err)
		os.Exit(2)
	}
	if len(os.Args) > 1 {
		repoRoot = os.Args[1]
	}
	repoRoot, err = filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "abs: %v\n", err)
		os.Exit(2)
	}

	allowPath := filepath.Join(repoRoot, "vuln-allowlist.yaml")
	allows, err := loadAllowlist(allowPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "allowlist: %v\n", err)
		os.Exit(2)
	}

	cfg := defaultScannerConfig(repoRoot)
	findings, err := collectFindings(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan: %v\n", err)
		os.Exit(2)
	}

	result := matchFindings(allows, findings, time.Now())
	for _, f := range findings {
		fmt.Println("finding:", formatFinding(f))
	}
	if result.OK {
		fmt.Println("vuln-gate: ok")
		os.Exit(0)
	}
	for _, m := range result.Messages {
		fmt.Fprintln(os.Stderr, m)
	}
	fmt.Fprintln(os.Stderr, "vuln-gate: failed")
	os.Exit(1)
}
