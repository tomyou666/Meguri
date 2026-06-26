package wails_service

import (
	"fmt"
	"sync"

	"meguri-app/internal/model"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	exportPreviewWindowName = "export-preview"
	topicExportOpen         = "export:open"
	exportPreviewURL        = "/?view=export"
)

// ExportWindowManager はエクスポートウィンドウを管理する。
type ExportWindowManager struct {
	app           *application.App
	mainWindow    application.Window
	previewWindow *application.WebviewWindow
	snapshot      model.ExportSessionRequest
	hasSnapshot   bool
	mu            sync.Mutex
}

// NewExportWindowManager は ExportWindowManager を構築する。
func NewExportWindowManager(app *application.App) *ExportWindowManager {
	return &ExportWindowManager{app: app}
}

// SetMainWindow はメインウィンドウを登録し、終了時にエクスポートウィンドウを閉じる。
func (m *ExportWindowManager) SetMainWindow(w application.Window) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mainWindow = w
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		m.closePreviewLocked()
	})
}

// Show はスナップショットを保存し、エクスポートウィンドウを表示する。
func (m *ExportWindowManager) Show(req model.ExportSessionRequest) error {
	if m.app == nil {
		return fmt.Errorf("app not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.snapshot = req
	m.hasSnapshot = true

	preview, err := m.ensurePreviewWindowLocked()
	if err != nil {
		return err
	}
	title := req.Title
	if title == "" {
		title = "Export"
	}
	preview.SetTitle(title)
	preview.Show()
	preview.Focus()
	m.app.Event.Emit(topicExportOpen, req)
	return nil
}

// GetSnapshot は直近のエクスポートセッションスナップショットを返す。
func (m *ExportWindowManager) GetSnapshot() (model.ExportSessionRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.hasSnapshot {
		return model.ExportSessionRequest{}, fmt.Errorf("no export snapshot")
	}
	return m.snapshot, nil
}

func (m *ExportWindowManager) ensurePreviewWindowLocked() (*application.WebviewWindow, error) {
	if m.previewWindow != nil {
		return m.previewWindow, nil
	}
	if existing, ok := m.app.Window.GetByName(exportPreviewWindowName); ok {
		if wv, ok := existing.(*application.WebviewWindow); ok {
			m.previewWindow = wv
			return m.previewWindow, nil
		}
	}

	title := m.snapshot.Title
	if title == "" {
		title = "Export"
	}
	w := m.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             exportPreviewWindowName,
		Title:            title,
		Width:            1100,
		Height:           760,
		InitialPosition:  application.WindowCentered,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              exportPreviewURL,
	})
	m.previewWindow = w
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		m.mu.Lock()
		m.previewWindow = nil
		m.mu.Unlock()
	})
	return w, nil
}

func (m *ExportWindowManager) closePreviewLocked() {
	if m.previewWindow == nil {
		if existing, ok := m.app.Window.GetByName(exportPreviewWindowName); ok {
			existing.Close()
		}
		m.previewWindow = nil
		return
	}
	m.previewWindow.Close()
	m.previewWindow = nil
}
