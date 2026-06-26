// Command meguri-cli は CLI エントリーポイント。
// 副作用 import によりコンパイル時にプラグインセットを決める。
package main

import (
	"os"

	"meguri/internal/presentation/cli"

	// プラグインのinit処理の実行
	_ "meguri/plugins/fetcher-chromium"
	_ "meguri/plugins/fetcher-http"
	_ "meguri/plugins/filter-maincontent"
	_ "meguri/plugins/filter-selector"
	_ "meguri/plugins/linkextractor-default"
	_ "meguri/plugins/parser-html"
	_ "meguri/plugins/parser-pdf"
	_ "meguri/plugins/preprocessor-header"
	_ "meguri/plugins/transformer-html"
	_ "meguri/plugins/transformer-markdown"
	_ "meguri/plugins/transformer-raw-html"
)

// main は CLI を起動し、終了コードを os.Exit に渡す。
func main() {
	os.Exit(cli.Run())
}
