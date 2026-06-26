# version-mng

Meguri デスクトップアプリのバージョンを一元管理し、ビルド資産へ伝播する。

## 使い方

リポジトリルートで:

```bash
go run ./tools/version-mng 1.0.0
```

- 形式: `X.Y.Z` のみ（例: `1.0.0`）。`v` 接頭辞・`-beta` 等の suffix は不可
- 正: [`version.json`](version.json) の `appVersion`
- `file_version` / NSIS 用数値版: 自動で `X.Y.Z.0`

## 手動確認チェックリスト

- [ ] `go run ./tools/version-mng 9.9.9` 後、表示版 `9.9.9`、`file_version` `9.9.9.0`
- [ ] `go run ./tools/version-mng 1.0.0-beta` がエラー終了する
- [ ] `wails3 task windows:package INSTALL_SCOPE=user` が成功する
- [ ] `config.yml` 先頭 `version: '3'` が変わっていない
- [ ] 不正引数（`abc`, `v1.0.0`）でエラー終了する
- [ ] 確認後、実際の版に戻す
