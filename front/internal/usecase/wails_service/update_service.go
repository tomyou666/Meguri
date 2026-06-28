package wails_service

import (
	"context"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
)

// UpdateService は Wails updater の手動確認 RPC。
type UpdateService struct {
	app       *application.App
	promptWin *UpdateWindowManager

	mu             sync.RWMutex
	pendingRelease *updater.Release
}

// NewUpdateService は UpdateService を構築する。
func NewUpdateService() *UpdateService {
	return &UpdateService{}
}

// SetApp は Wails App を後から注入する。
func (s *UpdateService) SetApp(app *application.App) {
	s.app = app
	s.promptWin = NewUpdateWindowManager(app)
}

// WireUpdateMainWindow は UpdateService にメインウィンドウを注入する。
func WireUpdateMainWindow(s *UpdateService, window application.Window) {
	if s.promptWin != nil {
		s.promptWin.SetMainWindow(window)
	}
}

func (s *UpdateService) ctx() context.Context {
	return context.Background()
}

func (s *UpdateService) requireUpdater() error {
	if s.app == nil || s.app.Updater == nil {
		return ErrUpdaterUnavailable
	}
	return nil
}

// Check は更新を確認し、利用可能なら pending に保持する（ダウンロードはしない）。
func (s *UpdateService) Check() (UpdateStatus, error) {
	if err := s.requireUpdater(); err != nil {
		return UpdateStatus{Status: updateStatusUnavailable}, err
	}

	rel, err := s.app.Updater.Check(s.ctx())
	if err != nil {
		return UpdateStatus{}, err
	}

	s.mu.Lock()
	if rel == nil {
		s.pendingRelease = nil
		s.mu.Unlock()
		return UpdateStatus{Status: updateStatusUpToDate}, nil
	}
	s.pendingRelease = rel
	s.mu.Unlock()

	return UpdateStatus{
		Status:     updateStatusAvailable,
		Version:    rel.Version,
		ReleaseURL: releaseURLFrom(rel),
	}, nil
}

// GetStatus は現在の更新状態を返す。
func (s *UpdateService) GetStatus() (UpdateStatus, error) {
	if err := s.requireUpdater(); err != nil {
		return UpdateStatus{Status: updateStatusUnavailable}, err
	}

	if s.app.Updater.State() == updater.StateReady {
		s.mu.RLock()
		rel := s.pendingRelease
		s.mu.RUnlock()
		status := UpdateStatus{Status: updateStatusReady}
		if rel != nil {
			status.Version = rel.Version
			status.ReleaseURL = releaseURLFrom(rel)
		}
		return status, nil
	}

	s.mu.RLock()
	rel := s.pendingRelease
	s.mu.RUnlock()
	if rel != nil {
		return UpdateStatus{
			Status:     updateStatusAvailable,
			Version:    rel.Version,
			ReleaseURL: releaseURLFrom(rel),
		}, nil
	}

	return UpdateStatus{Status: updateStatusUpToDate}, nil
}

// PromptUpdate は別 WebviewWindow で更新確認を行う。
func (s *UpdateService) PromptUpdate() (UpdatePromptResult, error) {
	s.mu.RLock()
	rel := s.pendingRelease
	s.mu.RUnlock()
	if rel == nil {
		if err := s.requireUpdater(); err != nil {
			return UpdatePromptResult{}, err
		}
		return UpdatePromptResult{Action: promptActionDismissed}, nil
	}

	if s.promptWin == nil {
		return UpdatePromptResult{}, ErrUpdaterUnavailable
	}

	version := rel.Version
	releaseURL := releaseURLFrom(rel)
	action, err := s.promptWin.ShowAndWait(version, releaseURL)
	if err != nil {
		return UpdatePromptResult{}, err
	}
	return UpdatePromptResult{
		Action:     action,
		Version:    version,
		ReleaseURL: releaseURL,
	}, nil
}

// SubmitUpdatePrompt は更新確認ウィンドウからの選択を受け取る。
//
// action は confirmed / open_release / dismissed のいずれか。
func (s *UpdateService) SubmitUpdatePrompt(action string) error {
	if s.promptWin == nil {
		return ErrUpdaterUnavailable
	}
	return s.promptWin.Submit(action)
}

// GetUpdatePromptSnapshot は更新確認ウィンドウ用の直近スナップショットを返す。
func (s *UpdateService) GetUpdatePromptSnapshot() (UpdatePromptSnapshot, error) {
	if s.promptWin == nil {
		return UpdatePromptSnapshot{}, ErrUpdaterUnavailable
	}
	return s.promptWin.GetSnapshot()
}

// ApplyUpdate は pending 更新をダウンロードしてステージし、exe を差し替えて再起動する。
func (s *UpdateService) ApplyUpdate() error {
	if err := s.requireUpdater(); err != nil {
		return err
	}

	s.mu.RLock()
	hasPending := s.pendingRelease != nil
	s.mu.RUnlock()
	if !hasPending && s.app.Updater.State() != updater.StateReady {
		return updater.ErrNoPendingRelease
	}

	if s.app.Updater.State() != updater.StateReady {
		if err := s.app.Updater.DownloadAndInstall(s.ctx()); err != nil {
			return err
		}
	}

	return s.restart(s.ctx())
}

// CheckForUpdates は手動更新確認（Check → 利用可能なら PromptUpdate）を行う。
func (s *UpdateService) CheckForUpdates() (CheckForUpdatesResult, error) {
	status, err := s.Check()
	if err != nil {
		return CheckForUpdatesResult{}, err
	}
	if status.Status == updateStatusUpToDate {
		return CheckForUpdatesResult{Status: updateStatusUpToDate}, nil
	}

	prompt, err := s.PromptUpdate()
	if err != nil {
		return CheckForUpdatesResult{}, err
	}
	return CheckForUpdatesResult{
		Status:     updateStatusAvailable,
		Action:     prompt.Action,
		Version:    prompt.Version,
		ReleaseURL: prompt.ReleaseURL,
	}, nil
}

// CheckOnStartup は起動時の更新確認を行い、利用可能ならイベントを発火する。
func (s *UpdateService) CheckOnStartup() {
	status, err := s.Check()
	if err != nil {
		if s.app != nil {
			s.app.Logger.Error("update check on startup", "error", err)
		}
		return
	}
	if status.Status != updateStatusAvailable {
		return
	}
	if s.app != nil {
		s.app.Event.Emit(topicUpdateAvailable, UpdateAvailableEvent{
			Version:    status.Version,
			ReleaseURL: status.ReleaseURL,
		})
	}
}

// StartPeriodicCheck は interval ごとに Check のみを実行するバックグラウンドループを開始する。
func (s *UpdateService) StartPeriodicCheck(interval time.Duration) {
	if interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			status, err := s.Check()
			if err != nil {
				if s.app != nil {
					s.app.Logger.Error("periodic update check", "error", err)
				}
				continue
			}
			if status.Status != updateStatusAvailable || s.app == nil {
				continue
			}
			s.app.Event.Emit(topicUpdateAvailable, UpdateAvailableEvent{
				Version:    status.Version,
				ReleaseURL: status.ReleaseURL,
			})
		}
	}()
}
