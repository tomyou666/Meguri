// Package cli は CLI プレゼンテーション層の入口。
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/infrastructure/configloader"
	"scraperbot/internal/infrastructure/httpclient"
	"scraperbot/internal/infrastructure/logging"
	"scraperbot/internal/infrastructure/robots"
	"scraperbot/internal/infrastructure/storage"
	"scraperbot/internal/usecase"
)

// App は CLI 実行に必要な I/O 依存をまとめる。テスト時はここを差し替える。
type App struct {
	Args   []string
	Stdout io.Writer
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

func (a *App) RunApp() int {
	flags, err := ParseArgs(a.Args)
	if err != nil {
		fmt.Fprintln(a.Stderr, "引数エラー:", err)
		return 2
	}

	cfg := model.Default()
	if flags.ConfigPath != "" {
		loaded, err := configloader.LoadYAMLFile(flags.ConfigPath)
		if err != nil {
			fmt.Fprintln(a.Stderr, "設定読み込みエラー:", err)
			return 2
		}
		cfg = *loaded
	}
	Merge(&cfg, flags)

	if err := cfg.Validate(); err != nil {
		fmt.Fprintln(a.Stderr, "設定検証エラー:", err)
		return 2
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := logging.NewDefault()
	client := httpclient.New(cfg.Request)
	host := core.NewHost(logger, &cfg, client)
	kernel := core.NewKernel(&cfg, host, core.Default())
	if err := kernel.Init(ctx); err != nil {
		fmt.Fprintln(a.Stderr, "Kernel初期化エラー:", err)
		return 1
	}
	defer kernel.Close(ctx)

	if cfg.Crawl.Enabled {
		return a.runCrawl(ctx, &cfg, kernel, client, logger)
	}
	return a.runSingle(ctx, &cfg, kernel, client, flags)
}

func (a *App) runSingle(ctx context.Context, cfg *model.Config, k *core.Kernel, client *httpclient.Client, flags *Flags) int {
	uc := usecase.NewScrape(k, client)
	res, err := uc.Run(ctx, cfg.Targets[0])
	if err != nil {
		fmt.Fprintln(a.Stderr, "スクレイピング失敗:", err)
		return 1
	}
	if flags.Stdout {
		fmt.Fprintln(a.Stdout, res.Markdown)
		return 0
	}
	w := storage.NewFileWriter(cfg.Output, cfg.Content.Formats)
	if err := w.Write(res); err != nil {
		fmt.Fprintln(a.Stderr, "出力書き込み失敗:", err)
		return 1
	}
	fmt.Fprintf(a.Stdout, "保存: %s (formats=%v)\n", cfg.Output.Dir, cfg.Content.Formats)
	return 0
}

func (a *App) runCrawl(ctx context.Context, cfg *model.Config, k *core.Kernel, client *httpclient.Client, logger *logging.SlogAdapter) int {
	w := storage.NewFileWriter(cfg.Output, cfg.Content.Formats)
	robotsCache := robots.NewCache(client, logger)

	uc := usecase.NewCrawl(k, client, robotsCache, func(r *model.Result) {
		if err := w.Write(r); err != nil {
			logger.Warn("出力書き込み失敗", "url", r.URL.String(), "err", err.Error())
		}
	})

	stats, _, err := uc.Run(ctx, cfg.Targets)
	if err != nil {
		fmt.Fprintln(a.Stderr, "クロール失敗:", err)
		return 1
	}
	fmt.Fprintf(a.Stdout, "クロール完了: enqueued=%d succeeded=%d failed=%d skipped=%d\n",
		stats.Enqueued, stats.Succeeded, stats.Failed, stats.Skipped)
	return 0
}
