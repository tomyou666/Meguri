#!/usr/bin/env bash
# ローカル（Docker 必須）で fyne-cross を使い Windows / macOS 向けバイナリをビルドします。
#
# 使い方:
#   ./scripts/fyne-cross-build.sh windows          # windows/amd64 + windows/arm64
#   ./scripts/fyne-cross-build.sh darwin           # darwin 用（要 MACOS_SDK_PATH）
#   ./scripts/fyne-cross-build.sh all              # 両方（darwin は SDK 必須）
#
# macOS 向け: Apple の利用条件に従い、Xcode Command Line Tools の .dmg から SDK を展開してください。
#   go install github.com/fyne-io/fyne-cross@v1.6.1
#   fyne-cross darwin-sdk-extract --xcode-path /path/to/Command_Line_Tools_for_Xcode_12.5.1.dmg
#   export MACOS_SDK_PATH="$PWD/SDKs/MacOSX11.3.sdk"
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

APP_ID="${FYNE_APP_ID:-github.com/cistec/urls-scraper}"
PKG="./cmd/app"
BUILD_NUM="${CI_PIPELINE_IID:-1}"

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker が PATH にありません。Docker Desktop などを起動してください。" >&2
  exit 1
fi

if ! command -v fyne-cross >/dev/null 2>&1; then
  echo "fyne-cross をインストールしています: go install github.com/fyne-io/fyne-cross@v1.6.1" >&2
  go install github.com/fyne-io/fyne-cross@v1.6.1
  export PATH="$(go env GOPATH)/bin:$PATH"
fi

TARGET="${1:-all}"

build_windows() {
  echo ">>> fyne-cross windows (amd64, arm64)"
  fyne-cross windows --pull \
    -arch=amd64,arm64 \
    -app-id "$APP_ID" \
    -app-build "$BUILD_NUM" \
    "$PKG"
}

build_darwin() {
  if [[ -z "${MACOS_SDK_PATH:-}" ]]; then
    echo "darwin ビルドには環境変数 MACOS_SDK_PATH を設定してください（例: .../MacOSX11.3.sdk）。" >&2
    exit 1
  fi
  if [[ ! -d "$MACOS_SDK_PATH" ]]; then
    echo "MACOS_SDK_PATH がディレクトリではありません: $MACOS_SDK_PATH" >&2
    exit 1
  fi
  echo ">>> fyne-cross darwin (amd64, arm64), SDK=$MACOS_SDK_PATH"
  fyne-cross darwin --pull \
    -arch=amd64,arm64 \
    -app-id "$APP_ID" \
    -app-build "$BUILD_NUM" \
    --macosx-sdk-path "$MACOS_SDK_PATH" \
    "$PKG"
}

case "$TARGET" in
  windows)
    build_windows
    ;;
  darwin)
    build_darwin
    ;;
  all)
    build_windows
    if [[ -n "${MACOS_SDK_PATH:-}" ]]; then
      build_darwin
    else
      echo ">>> MACOS_SDK_PATH 未設定のため darwin はスキップしました。"
    fi
    ;;
  *)
    echo "Usage: $0 [windows|darwin|all]" >&2
    exit 1
    ;;
esac

echo "完了。成果物は $ROOT/fyne-cross/ 以下を参照してください。"
