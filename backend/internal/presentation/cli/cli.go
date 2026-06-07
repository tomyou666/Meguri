// Package cli は CLI プレゼンテーション層の入口。
package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"scraperbot/internal/app"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/infrastructure/configloader"
	"scraperbot/internal/logger"
	"scraperbot/pkg/runner"
)

// App は CLI 実行に必要な I/O 依存をまとめる。テスト時はここを差し替える。
type App struct {
	// Args は os.Args[1:] 相当の CLI 引数。
	Args []string
	// Stdout は標準出力の代替（テスト用）。
	Stdout io.Writer
	// Stderr は標準エラーの代替（テスト用）。
	Stderr io.Writer
}

// Run は CLI のメインエントリ。終了コードを返す。
func Run() int {
	return (&App{
		Args:   os.Args[1:],
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}).RunApp()
}

// RunApp は設定読み込み・Kernel 初期化・単一 URL またはクロールを実行する。
func (a *App) RunApp() int {
	logger.Init(a.Stderr, slog.LevelInfo)

	flags, err := ParseArgs(a.Args)
	if err != nil {
		slog.Error("引数エラー", "err", err)
		return 2
	}

	cfg := model.Default()
	if flags.ConfigPath != "" {
		loaded, err := configloader.LoadYAMLFile(flags.ConfigPath)
		if err != nil {
			slog.Error("設定読み込みエラー", "err", err)
			return 2
		}
		cfg = *loaded
	}
	Merge(&cfg, flags)

	if err := cfg.Validate(); err != nil {
		slog.Error("設定検証エラー", "err", err)
		return 2
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if cfg.Crawl.Enabled {
		return a.runCrawl(ctx, &cfg, flags)
	}

	application, cleanup, err := app.Initialize(ctx, &cfg)
	if err != nil {
		slog.Error("Kernel初期化エラー", "err", err)
		return 1
	}
	defer cleanup()
	return a.runSingle(ctx, &cfg, application, flags)
}

// runSingle は単一 URL モードでスクレイプしファイルまたは標準出力へ出す。
func (a *App) runSingle(ctx context.Context, cfg *model.Config, application *app.Application, flags *Flags) int {
	res, err := application.Scrape.Run(ctx, cfg.Targets[0])
	if err != nil {
		slog.Error("スクレイピング失敗", "err", err)
		return 1
	}
	if flags.Stdout {
		fmt.Fprintln(a.Stdout, res.Markdown)
		return 0
	}
	if err := application.FileWriter.Write(res); err != nil {
		slog.Error("出力書き込み失敗", "err", err)
		return 1
	}
	slog.Info("保存完了", "dir", cfg.Output.Dir, "formats", cfg.Content.Formats)
	return 0
}

// runCrawl は pkg/runner 経由で BFS クロールを実行する。
func (a *App) runCrawl(ctx context.Context, cfg *model.Config, flags *Flags) int {
	var progress runner.ProgressSink
	if flags.ProgressJSON {
		progress = newProgressJSONSink(a.Stderr)
	}
	stats, err := runner.CrawlWithProgress(ctx, cfg, cfg.Targets, progress, nil)
	if err != nil {
		slog.Error("クロール失敗", "err", err)
		return 1
	}
	slog.Info("クロール完了",
		"enqueued", stats.Enqueued,
		"succeeded", stats.Succeeded,
		"failed", stats.Failed,
		"skipped", stats.Skipped,
	)
	return 0
}
