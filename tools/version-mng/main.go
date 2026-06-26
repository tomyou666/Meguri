// Command version-mng はデスクトップアプリの版を一元管理し、ビルド資産へ伝播する。
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run ./tools/version-mng <X.Y.Z>")
		fmt.Fprintln(os.Stderr, "Example: go run ./tools/version-mng 1.0.0")
		os.Exit(1)
	}

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "repo root: %v\n", err)
		os.Exit(1)
	}

	changed, err := setVersion(root, os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	for _, path := range changed {
		fmt.Println(path)
	}
}
