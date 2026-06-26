package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"
	"time"

	"meguri-app/internal/app"
	"meguri-app/internal/logger"
	"meguri-app/internal/usecase/wails_service"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"

	_ "meguri/pkg/runner"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed updater-key.pub
var updaterPublicKey []byte

// currentVersion は CI が -ldflags "-X main.currentVersion=1.0.0" で注入する。
var currentVersion = "dev"

const githubRepository = "tomyou666/scraper-bot"

func main() {
	logger.Init(os.Stderr, slog.LevelInfo)

	ctx := context.Background()
	wailsApp, cleanup, err := app.Initialize(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	updateSvc := wails_service.NewUpdateService()

	webApp := application.New(application.Options{
		Name:        "meguri",
		Description: "Meguri desktop UI",
		Services: []application.Service{
			application.NewService(wailsApp.StoreService),
			application.NewService(wailsApp.ProjectService),
			application.NewService(wailsApp.ScraperService),
			application.NewService(updateSvc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	wailsApp.ProjectService.SetApp(webApp)
	wailsApp.ScraperService.SetApp(webApp)
	wailsApp.StoreService.SetApp(webApp)
	updateSvc.SetApp(webApp)

	if err := initUpdater(webApp); err != nil {
		log.Fatal(err)
	}

	mainWindow := webApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "Meguri",
		Width:            1024,
		Height:           680,
		InitialPosition:  application.WindowCentered,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
	})
	wails_service.WireMainWindow(wailsApp.StoreService, mainWindow)

	appMenu := webApp.Menu.New()
	appMenu.Add("更新を確認…").OnClick(func(*application.Context) {
		go func() {
			if err := webApp.Updater.CheckAndInstall(context.Background()); err != nil {
				webApp.Logger.Error("update check and install", "error", err)
			}
		}()
	})
	webApp.Menu.SetApplicationMenu(appMenu)

	go func() {
		if _, err := webApp.Updater.Check(context.Background()); err != nil {
			webApp.Logger.Error("update check", "error", err)
		}
	}()

	if err := webApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func initUpdater(webApp *application.App) error {
	gh, err := github.New(github.Config{
		Repository:    githubRepository,
		ChecksumAsset: "SHA256SUMS",
	})
	if err != nil {
		return err
	}

	return webApp.Updater.Init(updater.Config{
		CurrentVersion: currentVersion,
		PublicKey:      updaterPublicKey,
		Providers:      []updater.Provider{gh},
		CheckInterval:  6 * time.Hour,
		Window:         updater.WindowNone,
	})
}
