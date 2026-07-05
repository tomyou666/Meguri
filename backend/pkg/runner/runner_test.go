package runner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
	"meguri/pkg/runner"
)

// TestCrawlWithProgress は runner 委譲ラッパーの smoke test。
func TestCrawlWithProgress(t *testing.T) {
	t.Run("異常系: シード URL なしではエラー", func(t *testing.T) {
		cfg := &model.Config{}
		_, err := runner.CrawlWithProgress(context.Background(), cfg, nil, nil, nil)
		require.Error(t, err)
	})
}
