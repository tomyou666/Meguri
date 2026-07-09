package plugin

import (
	"context"

	"meguri/internal/domain/model"
)

// Kind はプラグインのパイプライン上の役割を表す。
type Kind string

const (
	// KindPreProcessor は P2 リクエスト前処理。
	KindPreProcessor Kind = "preprocessor"
	// KindFetcher は P3 URL 取得（ページ・robots.txt 等）。
	KindFetcher Kind = "fetcher"
	// KindParser は P5 レスポンス解析。
	KindParser Kind = "parser"
	// KindTransformer は P6 結果変換。
	KindTransformer Kind = "transformer"
	// KindFilter は P7 コンテンツ絞り込み。
	KindFilter Kind = "filter"
	// KindLinkExtractor は P8 リンク抽出。
	KindLinkExtractor Kind = "link_extractor"
)

// Metadata はプラグインの識別情報を保持する。
type Metadata struct {
	// Name はレジストリ登録名。
	Name string
	// Version はプラグインのセマンティックバージョン。
	Version string
	// Kind はパイプライン上の種別。
	Kind Kind
	// Description は人間向けの短い説明。
	Description string
}

// Plugin は全プラグイン種別が実装するライフサイクルインタフェース。
type Plugin interface {
	Metadata() Metadata
	Init(ctx context.Context, host Host) error
	Close(ctx context.Context) error
}

// Host はプラグインに渡される最小の依存集合。
type Host interface {
	Config(key string) (string, bool)
	RequestConfig() model.RequestConfig
	FetcherConfig() model.FetcherConfig
	StealthConfig() model.StealthConfig
	FetcherKind() model.FetcherKind
}
