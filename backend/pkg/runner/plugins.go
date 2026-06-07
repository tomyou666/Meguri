package runner

// プラグインをレジストリへ登録する（main またはテストから import する）。
import (
	_ "scraperbot/plugins/fetcher-chromium"
	_ "scraperbot/plugins/fetcher-http"
	_ "scraperbot/plugins/filter-maincontent"
	_ "scraperbot/plugins/filter-selector"
	_ "scraperbot/plugins/linkextractor-default"
	_ "scraperbot/plugins/parser-html"
	_ "scraperbot/plugins/parser-pdf"
	_ "scraperbot/plugins/preprocessor-header"
	_ "scraperbot/plugins/transformer-markdown"
)
