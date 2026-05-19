// Package configloader は YAML 設定ファイルから model.Config を構築する。
package configloader

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"scraperbot/internal/domain/model"
)

// LoadYAMLFile は指定ファイルの YAML を読み込み、デフォルト値とマージした Config を返す。
// Validate は呼ばれない（呼び出し側がCLIフラグマージ後に行う想定）。
func LoadYAMLFile(path string) (*model.Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	return ParseYAML(b)
}

// ParseYAML はバイト列の YAML を Default() にマージして返す。
func ParseYAML(b []byte) (*model.Config, error) {
	cfg := model.Default()
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}
	return &cfg, nil
}
