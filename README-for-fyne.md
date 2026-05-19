# urls-scraper

指定したwebサイトをクロールして結果を保存する Fyne デスクトップアプリです。

## できること
- クロールするサイトを指定
- クロールする深さや幅を指定
- 保存方法を選択
  - URL ごとに分割
  - 1ファイルに統合（区切り文字を指定可能）
- URL 重複は自動で除外（重複件数をログ表示）

## 必要なもの

- Go 1.26 以上（`go.mod` 準拠）
- ソースからビルドする場合のみ:
  - `CGO_ENABLED=1`
  - C コンパイラ（gcc など）

> 配布済みバイナリを実行するだけなら `CGO_ENABLED` は不要です。

## ローカルでビルド

### Windows (PowerShell)

```powershell
$env:CGO_ENABLED=1
go build -o urls-scraper.exe ./cmd/app/...
```

### macOS / Linux

```bash
export CGO_ENABLED=1
go build -o urls-scraper ./cmd/app/...
```

## 実行

```powershell
# Windows
.\urls-scraper.exe
```

```bash
# macOS / Linux
./urls-scraper
```

## テスト

```bash
go test ./...
```

## fyne-cross でクロスビルド（Docker）

このリポジトリには補助スクリプトがあります。

```bash
./scripts/fyne-cross-build.sh windows   # windows/amd64, windows/arm64
./scripts/fyne-cross-build.sh darwin    # darwin/amd64, darwin/arm64（MACOS_SDK_PATH 必須）
./scripts/fyne-cross-build.sh all
```

### macOS 向けの注意

- Linux/Windows ホストから `darwin` をビルドするには、Apple の条件に従って抽出した macOS SDK が必要です。
- `MACOS_SDK_PATH` を設定して実行してください。

例:

```bash
fyne-cross darwin-sdk-extract --xcode-path /path/to/Command_Line_Tools_for_Xcode_12.5.1.dmg
export MACOS_SDK_PATH="$PWD/SDKs/MacOSX11.3.sdk"
./scripts/fyne-cross-build.sh darwin
```

## Docker 上でビルド環境を再現する（`docker/DockerFile-build`）

ローカルに Go や fyne-cross を入れたくない場合は、`docker/DockerFile-build` を使って
`.devcontainer/devcontainer.json` 相当の「Go 1.26 + Docker-in-Docker + fyne-cross」環境を
コンテナとして用意できます。

内部で fyne-cross が Docker daemon を起動するため、**`--privileged` で実行する必要があります**。

### イメージのビルド

リポジトリのルートで実行してください（`docker/` 配下の entrypoint スクリプトも参照されます）。

```bash
docker build -f docker/DockerFile-build -t scraper-bot-fyne-cross .
```

### コンテナでクロスビルド

```bash
# Windows (amd64, arm64)
docker run --privileged --rm \
  -v "$PWD":/workspace \
  scraper-bot-fyne-cross \
  ./scripts/fyne-cross-build.sh windows

# macOS (amd64, arm64) — MACOS_SDK_PATH をホスト側で展開しておき、コンテナにマウント
docker run --privileged --rm \
  -v "$PWD":/workspace \
  -v "$MACOS_SDK_PATH":/sdk:ro \
  -e MACOS_SDK_PATH=/sdk \
  scraper-bot-fyne-cross \
  ./scripts/fyne-cross-build.sh darwin

# fyne-cross を直接呼ぶことも可能
docker run --privileged --rm \
  -v "$PWD":/workspace \
  scraper-bot-fyne-cross \
  fyne-cross windows --pull -arch=amd64,arm64 \
    -app-id github.com/cistec/urls-scraper ./cmd/app
```

成果物はホストの `./fyne-cross/` に出力されます。

### 対話的に使う

開発用に対話シェルへ入りたい場合:

```bash
docker run --privileged --rm -it \
  -v "$PWD":/workspace \
  scraper-bot-fyne-cross
```