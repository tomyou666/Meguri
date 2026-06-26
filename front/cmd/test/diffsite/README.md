# diffsite — 差分 UI 手動確認用 HTTP サーバー

**テスト専用。本番・CI では使わない。**

Wails App 本体で Play → 差分 UI を確認するための静的サイト。各フィクスチャは 1 種類の差分だけが出るよう設計している。

| フィクスチャ | 変えるもの | 揃えるもの |
|-------------|-----------|-----------|
| `content/a` vs `content/b` | 見出し・段落（本文） | 出リンク一覧 |
| `links/a` vs `links/b` | `<a href>` URL 集合 | 本文 |
| `fetch/a` vs `fetch/b` | `/error.html` の HTTP ステータス（200 vs 500） | `/` の本文・links |

## 起動

```bash
cd front/cmd/test/diffsite
go run . -variant=content-a
```

`-variant`: `content-a` | `content-b` | `links-a` | `links-b` | `fetch-a` | `fetch-b`（既定: `content-a`）  
`-addr`: 既定 `:18765`

## 手動確認フロー

各シナリオ共通: **`*-a` で 1 回目 Play（baseline）→ 停止 → `*-b` で 2 回目 Play（差分）**

### A. content 差分

1. `go run . -variant=content-a`
2. WS ルート URL = `http://localhost:18765/` → **Play** → baseline 設定、差分なし
3. 停止 → `go run . -variant=content-b` → **Play** → **content** 差分のみ

### B. links 差分

1. `go run . -variant=links-a` → **Play**
2. `go run . -variant=links-b` → **Play** → **links** 差分のみ

### C. fetch 差分（任意）

1. `go run . -variant=fetch-a`。ノード: `/` と `/error.html` → **Play**
2. `go run . -variant=fetch-b` → **Play** → `/error.html` で **fetch** 差分

全シナリオ後: ControlBar「変更を確認済みにする」でバッジ消去。
