package main

import (
	"fmt"
	"strings"
	"time"
)

// GateResult is the outcome of matching findings against the allowlist.
type GateResult struct {
	OK       bool
	Messages []string
}

func matchFindings(allows []AllowEntry, findings []Finding, now time.Time) GateResult {
	used := make([]bool, len(allows))
	var msgs []string

	for _, f := range findings {
		idx := -1
		for i, a := range allows {
			if a.Ecosystem == f.Ecosystem && a.ID == f.ID && a.Package == f.Package && a.Module == f.Module {
				idx = i
				break
			}
		}
		if idx < 0 {
			msgs = append(msgs, "unallowed: "+formatFinding(f))
			continue
		}
		a := allows[idx]
		if !sameStringSet(a.Via, f.Via) {
			msgs = append(msgs, fmt.Sprintf("via mismatch for %s: allowlist=%v finding=%v",
				formatFinding(f), sortedCopy(a.Via), sortedCopy(f.Via)))
			continue
		}
		if !allowExpiresOn(a, now) {
			msgs = append(msgs, fmt.Sprintf("expired allowlist entry %s (expires %s): %s",
				a.ID, a.Expires, formatFinding(f)))
			continue
		}
		used[idx] = true
	}

	for i, a := range allows {
		if used[i] {
			continue
		}
		msgs = append(msgs, fmt.Sprintf("stale allowlist entry: %s %s package=%s module=%s via=[%s] expires=%s",
			a.Ecosystem, a.ID, a.Package, a.Module, strings.Join(a.Via, ", "), a.Expires))
	}

	return GateResult{OK: len(msgs) == 0, Messages: msgs}
}
