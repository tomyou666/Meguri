# vuln-gate

`npm audit` と `govulncheck` を実行し、リポジトリルートの [`vuln-allowlist.yaml`](../../vuln-allowlist.yaml) と突合して成否を判定するゲートツールです。

許容ルールの詳細やエントリの追加手順は [docs/vuln-allowlist.md](../../docs/vuln-allowlist.md) を参照してください。

## 使い方

リポジトリルートで:

```bash
# make 経由（推奨: 先に単体テストを実行してから実行）
make vuln-gate

# 直接実行（第 1 引数にリポジトリルートのパスを指定、省略時はカレントディレクトリ）
go run ./tools/vuln-gate .
```

## スキャン対象

- **npm**: `front/frontend` の `npm audit --json`（重大度 `high` / `critical` のみ）
- **Go**: `backend`、`front`、および `tools/*/go.mod` 配下の各 Go モジュール（`govulncheck -json ./...`）

環境変数 `NPM` / `GOVULNCHECK` で各スキャナーの実行バイナリを上書きできます。

```bash
NPM="npm" GOVULNCHECK="govulncheck" go run ./tools/vuln-gate .
```

## 判定ロジック

- 検出 1 件は **advisory 1 件**（`ecosystem` + `id` + `package` + `module`）単位
- `via` はその脆弱性に到達する **直接依存パッケージの集合**（完全一致が必要）
- 期限（`expires`）は **Asia/Tokyo** の暦日でその日を含む。翌日以降は失敗
- 検出に対応しない「古い・不要になった」許容エントリ（stale）が存在する場合も失敗

## テスト

```bash
go test ./tools/vuln-gate/...
```
