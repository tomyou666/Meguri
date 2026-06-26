# RunComfy CLI (Docker)

[RunComfy CLI](https://docs.runcomfy.com/cli/introduction) は Linux / macOS のみ対応。Windows では Docker Desktop 上の Linux コンテナ経由で実行する。

## 前提

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)（WSL2 バックエンド推奨）
- RunComfy アカウントと API トークン

## セットアップ

```powershell
cd tools/runcomfy
copy .env.example .env
```

`.env` にトークンを記入:

```
RUNCOMFY_TOKEN=your_token_here
```

トークンは [RunComfy](https://www.runcomfy.com) のダッシュボードから取得。コンテナ内の `runcomfy login`（ブラウザ device-code）は使わない。

イメージをビルド:

```powershell
docker compose build
```

## 使い方

`docker compose run --rm runcomfy` の後ろに、通常の `runcomfy` サブコマンドをそのまま渡す。生成物は `--output-dir /output` に保存するとホストの `tools/runcomfy/output/` に落ちる。

### バージョン確認

```powershell
docker compose run --rm runcomfy --version
```

### GPT Image 2 — text-to-image

PowerShell では JSON をシングルクォートで囲むとそのまま渡せる:

```powershell
docker compose run --rm runcomfy run openai/gpt-image-2/text-to-image `
  --input '{"prompt":"A minimal hero product still life","size":"1024_1024"}' `
  --output-dir /output
```

JSON 内にシングルクォートを含む場合は、外側をダブルクォートにして `"` をバッククォートでエスケープ:

```powershell
docker compose run --rm runcomfy run openai/gpt-image-2/text-to-image `
  --input "{`"prompt`":`"the word AQUA+ on a bottle`",`"size`":`"1024_1024`"}" `
  --output-dir /output
```

bash / WSL / Git Bash:

```bash
docker compose run --rm runcomfy run openai/gpt-image-2/text-to-image \
  --input '{"prompt":"A minimal hero product still life","size":"1024_1024"}' \
  --output-dir /output
```

### GPT Image 2 — edit

```powershell
docker compose run --rm runcomfy run openai/gpt-image-2/edit `
  --input '{"prompt":"Turn the background into a bright white studio","images":["https://example.com/photo.jpg"]}' `
  --output-dir /output
```

`images` は公開 HTTPS URL のみ（ローカルファイルは不可）。

### 任意のモデル / エンドポイント

```powershell
docker compose run --rm runcomfy run <model>/<endpoint> `
  --input '{"prompt":"..."}' `
  --output-dir /output
```

## トラブルシュート

| 症状 | 対処 |
|------|------|
| `Set RUNCOMFY_TOKEN in .env` | `.env` を作成しトークンを設定 |
| exit 77 | トークン未設定・無効・期限切れ — `.env` を確認しダッシュボードで再発行 |
| exit 65 | `--input` の JSON / スキーマ不正（例: 未対応の `size`） |
| `output/` にファイルがない | `--output-dir /output` を指定したか確認 |

公式: [CLI troubleshooting](https://docs.runcomfy.com/cli/troubleshooting)

## ビルドオプション

CLI バージョンをピン留め（デフォルト `v0.1.2`）:

```powershell
docker compose build --build-arg RUNCOMFY_VERSION=v0.1.1
```

バイナリは [GitHub Releases](https://github.com/runcomfy-com/runcomfy-cli/releases) から取得（`install.sh` の代替）。
