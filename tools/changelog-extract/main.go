// Command changelog-extract は CHANGELOG.md から指定版の節を標準出力へ書き出す。
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	changelogPath := flag.String("changelog", "", "path to CHANGELOG.md (default: <repo>/CHANGELOG.md)")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: go run ./tools/changelog-extract <X.Y.Z> [-changelog path]")
		fmt.Fprintln(os.Stderr, "Example: go run ./tools/changelog-extract 1.0.0")
		os.Exit(1)
	}

	version, err := parseVersionArg(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	path := *changelogPath
	if path == "" {
		path, err = defaultChangelogPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "changelog path: %v\n", err)
			os.Exit(1)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
	}

	section, err := extractSection(string(data), version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err := os.Stdout.WriteString(section); err != nil {
		fmt.Fprintf(os.Stderr, "write stdout: %v\n", err)
		os.Exit(1)
	}
}
