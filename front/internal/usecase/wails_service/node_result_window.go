package wails_service

import (
	"fmt"
	"sync"

	"meguri-app/internal/model"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	nodeResultPreviewWindowName = "node-result-preview"
	mainWindowName              = "main"
	topicNodeResultMaximize     = "node-result:maximize"
	nodeResultPreviewURL        = "/?view=maximized-node-result"
)

// NodeResultWindowManager はノード結果最大化ウィンドウを管理する。
type NodeResultWindowManager struct {
	app           *application.App
	mainWindow    application.Window
	previewWindow *application.WebviewWindow
	snapshot      model.MaximizedNodeResultRequest
	hasSnapshot   bool
	mu            sync.Mutex
}

// NewNodeResultWindowManager は NodeResultWindowManager を構築する。
func NewNodeResultWindowManager(app *application.App) *NodeResultWindowManager {
	return &NodeResultWindowManager{app: app}
}

// SetMainWindow はメインウィンドウを登録し、終了時にプレビューを閉じる。
func (m *NodeResultWindowManager) SetMainWindow(w application.Window) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mainWindow = w
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		m.closePreviewLocked()
	})
}

// Show はスナップショットを保存し、最大化ウィンドウを表示する。
func (m *NodeResultWindowManager) Show(req model.MaximizedNodeResultRequest) error {
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
	if req.Title != "" {
		preview.SetTitle(req.Title)
	}
	preview.Show()
	preview.Focus()
	m.app.Event.Emit(topicNodeResultMaximize, req)
	return nil
}

// GetSnapshot は直近の最大化スナップショットを返す。
func (m *NodeResultWindowManager) GetSnapshot() (model.MaximizedNodeResultRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.hasSnapshot {
		return model.MaximizedNodeResultRequest{}, fmt.Errorf("no maximized snapshot")
	}
	return m.snapshot, nil
}

func (m *NodeResultWindowManager) ensurePreviewWindowLocked() (*application.WebviewWindow, error) {
	if m.previewWindow != nil {
		return m.previewWindow, nil
	}
	if existing, ok := m.app.Window.GetByName(nodeResultPreviewWindowName); ok {
		if wv, ok := existing.(*application.WebviewWindow); ok {
			m.previewWindow = wv
			return m.previewWindow, nil
		}
	}

	title := m.snapshot.Title
	if title == "" {
		title = "Node Result"
	}
	w := m.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             nodeResultPreviewWindowName,
		Title:            title,
		Width:            960,
		Height:           720,
		InitialPosition:  application.WindowCentered,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              nodeResultPreviewURL,
	})
	m.previewWindow = w
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		m.mu.Lock()
		m.previewWindow = nil
		m.mu.Unlock()
	})
	return w, nil
}

func (m *NodeResultWindowManager) closePreviewLocked() {
	if m.previewWindow == nil {
		if existing, ok := m.app.Window.GetByName(nodeResultPreviewWindowName); ok {
			existing.Close()
		}
		m.previewWindow = nil
		return
	}
	m.previewWindow.Close()
	m.previewWindow = nil
}
