package fetchlimit_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"scraperbot/internal/core/fetchlimit"
	"scraperbot/internal/domain/model"
)

// TestCalibrateChromium は browser path 空時のフォールバックを検証する。
func TestCalibrateChromium(t *testing.T) {
	lim := fetchlimit.NewFromConfig(model.FetchLimitsConfig{ChromiumMaxInflight: 5})
	err := fetchlimit.CalibrateChromium(context.Background(), lim, "")
	assert.NoError(t, err)
	assert.Equal(t, model.DefaultChromiumMaxInflight, lim.ChromiumCapacity())
}
