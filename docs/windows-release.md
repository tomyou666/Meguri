# Meguri — Windows リリース手順

Wails v3 デスクトップアプリの Windows 版を GitHub Releases に公開する手順。

**配布の考え方（要点のみ）**

- 初回インストール: NSIS（`meguri-amd64-installer.exe`）
- 既存ユーザー更新: Wails updater が `meguri-windows-amd64.zip` を取得して exe 差し替え
- インストール先: user スコープ（`%LOCALAPPDATA%`、管理者権限不要）
- コード署名: 当面なし（SmartScreen「発行元不明」は許容）

---

## 前提

| 項目 | 内容 |
|------|------|
| CI ワークフロー | [`.github/workflows/release-windows.yml`](../.github/workflows/release-windows.yml) |
| トリガー | `v*` タグの push |
| リポジトリ（updater） | `tomyou666/scraper-bot`（[`front/main.go`](../front/main.go) の `githubRepository`） |
| ローカル手動ビルド | NSIS（`makensis`）が PATH にあること |

### GitHub Secrets（初回セットアップ）

| Secret | 必須 | 用途 |
|--------|------|------|
| `GITHUB_TOKEN` | 自動 | Release 公開 |
| `UPDATER_PRIVATE_KEY` | 任意 | `SHA256SUMS` の Ed25519 署名（[`sign-release`](../tools/sign-release/main.go) が解釈する形式のみ） |

公開鍵は [`front/updater-key.pub`](../front/updater-key.pub) にコミット。秘密鍵は **リポジトリに置かず** GitHub Secrets へ登録する。**本番前に鍵ペアを再生成すること。**

`UPDATER_PRIVATE_KEY` に入れる値は次のいずれか（[`parseEd25519Private`](../tools/sign-release/main.go) 準拠）:

| 形式 | 内容 |
|------|------|
| seed の base64（推奨） | Ed25519 の 32 バイト seed を base64 した **1 行**（パスワードのハッシュ等ではない） |
| PEM | seed 32 バイト、または Ed25519 秘密鍵 64 バイトを PEM で包んだもの |


#### 鍵ペアの生成（初回・ローテーション）

リポジトリルートで実行:

```bash
go run ./tools/updater-keygen
```

1. `front/updater-key.pub` が上書きされる → **コミットする**
2. 標準出力の 1 行（seed の base64）をコピー → GitHub **Settings → Secrets → Actions** で `UPDATER_PRIVATE_KEY` に登録
3. 秘密鍵は Secret 登録後にローカルから消してよい（コミットしない）

Secret 未設定でも Release は動くが、`SHA256SUMS.sig` は生成されない。

---

## Release に載るファイル

| ファイル | 用途 |
|----------|------|
| `meguri-amd64-installer.exe` | 新規ユーザー向け NSIS インストーラー |
| `meguri-windows-amd64.zip` | updater 用（`meguri.exe` のみ） |
| `SHA256SUMS` | 各資産の SHA-256 |
| `SHA256SUMS.sig` | `UPDATER_PRIVATE_KEY` 設定時のみ（任意） |

セキュリティ修正は **必ず新しいバージョン番号**で出す（アプリに「このバージョンをスキップ」があるため）。

---

## バージョンの揃え方

正: [`tools/version-mng/version.json`](../tools/version-mng/version.json)

形式は **`X.Y.Z` のみ**（`1.0.0` 等）。`v` 接頭辞や `-beta` 等の suffix は使わない。

```bash
go run ./tools/version-mng 1.0.0
```

`version-mng` が更新するファイル:

| ファイル | 値 |
|----------|-----|
| `front/build/config.yml` → `info.version` | `X.Y.Z` |
| `front/build/windows/info.json` → `ProductVersion` | `X.Y.Z` |
| `front/build/windows/info.json` → `file_version` | `X.Y.Z.0` |
| `front/frontend/src/i18n/messages.ts` → `version` | `X.Y.Z` |
| `front/build/windows/version-nsis.nsh` | NSIS 用 `X.Y.Z` |

**別系統の版**

| 用途 | 更新方法 |
|------|----------|
| updater 実行時（`main.currentVersion`） | CI が tag から先頭 `v` を除去して `-ldflags "-X main.currentVersion=..."` 注入 |

git tag は `v` + appVersion（例: `v1.0.0`）。tag と `version-mng` の版は一致させる。

---

## リリース手順（通常・CI）

1. 版を決める（`X.Y.Z`）
2. `go run ./tools/version-mng <X.Y.Z>`
3. 変更をコミット（`version-nsis.nsh` 含む）
4. タグを付けて push:

```bash
git tag v1.0.0
git push origin v1.0.0
```

5. [Actions](https://github.com/tomyou666/scraper-bot/actions) で `Release Windows` を確認
6. GitHub Release の資産・リリースノートを確認

CI の処理概要: tag 抽出 → `version-mng` → `wails3 task windows:package INSTALL_SCOPE=user APP_VERSION=<版>` → zip 作成 → `go run ./tools/sign-release ./front/bin` → Release 公開

---

## ローカル手動ビルド（検証用）

```powershell
cd front
wails3 task windows:package INSTALL_SCOPE=user APP_VERSION=1.0.0
```

出力:

- `front/bin/meguri.exe`
- `front/bin/meguri-amd64-installer.exe`（NSIS 成功時）

updater 用 zip とチェックサム:

```powershell
Compress-Archive -Path bin/meguri.exe -DestinationPath bin/meguri-windows-amd64.zip -Force
cd ..
# `SHA256SUMS.sig` も生成する場合は、.env に UPDATER_PRIVATE_KEY=<seed の base64 1 行> を記入しておく
copy .env.example .env
# .env に UPDATER_PRIVATE_KEY=<seed の base64 1 行> を記入
go run ./tools/sign-release ./front/bin
```

- 未設定のまま実行 → `SHA256SUMS` のみ生成（CI の Secret 未設定時と同じ）

---

## リリース後の確認

- [ ] Release に installer / zip / `SHA256SUMS` がある
- [ ] インストール後、アプリ表示版が tag と一致（設定メニュー上の版）
- [ ] 既存インストールから「更新を確認…」で新 Release を検知できる（テスト用に旧版を入れて確認）
- [ ] セキュリティ修正なら版を上げた（スキップ済み版の上書き配信になっていない）

---

## ユーザー向けメモ（README 等に転記可）

**初回インストール:** Release から `meguri-amd64-installer.exe` を取得。SmartScreen が出たら「詳細情報」→「実行」。

**更新:** 起動時に自動確認。手動は設定メニュー「更新を確認…」。

**アンインストール:** 「アプリの追加と削除」から Meguri を削除。

---

## 参考

- [Wails v3 — Self-Updating Wails App](https://v3.wails.io/tutorials/04-self-update-a-wails-app/)
- [`tools/version-mng/README.md`](../tools/version-mng/README.md)（手動確認チェックリスト）
