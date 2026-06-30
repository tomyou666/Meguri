# url-fetch

固定 URL を HTTP / Chromium の複数取得方法で試行し、各試行のメタ情報を標準出力へ書き出す診断用 CLI。

## 前提

- Go 1.22 以上
- Chromium バリアント利用時: Chrome / Edge / Chromium のいずれか、または環境変数 `MEGURI_CHROMIUM_PATH` で実行ファイルを指定

## 使い方

リポジトリルートで:

```bash
go run ./tools/url-fetch
go run ./tools/url-fetch https://example.com
```

- 引数なし: コード内の既定 URL を使う
- 引数1件: その URL を `http` / `https` で検証して使う

HTTP 6 種と Chromium 4 種を並列実行する（HTTP 群と Chromium 群も同時に走る）。stdout は HTTP → Chromium の定義順。

## 設定

[`main.go`](main.go) の `cfg` に集約:

| フィールド | 既定値 | 意味 |
|-----------|--------|------|
| `DefaultTargetURL` | （コード内） | 引数未指定時の URL |
| `HTTPVariantTimeout` | `5s` | HTTP 1 試行あたり |
| `ChromiumOverallTimeout` | `60s` | Chromium 群全体 |
| `ChromeUserAgent` | Chrome/120 | backend 既定 UA と同じ |
| `MaxWaitAfterNavigate` | `2s` | `headless_wait` の待機上限 |

HTTP / Chromium バリアント定義も `main.go` の `httpVariantsFor` / `chromiumVariants`。

## 並列実行

- HTTP 6 バリアント: 互いに並列（各 5 秒タイムアウト）
- Chromium 4 バリアント: 互いに並列（バリアントごとにブラウザを 1 つ起動）
- HTTP 群と Chromium 群: 同時並列

Chromium 並列時は最大 4 プロセス分のメモリを使う。診断用途向け。

## 時間予算

- HTTP: バリアントごとに最大 5 秒（並列のため壁時計はおおよそ 5 秒）
- Chromium: 全体で最大 60 秒（並列のため壁時計は最遅バリアントに依存）
- リトライ: なし
- `headless_wait` の待機は最大 2 秒。Chromium 用 context の残り時間が短い場合はそちらに合わせて短縮される

## HTTP バリアント

| variant | 内容 |
|---------|------|
| `default` | カスタムヘッダなし（`net/http` 既定 User-Agent） |
| `chrome_ua` | Chrome 風 User-Agent |
| `chrome_ua_lang` | 上記 + `Accept-Language: ja` + `Accept: text/html` |
| `chrome_ua_referer` | 上記 + `Referer`（対象 URL のルート） |
| `utls_chrome_ua` | utls `HelloChrome_Auto` + HTTP/2 + Chrome User-Agent |
| `utls_chrome_ua_http1` | 上記 + ALPN で HTTP/1.1 強制（TLS 指紋検証用。本番 `fetcher: http` では未使用） |

## Chromium バリアント

| variant | 内容 |
|---------|------|
| `headless_default` | headless、backend 既定 UA（allocator で Chrome/120） |
| `headless_chrome_ua` | headless、上記に加え emulation で UA 再指定 |
| `headless_wait` | headless、Chrome User-Agent、navigate 後に待機してから HTML 取得 |
| `chrome_ua` | headless=false、backend 既定 UA（ウィンドウ表示あり） |

Chromium 成功時の `status=200` は chromedp が HTTP ステータスを返せないための慣例表記。

## 出力形式

1 試行 1 行。フィールド: `method`, `variant`, `status`, `bytes`, `duration`。失敗時は `error` を追加。

```
method=http variant=default status=200 bytes=123456 duration=0.432s
method=chromium variant=headless_default status=200 bytes=125000 duration=1.102s
method=chromium variant=headless_wait status=0 bytes=0 duration=0.000s error="context deadline exceeded"
```

ブラウザ未検出時は Chromium 4 行とも `error` 付きで出力し、終了コードは `0`（HTTP 結果はそのまま表示）。

## 開発

```bash
make -C tools check
```
