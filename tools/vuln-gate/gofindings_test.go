package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGovulncheckIndirect(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "govulncheck_indirect.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	direct := map[string]struct{}{
		"example.com/mid":  {},
		"golang.org/x/net": {},
	}
	findings, err := parseGovulncheckFindings(raw, "tools/example", "meguri/example", direct)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings: %+v", len(findings), findings)
	}
	f := findings[0]
	if f.ID != "GO-2024-9999" {
		t.Errorf("id = %q", f.ID)
	}
	if f.Package != "example.com/vuln" {
		t.Errorf("package = %q", f.Package)
	}
	if f.Module != "tools/example" {
		t.Errorf("module = %q", f.Module)
	}
	if !sameStringSet(f.Via, []string{"example.com/mid"}) {
		t.Errorf("via = %v, want [example.com/mid]", f.Via)
	}
}

func TestParseGovulncheckSelfModule(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "govulncheck_self.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	direct := map[string]struct{}{"example.com/mid": {}}
	findings, err := parseGovulncheckFindings(raw, "tools/example", "meguri/example", direct)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings: %+v", len(findings), findings)
	}
	if len(findings[0].Via) != 0 {
		t.Errorf("via = %v, want empty", findings[0].Via)
	}
	if findings[0].Package != "meguri/example" {
		t.Errorf("package = %q", findings[0].Package)
	}
}
