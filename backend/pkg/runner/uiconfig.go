package runner

import (
	"encoding/json"

	"meguri/internal/domain/model"
	"meguri/internal/usecase"
)

//go:generate go tool gowrap gen -p meguri/pkg/runner -i UIConfigurator -t templates/slog_debug.gotmpl -o uiconfigurator_with_debug_log.go

// UIConfigurator は UI 由来設定のマージと変換を抽象化する。
type UIConfigurator interface {
	// MergeUIConfigLayers は UI 由来の PartialConfig JSON を深くマージする。
	MergeUIConfigLayers(layers ...json.RawMessage) (json.RawMessage, error)
	// ParseUIConfig はマージ済み UI JSON を backend model.Config に変換する。
	ParseUIConfig(raw json.RawMessage) (*model.Config, error)
	// DeriveContentFormats は transformer と extract フラグから content.formats を導出する。
	DeriveContentFormats(cfg *model.Config)
}

type uiConfiguratorImpl struct{}

func (uiConfiguratorImpl) MergeUIConfigLayers(layers ...json.RawMessage) (json.RawMessage, error) {
	return usecase.MergeUIConfigLayers(layers...)
}

func (uiConfiguratorImpl) ParseUIConfig(raw json.RawMessage) (*model.Config, error) {
	return usecase.ParseUIConfig(raw)
}

func (uiConfiguratorImpl) DeriveContentFormats(cfg *model.Config) {
	usecase.DeriveContentFormats(cfg)
}

// MergeUIConfigLayers は UI 由来の PartialConfig JSON を深くマージする。
func MergeUIConfigLayers(layers ...json.RawMessage) (json.RawMessage, error) {
	return defaultUIConfigurator.MergeUIConfigLayers(layers...)
}

// ParseUIConfig はマージ済み UI JSON を backend model.Config に変換する。
func ParseUIConfig(raw json.RawMessage) (*model.Config, error) {
	return defaultUIConfigurator.ParseUIConfig(raw)
}

// DeriveContentFormats は transformer と extract フラグから content.formats を導出する。
func DeriveContentFormats(cfg *model.Config) {
	defaultUIConfigurator.DeriveContentFormats(cfg)
}
