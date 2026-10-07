package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMatchAllowed(t *testing.T) {
	allows := []AllowEntry{{
		ID:         "GHSA-vfj7-8cjw-p6xm",
		Ecosystem:  ecosystemNPM,
		Package:    "braces",
		Module:     "front/frontend",
		Via:        []string{"shadcn"},
		Reason:     "test",
		Expires:    "2027-01-07",
		expiresDay: mustDay(t, "2027-01-07"),
	}}
	findings := []Finding{{
		ID:        "GHSA-vfj7-8cjw-p6xm",
		Ecosystem: ecosystemNPM,
		Package:   "braces",
		Module:    "front/frontend",
		Via:       []string{"shadcn"},
	}}
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, tokyo)
	res := matchFindings(allows, findings, now)
	if !res.OK {
		t.Fatalf("expected ok, got %v", res.Messages)
	}
}

func TestMatchViaMismatch(t *testing.T) {
	allows := []AllowEntry{{
		ID:         "GHSA-vfj7-8cjw-p6xm",
		Ecosystem:  ecosystemNPM,
		Package:    "braces",
		Module:     "front/frontend",
		Via:        []string{"shadcn"},
		Expires:    "2027-01-07",
		expiresDay: mustDay(t, "2027-01-07"),
	}}
	findings := []Finding{{
		ID:        "GHSA-vfj7-8cjw-p6xm",
		Ecosystem: ecosystemNPM,
		Package:   "braces",
		Module:    "front/frontend",
		Via:       []string{"shadcn", "micromatch"},
	}}
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, tokyo)
	res := matchFindings(allows, findings, now)
	if res.OK {
		t.Fatal("expected failure for via mismatch")
	}
}

func TestMatchExpiredInclusiveBoundary(t *testing.T) {
	allows := []AllowEntry{{
		ID:         "GHSA-vfj7-8cjw-p6xm",
		Ecosystem:  ecosystemNPM,
		Package:    "braces",
		Module:     "front/frontend",
		Via:        []string{"shadcn"},
		Expires:    "2027-01-07",
		expiresDay: mustDay(t, "2027-01-07"),
	}}
	findings := []Finding{{
		ID:        "GHSA-vfj7-8cjw-p6xm",
		Ecosystem: ecosystemNPM,
		Package:   "braces",
		Module:    "front/frontend",
		Via:       []string{"shadcn"},
	}}

	onDay := time.Date(2027, 1, 7, 23, 59, 0, 0, tokyo)
	if res := matchFindings(allows, findings, onDay); !res.OK {
		t.Fatalf("expires day should still pass: %v", res.Messages)
	}
	nextDay := time.Date(2027, 1, 8, 0, 0, 0, 0, tokyo)
	if res := matchFindings(allows, findings, nextDay); res.OK {
		t.Fatal("day after expires should fail")
	}
}

func TestMatchStaleAllow(t *testing.T) {
	allows := []AllowEntry{{
		ID:         "GHSA-vfj7-8cjw-p6xm",
		Ecosystem:  ecosystemNPM,
		Package:    "braces",
		Module:     "front/frontend",
		Via:        []string{"shadcn"},
		Expires:    "2027-01-07",
		expiresDay: mustDay(t, "2027-01-07"),
	}}
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, tokyo)
	res := matchFindings(allows, nil, now)
	if res.OK {
		t.Fatal("expected stale allowlist failure")
	}
}

func TestLoadAllowlist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vuln-allowlist.yaml")
	content := `
allows:
  - id: GHSA-vfj7-8cjw-p6xm
    ecosystem: npm
    package: braces
    module: front/frontend
    via: [shadcn]
    reason: test reason
    expires: 2027-01-07
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	allows, err := loadAllowlist(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(allows) != 1 {
		t.Fatalf("len=%d", len(allows))
	}
	if allows[0].ID != "GHSA-vfj7-8cjw-p6xm" {
		t.Errorf("id=%q", allows[0].ID)
	}
}

func mustDay(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation("2006-01-02", s, tokyo)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
