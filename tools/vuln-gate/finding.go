package main

import (
	"fmt"
	"sort"
	"strings"
)

const (
	ecosystemNPM = "npm"
	ecosystemGo  = "go"
)

// Finding is one advisory observed in one module, with the direct-dependency
// entry points that reach it.
type Finding struct {
	ID        string
	Ecosystem string
	Package   string
	Module    string
	Via       []string
	Severity  string // npm only; empty for go
}

func sortedCopy(ss []string) []string {
	out := append([]string(nil), ss...)
	sort.Strings(out)
	return out
}

func sameStringSet(a, b []string) bool {
	a = sortedCopy(a)
	b = sortedCopy(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func formatFinding(f Finding) string {
	via := "(empty)"
	if len(f.Via) > 0 {
		via = strings.Join(sortedCopy(f.Via), ", ")
	}
	return fmt.Sprintf("%s %s package=%s module=%s via=[%s]",
		f.Ecosystem, f.ID, f.Package, f.Module, via)
}
