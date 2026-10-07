package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var ghsaURLRe = regexp.MustCompile(`(?i)GHSA-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}`)

type npmAuditReport struct {
	AuditReportVersion int                    `json:"auditReportVersion"`
	Vulnerabilities    map[string]npmVulnNode `json:"vulnerabilities"`
}

type npmVulnNode struct {
	Name     string        `json:"name"`
	Severity string        `json:"severity"`
	Via      []npmViaEntry `json:"via"`
	Effects  []string      `json:"effects"`
}

// npmViaEntry is either a dependency name (string) or an advisory object.
type npmViaEntry struct {
	DepName string
	URL     string
	Title   string
	Name    string
}

func (v *npmViaEntry) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		v.DepName = s
		return nil
	}
	var obj struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	v.URL = obj.URL
	v.Title = obj.Title
	v.Name = obj.Name
	return nil
}

type npmPackageJSON struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

func extractNPMAuditJSON(raw []byte) ([]byte, error) {
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	for {
		var msg json.RawMessage
		if err := dec.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// Fall through to substring search when decoder hits trailing noise.
			break
		}
		var probe struct {
			AuditReportVersion int `json:"auditReportVersion"`
		}
		if err := json.Unmarshal(msg, &probe); err == nil && probe.AuditReportVersion > 0 {
			return msg, nil
		}
	}
	idx := bytes.Index(raw, []byte(`"auditReportVersion"`))
	if idx < 0 {
		return nil, fmt.Errorf("npm audit JSON: auditReportVersion not found")
	}
	start := bytes.LastIndexByte(raw[:idx], '{')
	if start < 0 {
		return nil, fmt.Errorf("npm audit JSON: object start not found")
	}
	dec = json.NewDecoder(bytes.NewReader(raw[start:]))
	var msg json.RawMessage
	if err := dec.Decode(&msg); err != nil {
		return nil, fmt.Errorf("npm audit JSON: %w", err)
	}
	return msg, nil
}

func parseNPMFindings(auditJSON []byte, modulePath string, directDeps map[string]struct{}) ([]Finding, error) {
	body, err := extractNPMAuditJSON(auditJSON)
	if err != nil {
		return nil, err
	}
	var report npmAuditReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, fmt.Errorf("npm audit JSON: %w", err)
	}

	// dependents[p] = packages that list p in their dependency chain toward the root
	// (i.e. effects of p).
	dependents := make(map[string][]string, len(report.Vulnerabilities))
	for name, node := range report.Vulnerabilities {
		pkg := node.Name
		if pkg == "" {
			pkg = name
		}
		dependents[pkg] = append(dependents[pkg], node.Effects...)
	}

	type advisoryHit struct {
		id       string
		severity string
	}
	// package -> advisory id (first wins; npm reports one GHSA per leaf package)
	advisories := make(map[string]advisoryHit)
	for name, node := range report.Vulnerabilities {
		pkg := node.Name
		if pkg == "" {
			pkg = name
		}
		for _, via := range node.Via {
			if via.URL == "" && via.Title == "" {
				continue
			}
			id := ghsaFromURL(via.URL)
			if id == "" {
				id = ghsaFromURL(via.Title)
			}
			if id == "" {
				continue
			}
			advisories[pkg] = advisoryHit{id: id, severity: strings.ToLower(node.Severity)}
		}
	}

	merged := make(map[string]*Finding)
	for pkg, hit := range advisories {
		if !severityAtLeastHigh(hit.severity) {
			continue
		}
		via := npmDirectRoots(pkg, dependents, directDeps)
		key := ecosystemNPM + "\x00" + hit.id + "\x00" + pkg + "\x00" + modulePath
		if existing, ok := merged[key]; ok {
			existing.Via = unionStrings(existing.Via, via)
			continue
		}
		merged[key] = &Finding{
			ID:        hit.id,
			Ecosystem: ecosystemNPM,
			Package:   pkg,
			Module:    modulePath,
			Via:       via,
			Severity:  hit.severity,
		}
	}

	out := make([]Finding, 0, len(merged))
	for _, f := range merged {
		f.Via = sortedCopy(f.Via)
		out = append(out, *f)
	}
	return out, nil
}

func severityAtLeastHigh(sev string) bool {
	switch strings.ToLower(sev) {
	case "high", "critical":
		return true
	default:
		return false
	}
}

func ghsaFromURL(s string) string {
	m := ghsaURLRe.FindString(s)
	if m == "" {
		return ""
	}
	return strings.ToUpper(m[:4]) + strings.ToLower(m[4:])
}

// npmDirectRoots walks effects upward from the advisory package and returns
// the set of package.json direct dependencies that reach it.
func npmDirectRoots(advisoryPkg string, dependents map[string][]string, directDeps map[string]struct{}) []string {
	if _, ok := directDeps[advisoryPkg]; ok {
		// Direct dependency is itself vulnerable: it is the sole entry point
		// unless other direct deps also reach it through the tree.
	}
	reached := map[string]struct{}{advisoryPkg: {}}
	queue := []string{advisoryPkg}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, parent := range dependents[cur] {
			if _, seen := reached[parent]; seen {
				continue
			}
			reached[parent] = struct{}{}
			queue = append(queue, parent)
		}
	}
	var via []string
	for name := range reached {
		if _, ok := directDeps[name]; ok {
			via = append(via, name)
		}
	}
	return via
}

func loadNPMDirectDeps(packageJSONPath string) (map[string]struct{}, error) {
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, err
	}
	var pkg npmPackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("package.json: %w", err)
	}
	out := make(map[string]struct{})
	for name := range pkg.Dependencies {
		out[name] = struct{}{}
	}
	for name := range pkg.DevDependencies {
		out[name] = struct{}{}
	}
	for name := range pkg.OptionalDependencies {
		out[name] = struct{}{}
	}
	return out, nil
}

func unionStrings(a, b []string) []string {
	set := make(map[string]struct{}, len(a)+len(b))
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		set[s] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	return out
}
