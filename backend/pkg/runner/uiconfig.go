package runner

import (
	"encoding/json"

	"meguri/internal/domain/model"
	"meguri/internal/usecase"
)

// MergeUIConfigLayers は UI 由来の PartialConfig JSON を深くマージする。
func MergeUIConfigLayers(layers ...json.RawMessage) (json.RawMessage, error) {
	return usecase.MergeUIConfigLayers(layers...)
}

// ParseUIConfig はマージ済み UI JSON を backend model.Config に変換する。
func ParseUIConfig(raw json.RawMessage) (*model.Config, error) {
	return usecase.ParseUIConfig(raw)
}

// DeriveContentFormats は transformer と extract フラグから content.formats を導出する。
func DeriveContentFormats(cfg *model.Config) {
	usecase.DeriveContentFormats(cfg)
}
