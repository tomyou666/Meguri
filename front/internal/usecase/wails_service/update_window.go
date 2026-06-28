package wails_service

import (
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	updatePromptWindowName = "update-prompt"
	topicUpdatePromptOpen  = "update-prompt:open"
	updatePromptURL        = "/?view=update-prompt"
	updatePromptTitle      = "更新の確認"
)

type promptWait struct {
	ch   chan string
	once sync.Once
}

// UpdateWindowManager は更新確認用の別 WebviewWindow を管理する。
type UpdateWindowManager struct {
	app          *application.App
	promptWindow *application.WebviewWindow
	snapshot     UpdatePromptSnapshot
	hasSnapshot  bool
	mu           sync.Mutex

	promptMu sync.Mutex
	waitMu   sync.Mutex
	wait     *promptWait
}

// NewUpdateWindowManager は UpdateWindowManager を構築する。
func NewUpdateWindowManager(app *application.App) *UpdateWindowManager {
	return &UpdateWindowManager{app: app}
}

// SetMainWindow はメインウィンドウ終了時に更新確認ウィンドウを閉じる。
func (m *UpdateWindowManager) SetMainWindow(w application.Window) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		m.mu.Lock()
		m.closePromptLocked()
		m.mu.Unlock()
		m.resolve(promptActionDismissed)
	})
}

// ShowAndWait は更新確認ウィンドウを表示し、ユーザーが選択するまでブロックする。
func (m *UpdateWindowManager) ShowAndWait(version, releaseURL string) (string, error) {
	if m.app == nil {
		return "", ErrUpdaterUnavailable
	}

	m.promptMu.Lock()
	defer m.promptMu.Unlock()

	wait := &promptWait{ch: make(chan string, 1)}
	m.waitMu.Lock()
	m.wait = wait
	m.waitMu.Unlock()

	m.mu.Lock()
	m.snapshot = UpdatePromptSnapshot{
		Version:    version,
		ReleaseURL: releaseURL,
	}
	m.hasSnapshot = true

	preview, err := m.ensurePromptWindowLocked()
	m.mu.Unlock()
	if err != nil {
		m.clearWait()
		return "", err
	}

	preview.Show()
	preview.Focus()
	m.app.Event.Emit(topicUpdatePromptOpen, m.snapshot)

	action := <-wait.ch
	m.clearWait()
	return action, nil
}

// Submit は更新確認ウィンドウからのユーザー選択を受け取り、ShowAndWait を解除する。
//
// action は confirmed / open_release / dismissed のいずれか。
func (m *UpdateWindowManager) Submit(action string) error {
	switch action {
	case promptActionConfirmed, promptActionOpenRelease, promptActionDismissed:
	default:
		return fmt.Errorf("invalid update prompt action: %s", action)
	}

	m.resolve(action)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.closePromptLocked()
	return nil
}

// GetSnapshot は直近の更新確認スナップショットを返す。
func (m *UpdateWindowManager) GetSnapshot() (UpdatePromptSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.hasSnapshot {
		return UpdatePromptSnapshot{}, fmt.Errorf("no update prompt snapshot")
	}
	return m.snapshot, nil
}

func (m *UpdateWindowManager) resolve(action string) {
	m.waitMu.Lock()
	wait := m.wait
	m.waitMu.Unlock()
	if wait == nil {
		return
	}
	wait.once.Do(func() {
		wait.ch <- action
	})
}

func (m *UpdateWindowManager) clearWait() {
	m.waitMu.Lock()
	m.wait = nil
	m.waitMu.Unlock()
}

func (m *UpdateWindowManager) ensurePromptWindowLocked() (*application.WebviewWindow, error) {
	if m.promptWindow != nil {
		return m.promptWindow, nil
	}
	if existing, ok := m.app.Window.GetByName(updatePromptWindowName); ok {
		if wv, ok := existing.(*application.WebviewWindow); ok {
			m.promptWindow = wv
			return m.promptWindow, nil
		}
	}

	w := m.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             updatePromptWindowName,
		Title:            updatePromptTitle,
		Width:            480,
		Height:           320,
		InitialPosition:  application.WindowCentered,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              updatePromptURL,
	})
	m.promptWindow = w
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		m.mu.Lock()
		m.promptWindow = nil
		m.mu.Unlock()
		m.resolve(promptActionDismissed)
	})
	return w, nil
}

func (m *UpdateWindowManager) closePromptLocked() {
	if m.promptWindow == nil {
		if existing, ok := m.app.Window.GetByName(updatePromptWindowName); ok {
			existing.Close()
		}
		return
	}
	m.promptWindow.Close()
	m.promptWindow = nil
}
