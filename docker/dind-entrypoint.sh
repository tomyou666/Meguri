#!/usr/bin/env bash
# fyne-cross は内部で Docker daemon を必要とするため、コンテナ起動時に
# dockerd をバックグラウンドで起動してから渡されたコマンドを実行する。
#
# 注意: このスクリプトを使うコンテナは `--privileged` で起動する必要がある。
set -e

if ! pgrep -x dockerd >/dev/null 2>&1; then
    dockerd >/var/log/dockerd.log 2>&1 &

    for _ in $(seq 1 30); do
        if docker info >/dev/null 2>&1; then
            break
        fi
        sleep 1
    done

    if ! docker info >/dev/null 2>&1; then
        echo "dockerd の起動に失敗しました。コンテナを --privileged で起動しているか確認してください。" >&2
        echo "--- /var/log/dockerd.log ---" >&2
        tail -n 50 /var/log/dockerd.log >&2 || true
        exit 1
    fi
fi

exec "$@"
