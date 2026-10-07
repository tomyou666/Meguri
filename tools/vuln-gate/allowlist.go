package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// AllowEntry is one accepted finding in vuln-allowlist.yaml.
type AllowEntry struct {
	ID         string   `yaml:"id"`
	Ecosystem  string   `yaml:"ecosystem"`
	Package    string   `yaml:"package"`
	Module     string   `yaml:"module"`
	Via        []string `yaml:"via"`
	Reason     string   `yaml:"reason"`
	Expires    string   `yaml:"expires"`
	expiresDay time.Time
}

type allowFile struct {
	Allows []AllowEntry `yaml:"allows"`
}

func loadAllowlist(path string) ([]AllowEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file allowFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	seen := make(map[string]struct{}, len(file.Allows))
	out := make([]AllowEntry, 0, len(file.Allows))
	for i, e := range file.Allows {
		if err := validateAllowEntry(e); err != nil {
			return nil, fmt.Errorf("allows[%d]: %w", i, err)
		}
		day, err := time.ParseInLocation("2006-01-02", e.Expires, tokyo)
		if err != nil {
			return nil, fmt.Errorf("allows[%d]: expires: %w", i, err)
		}
		e.expiresDay = day
		e.Via = sortedCopy(e.Via)
		key := e.Ecosystem + "\x00" + e.ID + "\x00" + e.Package + "\x00" + e.Module
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("allows[%d]: duplicate id/ecosystem/package/module %s %s %s %s",
				i, e.ID, e.Ecosystem, e.Package, e.Module)
		}
		seen[key] = struct{}{}
		out = append(out, e)
	}
	return out, nil
}

func validateAllowEntry(e AllowEntry) error {
	if strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("id is required")
	}
	switch e.Ecosystem {
	case ecosystemNPM, ecosystemGo:
	default:
		return fmt.Errorf("ecosystem must be npm or go, got %q", e.Ecosystem)
	}
	if strings.TrimSpace(e.Package) == "" {
		return fmt.Errorf("package is required")
	}
	if strings.TrimSpace(e.Module) == "" {
		return fmt.Errorf("module is required")
	}
	if e.Via == nil {
		return fmt.Errorf("via is required (use [] when the module itself is vulnerable)")
	}
	if strings.TrimSpace(e.Reason) == "" {
		return fmt.Errorf("reason is required")
	}
	if strings.TrimSpace(e.Expires) == "" {
		return fmt.Errorf("expires is required")
	}
	return nil
}

var tokyo = mustTokyo()

func mustTokyo() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		// Fixed offset keeps tests and CI deterministic when tzdata is missing.
		return time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	return loc
}

// allowExpiresOn reports whether the entry is still valid at now in Asia/Tokyo.
// The expires calendar day is inclusive.
func allowExpiresOn(e AllowEntry, now time.Time) bool {
	today := now.In(tokyo)
	y, m, d := today.Date()
	todayDay := time.Date(y, m, d, 0, 0, 0, 0, tokyo)
	return !todayDay.After(e.expiresDay)
}
