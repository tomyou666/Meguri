package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseNPMFindingsBraces(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "npm_braces_audit.json"))
	if err != nil {
		t.Fatal(err)
	}
	direct := map[string]struct{}{
		"shadcn": {},
		"react":  {},
	}
	findings, err := parseNPMFindings(raw, "front/frontend", direct)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1 (moderate left-pad ignored): %+v", len(findings), findings)
	}
	f := findings[0]
	if f.ID != "GHSA-vfj7-8cjw-p6xm" {
		t.Errorf("id = %q", f.ID)
	}
	if f.Package != "braces" {
		t.Errorf("package = %q", f.Package)
	}
	if f.Module != "front/frontend" {
		t.Errorf("module = %q", f.Module)
	}
	if !sameStringSet(f.Via, []string{"shadcn"}) {
		t.Errorf("via = %v, want [shadcn]", f.Via)
	}
}

func TestParseNPMFindingsExtraDirectRoot(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "npm_braces_audit.json"))
	if err != nil {
		t.Fatal(err)
	}
	// Pretend micromatch is also a direct dependency: via must include both.
	direct := map[string]struct{}{
		"shadcn":     {},
		"micromatch": {},
	}
	findings, err := parseNPMFindings(raw, "front/frontend", direct)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings: %+v", len(findings), findings)
	}
	if !sameStringSet(findings[0].Via, []string{"micromatch", "shadcn"}) {
		t.Errorf("via = %v", findings[0].Via)
	}
}

func TestExtractNPMAuditJSONWithPrefix(t *testing.T) {
	raw := []byte("{\"type\":\"message\",\"message\":\"noise\"}\n" +
		string(mustRead(t, filepath.Join("testdata", "npm_braces_audit.json"))))
	body, err := extractNPMAuditJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	direct := map[string]struct{}{"shadcn": {}}
	findings, err := parseNPMFindings(body, "front/frontend", direct)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings", len(findings))
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
