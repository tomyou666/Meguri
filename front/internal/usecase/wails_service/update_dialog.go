package wails_service

import (
	"fmt"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// showUpdatePrompt はネイティブ更新確認ダイアログを表示する。
//
// Windows は Yes/No のみのため 2 段ダイアログで 3 択を再現する。
// macOS / Linux は 3 ボタンを 1 ダイアログで表示する。
func showUpdatePrompt(app *application.App, window application.Window, version, releaseURL string) (string, error) {
	if app == nil {
		return "", ErrUpdaterUnavailable
	}
	message := buildUpdatePromptMessage(version, releaseURL)
	if runtime.GOOS == "windows" {
		return showUpdatePromptWindows(app, window, message)
	}
	return showUpdatePromptMultiButton(app, window, message)
}

func buildUpdatePromptMessage(version, releaseURL string) string {
	return fmt.Sprintf(
		"バージョン %s の更新が利用可能です。\n\nリリース:\n%s",
		version,
		releaseURL,
	)
}

func showUpdatePromptMultiButton(app *application.App, window application.Window, message string) (string, error) {
	done := make(chan string, 1)

	dlg := app.Dialog.Question().
		SetTitle("更新の確認").
		SetMessage(message).
		AttachToWindow(window)

	confirm := dlg.AddButton("更新して再起動").SetAsDefault()
	openRelease := dlg.AddButton("リリースを開く")
	later := dlg.AddButton("後で").SetAsCancel()

	confirm.OnClick(func() { done <- promptActionConfirmed })
	openRelease.OnClick(func() { done <- promptActionOpenRelease })
	later.OnClick(func() { done <- promptActionDismissed })

	dlg.Show()
	return <-done, nil
}

func showUpdatePromptWindows(app *application.App, window application.Window, message string) (string, error) {
	first := make(chan string, 1)

	dlg := app.Dialog.Question().
		SetTitle("更新の確認").
		SetMessage(message + "\n\n「はい」で更新、「いいえ」でその他の選択へ進みます。").
		AttachToWindow(window)

	yes := dlg.AddButton("Yes").SetAsDefault()
	no := dlg.AddButton("No").SetAsCancel()

	yes.OnClick(func() { first <- promptActionConfirmed })
	no.OnClick(func() { first <- "more" })

	dlg.Show()
	choice := <-first
	if choice == promptActionConfirmed {
		return promptActionConfirmed, nil
	}

	second := make(chan string, 1)
	dlg2 := app.Dialog.Question().
		SetTitle("更新の確認").
		SetMessage("リリースページを開きますか？\n「いいえ」で後で確認できます。").
		AttachToWindow(window)

	yes2 := dlg2.AddButton("Yes").SetAsDefault()
	no2 := dlg2.AddButton("No").SetAsCancel()

	yes2.OnClick(func() { second <- promptActionOpenRelease })
	no2.OnClick(func() { second <- promptActionDismissed })

	dlg2.Show()
	return <-second, nil
}
