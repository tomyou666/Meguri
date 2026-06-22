package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"

	"scraperbot-front/internal/app"
	"scraperbot-front/internal/logger"
	"scraperbot-front/internal/usecase/wails_service"

	"github.com/wailsapp/wails/v3/pkg/application"

	_ "scraperbot/pkg/runner"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logger.Init(os.Stderr, slog.LevelInfo)

	ctx := context.Background()
	wailsApp, cleanup, err := app.Initialize(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	webApp := application.New(application.Options{
		Name:        "scraperbot",
		Description: "Scraper Bot desktop UI",
		Services: []application.Service{
			application.NewService(wailsApp.StoreService),
			application.NewService(wailsApp.ProjectService),
			application.NewService(wailsApp.ScraperService),
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

	mainWindow := webApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "scraperbot",
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

	if err := webApp.Run(); err != nil {
		log.Fatal(err)
	}
}
