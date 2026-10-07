package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// Minimal govulncheck JSON message shapes (protocol v1).
type govulnMessage struct {
	OSV     *govulnOSV     `json:"osv,omitempty"`
	Finding *govulnFinding `json:"finding,omitempty"`
}

type govulnOSV struct {
	ID      string   `json:"id"`
	Aliases []string `json:"aliases"`
}

type govulnFinding struct {
	OSV   string        `json:"osv"`
	Trace []govulnFrame `json:"trace"`
}

type govulnFrame struct {
	Module  string `json:"module"`
	Package string `json:"package,omitempty"`
}

func parseGovulncheckFindings(raw []byte, moduleDir, modulePath string, directDeps map[string]struct{}) ([]Finding, error) {
	merged := make(map[string]*Finding)
	sc := bufio.NewScanner(bytes.NewReader(raw))
	// Findings can be large; raise the token limit.
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var msg govulnMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			// Non-JSON progress lines on stderr are not expected on stdout;
			// ignore undecodable lines that are clearly not messages.
			continue
		}
		if msg.Finding == nil {
			continue
		}
		f := msg.Finding
		if f.OSV == "" || len(f.Trace) == 0 {
			continue
		}
		vulnMod := f.Trace[0].Module
		if vulnMod == "" {
			continue
		}
		via := goDirectRoots(f.Trace, modulePath, directDeps)
		key := ecosystemGo + "\x00" + f.OSV + "\x00" + vulnMod + "\x00" + moduleDir
		if existing, ok := merged[key]; ok {
			existing.Via = unionStrings(existing.Via, via)
			continue
		}
		merged[key] = &Finding{
			ID:        f.OSV,
			Ecosystem: ecosystemGo,
			Package:   vulnMod,
			Module:    moduleDir,
			Via:       via,
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("govulncheck JSON: %w", err)
	}
	out := make([]Finding, 0, len(merged))
	for _, f := range merged {
		f.Via = sortedCopy(f.Via)
		out = append(out, *f)
	}
	return out, nil
}

func goDirectRoots(trace []govulnFrame, modulePath string, directDeps map[string]struct{}) []string {
	if len(trace) == 0 {
		return nil
	}
	if trace[0].Module == modulePath {
		return nil
	}
	seen := make(map[string]struct{})
	var via []string
	for _, fr := range trace {
		mod := fr.Module
		if mod == "" || mod == modulePath || mod == "stdlib" {
			continue
		}
		if _, ok := directDeps[mod]; !ok {
			continue
		}
		if _, dup := seen[mod]; dup {
			continue
		}
		seen[mod] = struct{}{}
		via = append(via, mod)
	}
	return via
}

func loadGoModule(goModPath string) (modulePath string, directDeps map[string]struct{}, err error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", nil, err
	}
	f, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return "", nil, fmt.Errorf("go.mod: %w", err)
	}
	if f.Module == nil || f.Module.Mod.Path == "" {
		return "", nil, fmt.Errorf("go.mod: missing module path")
	}
	directDeps = make(map[string]struct{})
	for _, r := range f.Require {
		if r.Indirect {
			continue
		}
		directDeps[r.Mod.Path] = struct{}{}
	}
	return f.Module.Mod.Path, directDeps, nil
}

func discoverGoModules(repoRoot string) ([]string, error) {
	var dirs []string
	for _, d := range []string{"backend", "front"} {
		if _, err := os.Stat(filepath.Join(repoRoot, d, "go.mod")); err == nil {
			dirs = append(dirs, d)
		}
	}
	toolsDir := filepath.Join(repoRoot, "tools")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rel := filepath.ToSlash(filepath.Join("tools", e.Name()))
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(rel), "go.mod")); err == nil {
			dirs = append(dirs, rel)
		}
	}
	return dirs, nil
}
