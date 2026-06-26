package domain

import (
	"context"
	"encoding/json"

	"meguri-app/internal/infrastructure/persistence"
)

// AppConfigService はアプリ既定設定のドメインサービス。
type AppConfigService struct {
	repo persistence.Repository
}

// NewAppConfigService は AppConfigService を構築する。
func NewAppConfigService(repo persistence.Repository) *AppConfigService {
	return &AppConfigService{repo: repo}
}

// Bootstrap は app_config の初期行を保証する。
func (s *AppConfigService) Bootstrap(ctx context.Context) error {
	return s.repo.BootstrapAppConfig(ctx)
}

// GetDefaults は defaults_json を返す。
func (s *AppConfigService) GetDefaults(ctx context.Context) (json.RawMessage, error) {
	row, err := s.repo.GetAppConfig(ctx)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return json.RawMessage([]byte("{}")), nil
	}
	return json.RawMessage(row.DefaultsJSON), nil
}

// SaveDefaults は defaults_json を保存する。
func (s *AppConfigService) SaveDefaults(ctx context.Context, raw json.RawMessage) error {
	return s.repo.SaveAppConfig(ctx, string(raw))
}
