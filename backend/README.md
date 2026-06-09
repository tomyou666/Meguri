# scraperbot（バックエンド）

Web ページとリンク先 PDF を取得し、Markdown などの形式に変換する Go 製スクレイピング CLI です。コンパイル時プラグイン（マイクロカーネル構成）で処理パイプラインを差し替えできます。

詳細な設計は [`doc/`](doc/) 配下の設計書を参照してください。

## 必要環境

- Go 1.26 以上（[`go.mod`](go.mod) 準拠）
- 開発時（任意）: [golangci-lint](https://golangci-lint.run/)（`make lint` 用）
- `plugins.fetcher: chromium` を使う場合: Chromium または Microsoft Edge（Chromium ベース）。Dev Container では `chromium` パッケージが入っています

## ビルド

```bash
# このディレクトリで実行
make build

# または直接 go build
go build -o bin/scraperbot ./cmd/scraperbot
```

リポジトリルートから実行する場合:

```bash
make -C backend build
```

## クイックスタート

### 単一 URL を Markdown で取得（標準出力）

```bash
./bin/scraperbot --url https://example.com/ --stdout
```

### 単一 URL をファイルに保存

```bash
./bin/scraperbot --url https://example.com/ --output-dir ./out
```

`./out/` に `{連番}-{ホスト}-{パス}.md` 形式で保存されます（デフォルトのファイル名パターン）。

### 設定ファイルを使う

[`configs/config.example.yaml`](configs/config.example.yaml) をコピーして編集し、`-config` で指定します。

```bash
cp configs/config.example.yaml my-config.yaml
# my-config.yaml の targets などを編集

./bin/scraperbot --config my-config.yaml --stdout
```

### サイト内クロール

```bash
./bin/scraperbot \
  --url https://example.com/docs/ \
  --crawl \
  --max-depth 2 \
  --max-pages 50 \
  --output-dir ./out
```

クロール完了後、標準出力に `enqueued` / `succeeded` / `failed` / `skipped` の件数サマリが表示されます。

## 設定の読み込み順

設定は次の順でマージされ、最後に `Validate()` が実行されます。

1. **組み込みデフォルト**（[`internal/domain/model/config.go`](internal/domain/model/config.go) の `Default()`）
2. **YAML 設定ファイル**（`--config` で指定した場合）
3. **CLI フラグ**（同名項目はフラグが優先）

`targets`（対象 URL）は、YAML の `targets` または `--url` / 位置引数のいずれかで **1 件以上** 指定する必要があります。

## YAML 設定

### 全体構造


| セクション     | 説明                                    |
| --------- | ------------------------------------- |
| `targets` | 起点 URL のリスト（`http://` または `https://`） |
| `request` | HTTP ヘッダ・タイムアウト・リトライ                  |
| `content` | 出力フォーマット・本文抽出・セレクタ                    |
| `pdf`     | PDF リンクの追跡・解析モード                      |
| `crawl`   | クロール深度・件数・パス制限・並行数                    |
| `plugins` | フェッチャ・各段プラグイン名                         |
| `output`  | 出力ディレクトリ・ファイル名パターン                    |


### 例（抜粋）

```yaml
targets:
  - https://example.com/docs

request:
  headers:
    User-Agent: "scraperbot/0.1"
  timeout: 60s
  retry_count: 2
  retry_interval: 1s

content:
  formats: [markdown, links]
  only_main_content: true
  exclude_tags: [script, style, noscript]
  selector: ""          # 指定時は only_main_content より優先
  extract_links: true
  extract_metadata: true

pdf:
  enabled: true
  mode: auto            # fast / auto / ocr
  max_pages: 0          # 0 = 無制限
  output: text          # text / markdown / raw

crawl:
  enabled: true
  max_depth: 2
  max_pages: 100
  include_paths:        # 空 = 全許可
    - "^/docs/.*"
  exclude_paths:
    - ".*\\.zip$"
  allow_external_links: false
  allow_subdomains: false
  request_delay: 0s     # > 0 のとき並行数は 1 に強制
  max_concurrency: 4
  respect_robots_txt: true

plugins:
  fetcher: http
  fetcher_config:
    browser_path: ""
    user_agent: ""
    headless: true
    wait_visible_selector: ""
    wait_timeout: 5s
  preprocessors: [header]
  parsers: [html, pdf]
  transformer: markdown
  filters: [maincontent]
  link_extractor: default

output:
  dir: "./out"
  file_pattern: "{seq}-{host}-{path}.{ext}"
```

完全な例は [`configs/config.example.yaml`](configs/config.example.yaml) を参照してください。

## CLI フラグ一覧

`./bin/scraperbot -h` でもヘルプを確認できます。

### 一般


| フラグ        | 説明                               |
| ---------- | -------------------------------- |
| `--config` | YAML 設定ファイルのパス                   |
| `--url`    | 対象 URL（1 件）。位置引数でも指定可            |
| `--stdout` | 単一 URL モードで結果を標準出力へ（主に Markdown） |


### リクエスト（`request`）


| フラグ                       | YAML キー                  | デフォルト |
| ------------------------- | ------------------------ | ----- |
| `--header KEY=VAL`（繰り返し可） | `request.headers`        | 空     |
| `--timeout`               | `request.timeout`        | `60s` |
| `--retry`                 | `request.retry_count`    | `2`   |
| `--retry-interval`        | `request.retry_interval` | `1s`  |


### コンテンツ（`content`）


| フラグ                    | YAML キー                     | デフォルト                     |
| ---------------------- | --------------------------- | ------------------------- |
| `--format`（繰り返し可）      | `content.formats`           | `markdown`                |
| `--only-main`          | `content.only_main_content` | `true`                    |
| `--include-tag`（繰り返し可） | `content.include_tags`      | 空                         |
| `--exclude-tag`（繰り返し可） | `content.exclude_tags`      | `script, style, noscript` |
| `--selector`           | `content.selector`          | 空                         |
| `--extract-links`      | `content.extract_links`     | `true`                    |
| `--extract-metadata`   | `content.extract_metadata`  | `true`                    |


**出力フォーマット**（`--format` / `content.formats`）: `markdown`, `html`, `raw_html`, `json`, `links`, `metadata`

### PDF（`pdf`）


| フラグ               | YAML キー         | デフォルト    |
| ----------------- | --------------- | -------- |
| `--pdf`           | `pdf.enabled`   | `true`   |
| `--pdf-mode`      | `pdf.mode`      | `auto`   |
| `--pdf-max-pages` | `pdf.max_pages` | `0`（無制限） |
| `--pdf-output`    | `pdf.output`    | `text`   |


`pdf.enabled=false` のとき、PDF リンクはクロール対象から除外され、PDF URL を直接指定した場合はエラーになります。

### クロール（`crawl`）


| フラグ                     | YAML キー                      | デフォルト   |
| ----------------------- | ---------------------------- | ------- |
| `--crawl`               | `crawl.enabled`              | `false` |
| `--max-depth`           | `crawl.max_depth`            | `2`     |
| `--max-pages`           | `crawl.max_pages`            | `100`   |
| `--include-path`（繰り返し可） | `crawl.include_paths`        | 空（全許可）  |
| `--exclude-path`（繰り返し可） | `crawl.exclude_paths`        | 空       |
| `--allow-external`      | `crawl.allow_external_links` | `false` |
| `--allow-subdomains`    | `crawl.allow_subdomains`     | `false` |
| `--delay`               | `crawl.request_delay`        | `0s`    |
| `--concurrency`         | `crawl.max_concurrency`      | `4`     |
| `--fetch-http-inflight` | `crawl.fetch_limits.http_max_inflight` | `16` |
| `--fetch-chromium-inflight` | `crawl.fetch_limits.chromium_max_inflight` | `2`（静的上限・キャリブレーション失敗時のフォールバック） |
| `--fetch-auto-calibrate` | `crawl.fetch_limits.auto_calibrate` | `true`（ジョブ開始時に Chromium 上限を 1〜8 で再計算） |
| `--fetch-dynamic-chromium` | `crawl.fetch_limits.dynamic_chromium` | `true`（実行中にメモリ使用率で Chromium 上限を ±1） |
| `--respect-robots`      | `crawl.respect_robots_txt`   | `true`  |

`max_concurrency` はワーカー数、`fetch-*-inflight` は実際の取得同時実行数。Chromium ではワーカーが上限を超えても待機するだけで失敗しない。`auto_calibrate` / `dynamic_chromium` は **Chromium 上限のみ**に効く（HTTP は `http_max_inflight` 固定）。

### フェッチャ（`plugins.fetcher`）


| フラグ | YAML キー | デフォルト |
| --- | --- | --- |
| `--fetcher` | `plugins.fetcher` | `http` |
| `--fetcher-browser-path` | `plugins.fetcher_config.browser_path` | 空（自動検出） |
| `--fetcher-user-agent` | `plugins.fetcher_config.user_agent` | 空 |
| `--fetcher-headless` | `plugins.fetcher_config.headless` | `true` |

**Fetcher の選択肢**: `http`（標準 HTTP 取得）, `chromium`（chromedp によるヘッドレスブラウザ取得）

chromium 使用時のブラウザ探索順: `fetcher_config.browser_path` → 環境変数 `SCRAPERBOT_CHROMIUM_PATH` → Chromium 系 → Edge 系

User-Agent の優先順位（chromium 時）: `fetcher_config.user_agent` → `request.headers["User-Agent"]` → 既定の Chromium 系 UA

### プラグイン（`plugins`）


| フラグ                     | YAML キー                  | デフォルト         |
| ----------------------- | ------------------------ | ------------- |
| `--preprocessor`（繰り返し可） | `plugins.preprocessors`  | 空             |
| `--parser`（繰り返し可）       | `plugins.parsers`        | `html`, `pdf` |
| `--transformer`         | `plugins.transformer`    | `markdown`    |
| `--filter`（繰り返し可）       | `plugins.filters`        | `maincontent` |
| `--link-extractor`      | `plugins.link_extractor` | `default`     |


### 出力（`output`）


| フラグ                | YAML キー               | デフォルト                       |
| ------------------ | --------------------- | --------------------------- |
| `--output-dir`     | `output.dir`          | `./out`                     |
| `--output-pattern` | `output.file_pattern` | `{seq}-{host}-{path}.{ext}` |


**ファイル名プレースホルダ**: `{seq}`（5 桁連番）, `{host}`, `{path}`（サニタイズ済み）, `{ext}`（フォーマットに応じた拡張子）

## 実行例

### CSS セレクタで本文を絞る

```bash
./bin/scraperbot \
  --url https://example.com/page \
  --selector "article.main" \
  --filter selector \
  --stdout
```

`selector` フィルタプラグインを有効にし、`--selector` で指定した範囲だけを残します。

### クロール＋パス制限

```bash
./bin/scraperbot \
  --config my-config.yaml \
  --crawl \
  --include-path '^/docs/.*' \
  --exclude-path '.*\.zip$' \
  --max-depth 1 \
  --output-dir ./out
```

### ヘッダを付与

```bash
./bin/scraperbot \
  --url https://example.com/ \
  --header 'User-Agent=scraperbot/0.1' \
  --preprocessor header \
  --stdout
```

`header` プリプロセッサは設定の `request.headers` をリクエストへ転写します。

### JavaScript ページを chromedp で取得

```bash
./bin/scraperbot \
  --url https://example.com/ \
  --fetcher chromium \
  --fetcher-headless \
  --stdout
```

ブラウザパスを明示する例:

```bash
export SCRAPERBOT_CHROMIUM_PATH=/usr/bin/chromium
./bin/scraperbot --url https://example.com/ --fetcher chromium --stdout
```

## 組み込みプラグイン


| 名前            | 種別                 | 説明                          |
| ------------- | ------------------ | --------------------------- |
| `http`        | Fetcher            | net/http による URL 取得（既定）    |
| `chromium`    | Fetcher            | chromedp によるヘッドレスブラウザ取得   |
| `header`      | PreProcessor (P2)  | 共通 HTTP ヘッダの付与              |
| `html`        | Parser (P5)        | HTML を goquery で解析          |
| `pdf`         | Parser (P5)        | PDF レスポンスの処理（MVP: 簡易テキスト抽出） |
| `markdown`    | Transformer (P6)   | HTML → Markdown 変換          |
| `maincontent` | Filter (P7)        | ヘッダ・フッタ・ナビ等の除去              |
| `selector`    | Filter (P7)        | CSS セレクタによる絞り込み             |
| `default`     | LinkExtractor (P8) | `<a href>` の抽出と URL 解決      |


プラグインの追加・差し替えは [`cmd/scraperbot/main.go`](cmd/scraperbot/main.go) の副作用 import を編集して再ビルドします。

## 開発

```bash
# フォーマット + vet + golangci-lint + テスト（race 付き）
make check

# 個別ターゲット
make fmt      # go fmt
make vet      # go vet
make lint     # golangci-lint run
make test     # go test -race
make tidy     # go mod tidy
make wire     # internal/app/wire_gen.go を再生成
```

リポジトリルートから実行する場合は `make -C backend <target>` を使います。

`internal/app/providers.go` または `wire.go` を変更した場合は **`make wire` を実行** して `wire_gen.go` を再生成してください。通常の clone / build では `wire_gen.go` がコミット済みのため **`make wire` は不要** です。

テストは `httptest` でテスト用 Web サーバーを起動し、[`testdata/html/`](testdata/html/) の HTML を返して検証しています。

## ディレクトリ構成

```text
cmd/scraperbot/          # CLI エントリ（プラグインの副作用 import）
internal/
  app/                   # Wire composition root（依存グラフ組み立て）
  domain/                # エンティティ・プラグイン抽象
  core/                  # カーネル・パイプライン・クローラ
  usecase/               # シナリオ（単一 URL / クロール）
  infrastructure/        # HTTP・chromedp・設定読込・出力・robots.txt
  presentation/cli/      # CLI
plugins/                 # 具体プラグイン実装
configs/                 # 設定ファイル例
testdata/html/           # 統合テスト用 HTML
doc/                     # 設計書
```
