package main

import (
	"context"
	"embed"
	"log"
	"time"

	"meguri-app/internal/app"
	"meguri-app/internal/usecase/wails_service"
	"meguri/pkg/logger"

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

const githubRepository = "tomyou666/Meguri"

const periodicUpdateCheckInterval = 6 * time.Hour

func main() {
	if err := app.InitLogger(); err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = logger.Flush()
		_ = logger.Shutdown()
	}()

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
	wails_service.WireUpdateMainWindow(updateSvc, mainWindow)

	appMenu := webApp.Menu.New()
	appMenu.Add("更新を確認…").OnClick(func(*application.Context) {
		go handleNativeMenuUpdateCheck(updateSvc, webApp)
	})
	webApp.Menu.SetApplicationMenu(appMenu)

	go updateSvc.CheckOnStartup()
	updateSvc.StartPeriodicCheck(periodicUpdateCheckInterval)

	if err := webApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func handleNativeMenuUpdateCheck(updateSvc *wails_service.UpdateService, webApp *application.App) {
	result, err := updateSvc.CheckForUpdates()
	if err != nil {
		webApp.Logger.Error("update check", "error", err)
		return
	}
	switch result.Action {
	case wails_service.PromptActionConfirmed:
		if err := updateSvc.ApplyUpdate(); err != nil {
			webApp.Logger.Error("apply update", "error", err)
		}
	case wails_service.PromptActionOpenRelease:
		if result.ReleaseURL != "" {
			if err := webApp.Browser.OpenURL(result.ReleaseURL); err != nil {
				webApp.Logger.Error("open release url", "error", err)
			}
		}
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
		CheckInterval:  0,
		Window:         updater.WindowNone,
	})
}
