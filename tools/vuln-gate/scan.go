package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type scannerConfig struct {
	NPM         string
	Govulncheck string
	RepoRoot    string
}

func defaultScannerConfig(repoRoot string) scannerConfig {
	npm := envOr("NPM", "npm")
	govuln := envOr("GOVULNCHECK", "govulncheck")
	return scannerConfig{NPM: npm, Govulncheck: govuln, RepoRoot: repoRoot}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func collectFindings(cfg scannerConfig) ([]Finding, error) {
	var all []Finding

	npmFindings, err := scanNPM(cfg)
	if err != nil {
		return nil, err
	}
	all = append(all, npmFindings...)

	goDirs, err := discoverGoModules(cfg.RepoRoot)
	if err != nil {
		return nil, fmt.Errorf("discover go modules: %w", err)
	}
	for _, dir := range goDirs {
		findings, err := scanGoModule(cfg, dir)
		if err != nil {
			return nil, err
		}
		all = append(all, findings...)
	}
	return all, nil
}

func scanNPM(cfg scannerConfig) ([]Finding, error) {
	modulePath := "front/frontend"
	pkgDir := filepath.Join(cfg.RepoRoot, filepath.FromSlash(modulePath))
	direct, err := loadNPMDirectDeps(filepath.Join(pkgDir, "package.json"))
	if err != nil {
		return nil, fmt.Errorf("npm direct deps: %w", err)
	}
	cmd := exec.Command(cfg.NPM, "audit", "--json", "--prefix", pkgDir)
	cmd.Dir = cfg.RepoRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("npm audit: %w\n%s", err, stderr.String())
		}
		// Exit code 1 with JSON is the normal "vulnerabilities found" case.
		exitErr := err.(*exec.ExitError)
		if exitErr.ExitCode() != 1 {
			return nil, fmt.Errorf("npm audit exit %d: %w\n%s", exitErr.ExitCode(), err, stderr.String())
		}
	}
	if stdout.Len() == 0 {
		return nil, fmt.Errorf("npm audit: empty stdout\n%s", stderr.String())
	}
	findings, err := parseNPMFindings(stdout.Bytes(), modulePath, direct)
	if err != nil {
		return nil, fmt.Errorf("npm audit parse: %w\nstderr: %s", err, stderr.String())
	}
	return findings, nil
}

func scanGoModule(cfg scannerConfig, moduleDir string) ([]Finding, error) {
	abs := filepath.Join(cfg.RepoRoot, filepath.FromSlash(moduleDir))
	modPath, direct, err := loadGoModule(filepath.Join(abs, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", moduleDir, err)
	}
	cmd := exec.Command(cfg.Govulncheck, "-json", "./...")
	cmd.Dir = abs
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			return nil, fmt.Errorf("govulncheck %s: %w\n%s", moduleDir, err, stderr.String())
		}
		// 3 = vulnerabilities found (success for our purposes).
		if code := exitErr.ExitCode(); code != 3 {
			return nil, fmt.Errorf("govulncheck %s exit %d: %w\n%s", moduleDir, code, err, stderr.String())
		}
	}
	findings, err := parseGovulncheckFindings(stdout.Bytes(), moduleDir, modPath, direct)
	if err != nil {
		return nil, fmt.Errorf("govulncheck %s parse: %w", moduleDir, err)
	}
	return findings, nil
}
