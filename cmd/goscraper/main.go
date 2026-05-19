// Command scraperbot は CLI エントリーポイント。
// 副作用 import によりコンパイル時にプラグインセットを決める。
package main

import (
	"os"

	"scraperbot/internal/presentation/cli"

	// プラグインのinit処理の実行
	_ "scraperbot/plugins/filter-maincontent"
	_ "scraperbot/plugins/filter-selector"
	_ "scraperbot/plugins/linkextractor-default"
	_ "scraperbot/plugins/parser-html"
	_ "scraperbot/plugins/parser-pdf"
	_ "scraperbot/plugins/preprocessor-header"
	_ "scraperbot/plugins/transformer-markdown"
)

func main() {
	os.Exit(cli.Run())
}
