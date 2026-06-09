package fetchlimit_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/internal/core/fetchlimit"
	"scraperbot/internal/domain/model"
)

// TestFetchLimiter は同時実行上限と解放を検証する。
func TestFetchLimiter(t *testing.T) {
	t.Run("正常系: 上限まで acquire でき release で再利用できる", func(t *testing.T) {
		lim := fetchlimit.NewFromConfig(model.FetchLimitsConfig{
			HTTPMaxInflight: 2,
		})
		ctx := context.Background()
		require.NoError(t, lim.Acquire(ctx, model.FetcherHTTP))
		require.NoError(t, lim.Acquire(ctx, model.FetcherHTTP))
		lim.Release(model.FetcherHTTP)
		require.NoError(t, lim.Acquire(ctx, model.FetcherHTTP))
		lim.Release(model.FetcherHTTP)
		lim.Release(model.FetcherHTTP)
	})

	t.Run("正常系: 上限超過は他 goroutine の release まで待つ", func(t *testing.T) {
		lim := fetchlimit.NewFromConfig(model.FetchLimitsConfig{
			HTTPMaxInflight: 1,
		})
		ctx := context.Background()
		require.NoError(t, lim.Acquire(ctx, model.FetcherHTTP))

		acquired := make(chan struct{})
		go func() {
			require.NoError(t, lim.Acquire(ctx, model.FetcherHTTP))
			close(acquired)
			lim.Release(model.FetcherHTTP)
		}()

		time.Sleep(50 * time.Millisecond)
		select {
		case <-acquired:
			t.Fatal("second acquire should block until release")
		default:
		}
		lim.Release(model.FetcherHTTP)
		select {
		case <-acquired:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for second acquire")
		}
		lim.Release(model.FetcherHTTP)
	})

	t.Run("異常系: ctx キャンセルで acquire が失敗する", func(t *testing.T) {
		lim := fetchlimit.NewFromConfig(model.FetchLimitsConfig{
			HTTPMaxInflight: 1,
		})
		parent, cancel := context.WithCancel(context.Background())
		require.NoError(t, lim.Acquire(parent, model.FetcherHTTP))

		ctx, childCancel := context.WithTimeout(parent, 2*time.Second)
		defer childCancel()
		var wg sync.WaitGroup
		wg.Add(1)
		var acquireErr error
		go func() {
			defer wg.Done()
			acquireErr = lim.Acquire(ctx, model.FetcherHTTP)
		}()
		time.Sleep(30 * time.Millisecond)
		cancel()
		wg.Wait()
		require.Error(t, acquireErr)
		lim.Release(model.FetcherHTTP)
	})

	t.Run("正常系: SetChromiumCapacity で上限を変更できる", func(t *testing.T) {
		lim := fetchlimit.NewFromConfig(model.FetchLimitsConfig{
			ChromiumMaxInflight: 1,
		})
		lim.SetChromiumCapacity(3)
		assert.Equal(t, 3, lim.ChromiumCapacity())
	})
}
