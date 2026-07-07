package model

// OutputFormat は結果ファイルの出力形式を表す。
type OutputFormat string

const (
	// FormatMarkdown は Markdown 本文出力。
	FormatMarkdown OutputFormat = "markdown"
	// FormatHTML は整形済み HTML 出力。
	FormatHTML OutputFormat = "html"
	// FormatRawHTML は取得した生 HTML 出力。
	FormatRawHTML OutputFormat = "raw_html"
	// FormatJSON は URL・メタデータ等を JSON で出力。
	FormatJSON OutputFormat = "json"
	// FormatLinks は抽出リンクを 1 行 1 URL のテキストで出力。
	FormatLinks OutputFormat = "links"
	// FormatMetadata はメタデータを key: value 形式で出力。
	FormatMetadata OutputFormat = "metadata"
)

// Valid は定義済みの OutputFormat かどうかを返す。
func (f OutputFormat) Valid() bool {
	switch f {
	case FormatMarkdown, FormatHTML, FormatRawHTML, FormatJSON, FormatLinks, FormatMetadata:
		return true
	}
	return false
}

// PDFParseMode は PDF 解析の実行モードを表す。
//
// "fast": 埋め込みテキストのみを優先し、高速に抽出する。
// "auto": テキスト抽出を試み、不十分な場合は OCR にフォールバックする。
// "ocr": 画像ベースの OCR 解析を優先する。
type PDFParseMode string

const (
	// PDFModeFast は高速テキスト抽出モード。
	PDFModeFast PDFParseMode = "fast"
	// PDFModeAuto は自動フォールバックモード。
	PDFModeAuto PDFParseMode = "auto"
	// PDFModeOCR は OCR 優先モード。
	PDFModeOCR PDFParseMode = "ocr"
)

// Valid は定義済みの PDFParseMode かどうかを返す。
func (m PDFParseMode) Valid() bool {
	switch m {
	case PDFModeFast, PDFModeAuto, PDFModeOCR:
		return true
	}
	return false
}

// PDFOutput は PDF からの出力表現形式を表す。
//
// "text": プレーンテキストとして出力する。
// "markdown": Markdown として整形して出力する。
// "raw": バイナリまたは生データをそのまま扱う。
type PDFOutput string

const (
	// PDFOutputText はプレーンテキスト出力。
	PDFOutputText PDFOutput = "text"
	// PDFOutputMarkdown は Markdown 出力。
	PDFOutputMarkdown PDFOutput = "markdown"
	// PDFOutputRaw は生データ出力。
	PDFOutputRaw PDFOutput = "raw"
)

// Valid は定義済みの PDFOutput かどうかを返す。
func (o PDFOutput) Valid() bool {
	switch o {
	case PDFOutputText, PDFOutputMarkdown, PDFOutputRaw:
		return true
	}
	return false
}

// FetcherKind は URL フェッチ実装の種別を表す。
//
// "http": net/http による静的取得。
// "chromium": chromedp によるヘッドレス Chromium 取得。
type FetcherKind string

const (
	// FetcherHTTP は標準 HTTP クライアント取得。
	FetcherHTTP FetcherKind = "http"
	// FetcherChromium は chromedp によるブラウザ取得。
	FetcherChromium FetcherKind = "chromium"
)

// Valid は定義済みの FetcherKind かどうかを返す。
func (f FetcherKind) Valid() bool {
	switch f {
	case FetcherHTTP, FetcherChromium:
		return true
	}
	return false
}

// WaitUntil は chromium 取得前のページ読み込み待機モードを表す。
//
// "none": Navigate 直後に HTML を取得する。
// "load": DOM load 相当まで待つ（既定）。
// "network_idle": アクティブなネットワーク接続が一定時間ゼロになるまで待つ。
// "selector": wait_visible_selector で指定した要素が可視になるまで待つ。
type WaitUntil string

const (
	// WaitUntilNone は追加の読み込み待機を行わない。
	WaitUntilNone WaitUntil = "none"
	// WaitUntilLoad は body の load 相当まで待つ。
	WaitUntilLoad WaitUntil = "load"
	// WaitUntilNetworkIdle はネットワーク静止まで待つ。
	WaitUntilNetworkIdle WaitUntil = "network_idle"
	// WaitUntilSelector は CSS セレクタの要素が可視になるまで待つ。
	WaitUntilSelector WaitUntil = "selector"
)

// Valid は定義済みの WaitUntil かどうかを返す。
func (w WaitUntil) Valid() bool {
	switch w {
	case WaitUntilNone, WaitUntilLoad, WaitUntilNetworkIdle, WaitUntilSelector:
		return true
	}
	return false
}
