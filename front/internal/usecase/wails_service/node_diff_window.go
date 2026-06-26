package wails_service

import (
	"fmt"
	"sync"

	"meguri-app/internal/model"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	nodeDiffWindowName = "node-diff-viewer"
	topicNodeDiffOpen  = "node-diff:open"
	nodeDiffURL        = "/?view=node-diff"
)

// NodeDiffWindowManager はノード差分ビューアウィンドウを管理する。
type NodeDiffWindowManager struct {
	app           *application.App
	previewWindow *application.WebviewWindow
	snapshot      model.NodeDiffViewerRequest
	hasSnapshot   bool
	mu            sync.Mutex
}

// NewNodeDiffWindowManager は NodeDiffWindowManager を構築する。
func NewNodeDiffWindowManager(app *application.App) *NodeDiffWindowManager {
	return &NodeDiffWindowManager{app: app}
}

// SetMainWindow はメインウィンドウ終了時に差分ウィンドウを閉じる。
func (m *NodeDiffWindowManager) SetMainWindow(w application.Window) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		m.closePreviewLocked()
	})
}

// Show はスナップショットを保存し、差分ウィンドウを表示する。
func (m *NodeDiffWindowManager) Show(req model.NodeDiffViewerRequest) error {
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
		title = "Node Diff"
	}
	preview.SetTitle(title)
	preview.Show()
	preview.Focus()
	m.app.Event.Emit(topicNodeDiffOpen, req)
	return nil
}

// GetSnapshot は直近の差分ビューアスナップショットを返す。
func (m *NodeDiffWindowManager) GetSnapshot() (model.NodeDiffViewerRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.hasSnapshot {
		return model.NodeDiffViewerRequest{}, fmt.Errorf("no node diff snapshot")
	}
	return m.snapshot, nil
}

func (m *NodeDiffWindowManager) ensurePreviewWindowLocked() (*application.WebviewWindow, error) {
	if m.previewWindow != nil {
		return m.previewWindow, nil
	}
	if existing, ok := m.app.Window.GetByName(nodeDiffWindowName); ok {
		if wv, ok := existing.(*application.WebviewWindow); ok {
			m.previewWindow = wv
			return m.previewWindow, nil
		}
	}

	title := m.snapshot.Title
	if title == "" {
		title = "Node Diff"
	}
	w := m.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             nodeDiffWindowName,
		Title:            title,
		Width:            960,
		Height:           720,
		InitialPosition:  application.WindowCentered,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              nodeDiffURL,
	})
	m.previewWindow = w
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		m.mu.Lock()
		m.previewWindow = nil
		m.mu.Unlock()
	})
	return w, nil
}

func (m *NodeDiffWindowManager) closePreviewLocked() {
	if m.previewWindow == nil {
		if existing, ok := m.app.Window.GetByName(nodeDiffWindowName); ok {
			existing.Close()
		}
		m.previewWindow = nil
		return
	}
	m.previewWindow.Close()
	m.previewWindow = nil
}
