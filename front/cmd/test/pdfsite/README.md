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

