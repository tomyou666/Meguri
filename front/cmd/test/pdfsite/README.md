# pdfsite — PDF クロール手動確認用 HTTP サーバー

**テスト専用。本番・CI では使わない。**

Wails App 本体で Play → PDF リンクのクロール・パースを確認するための静的サイト。

| パス | 内容 |
|------|------|
| `/` | PDF リンク付き HTML |
| `/sample-pdf.pdf` | `front/testdata/pdf/sample-pdf.pdf` |

## 起動

```bash
cd front/cmd/test/pdfsite
go run .
```

`-addr`: 既定 `:18766`

## 手動確認フロー

1. `go run .`
2. WS ルート URL = `http://localhost:18766/` → **Play**
3. ノードツリーに `/sample-pdf.pdf` が現れ、PDF パース結果が取れることを確認

### fetcher 別確認

| fetcher | 期待 |
|---------|------|
| `http` | PDF ノード成功。Markdown に論文タイトル/本文断片（ledongthuc/pdf 抽出） |
| `chromium` | 同上。PDF URL は HTTP フォールバックのため Chrome PDF Viewer の空 HTML にならない |

確認ポイント:

- ノード結果 metadata に `parse_strategy=ledongthuc` / `parse_mode=fast`
- Markdown に `%PDF-1.7 ... /FlateDecode ... stream` のようなバイナリ断片ゴミが **出ない**
- Chromium 時も `chrome-extension://.../pdf_embedder.css` の HTML が **出ない**
