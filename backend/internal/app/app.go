// Package app は Wire による composition root を提供する。
package app

import (
	"meguri/internal/core"
	"meguri/internal/infrastructure/storage"
	"meguri/internal/usecase"
)

// Application は実行時に必要な依存を束ねる。
type Application struct {
	// Kernel は初期化済みプラグインを束ねるカーネル。
	Kernel *core.Kernel
	// Scrape は単一 URL スクレイプ用ユースケース。
	Scrape *usecase.Scrape
	// Crawl はクロール用ユースケース。
	Crawl *usecase.Crawl
	// FileWriter は結果のファイル出力先。
	FileWriter *storage.FileWriter
}
