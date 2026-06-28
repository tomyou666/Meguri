//go:build dev

package wails_service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const envUpdaterSwapTarget = "MEGURI_UPDATER_SWAP_TARGET"

const (
	envHelperMode   = "WAILS_UPDATER_HELPER"
	envHelperTarget = "WAILS_UPDATER_HELPER_TARGET"
	envHelperNew    = "WAILS_UPDATER_HELPER_NEW"
	envHelperPID    = "WAILS_UPDATER_HELPER_PID"
	envHelperLog    = "WAILS_UPDATER_HELPER_LOG"
)

// restart はステージ済み更新を適用してアプリを再起動する。
//
// MEGURI_UPDATER_SWAP_TARGET が設定されている場合は差し替え先 exe を上書きする。
func (s *UpdateService) restart(ctx context.Context) error {
	if s.app == nil || s.app.Updater == nil {
		return ErrUpdaterUnavailable
	}
	target := os.Getenv(envUpdaterSwapTarget)
	if target != "" {
		return s.restartWithSwapTarget(target)
	}
	return s.app.Updater.Restart(ctx)
}

func (s *UpdateService) restartWithSwapTarget(swapTarget string) error {
	staged := s.app.Updater.DownloadedPath()
	if staged == "" {
		return fmt.Errorf("updater: no staged update")
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("updater: resolve self: %w", err)
	}

	logPath := filepath.Join(os.TempDir(), fmt.Sprintf("wails-update-%d.log", os.Getpid()))
	env := append(os.Environ(),
		envHelperMode+"=1",
		envHelperTarget+"="+swapTarget,
		envHelperNew+"="+staged,
		envHelperPID+"="+fmt.Sprint(os.Getpid()),
		envHelperLog+"="+logPath,
	)

	cmd := newDetachedHelperCommand(self)
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("updater: spawn helper: %w", err)
	}
	s.app.Quit()
	return nil
}

func newDetachedHelperCommand(path string) *exec.Cmd {
	cmd := exec.Command(path)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	applyDetachedHelperAttrs(cmd)
	return cmd
}
