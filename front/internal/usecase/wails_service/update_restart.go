//go:build !dev

package wails_service

import (
	"context"
)

// restart はステージ済み更新を適用してアプリを再起動する。
func (s *UpdateService) restart(ctx context.Context) error {
	if s.app == nil || s.app.Updater == nil {
		return ErrUpdaterUnavailable
	}
	return s.app.Updater.Restart(ctx)
}
