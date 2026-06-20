# scraperbot

Web ページとリンク先 PDF を取得し、Markdown などの形式に変換するスクレイピングツールです。

## 構成

| ディレクトリ | 説明 |
| --- | --- |
| [`backend/`](backend/) | Go 製 CLI（ビルド・設定・プラグイン・開発手順）→ **[backend/README.md](backend/README.md)** |
| [`front/`](front/) | Wails3 デスクトップ UI → [front/README.md](front/README.md) |
| [`docs/`](docs/) | API などプロジェクト横断のドキュメント |

バックエンドのクイックスタート（ビルド・CLI・YAML 設定・フラグ一覧・プラグイン）は [backend/README.md](backend/README.md) を参照してください。詳細な設計は [backend/doc/](backend/doc/) 配下の設計書にあります。

## 開発環境（Go）

このリポジトリの Go バージョンは [`.prototools`](.prototools) で管理しています（現在 `1.26.3`）。[proto](https://moonrepo.dev/proto) を使ってインストールしてください。

### `wails3 dev` が `compile: version "go1.26.x" does not match go tool version "go1.26.y"` で失敗する

Windows に別バージョンの Go を手動インストールしていると、`go` コマンドと `compile` などのビルドツールのバージョンが食い違うことがあります。典型的には次の状態です。

- システムの `GOROOT` が古い Go（例: `C:\go1.26.2`）を指している
- `go.mod` / proto は新しい Go（例: `1.26.3`）を使おうとする

**対処:**

1. Windows の環境変数から **`GOROOT` を削除**する（Go は通常、自動で設定されます）
2. PATH で proto の shims（`%USERPROFILE%\.proto\shims`）を、手動インストールした Go の `bin` より前に置く。使わないなら古い Go を PATH から外すかアンインストールする
3. 環境変数を直したあと、**新しいターミナル**を開いて `go clean -cache` を実行し、再度ビルドする

**確認（すべて同じバージョンであること）:**

```powershell
go version
go env GOROOT
go version -m (go tool -n compile)
```

## ライセンス

[LICENSE](LICENSE) を参照してください。
