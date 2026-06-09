package httpfetch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestIsRetryableHTTPError は HTTP リトライ対象エラーの判定を検証する。
func TestIsRetryableHTTPError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "異常系: deadline exceeded はリトライしない", err: context.DeadlineExceeded, want: false},
		{name: "異常系: canceled はリトライしない", err: context.Canceled, want: false},
		{name: "異常系: ラップされた deadline はリトライしない", err: fmt.Errorf("get: %w", context.DeadlineExceeded), want: false},
		{name: "異常系: DNS タイムアウトはリトライしない", err: &net.DNSError{IsTimeout: true}, want: false},
		{name: "正常系: connection refused はリトライする", err: errors.New("connection refused"), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isRetryableHTTPError(tt.err); got != tt.want {
				t.Fatalf("isRetryableHTTPError() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("異常系: net.Error の Timeout() はリトライしない", func(t *testing.T) {
		t.Parallel()
		err := &timeoutNetErr{}
		assert.False(t, isRetryableHTTPError(err))
	})

	t.Run("異常系: context.WithTimeout 由来の deadline はリトライしない", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		<-ctx.Done()
		assert.False(t, isRetryableHTTPError(ctx.Err()))
	})
}

type timeoutNetErr struct{}

func (e *timeoutNetErr) Error() string   { return "timeout" }
func (e *timeoutNetErr) Timeout() bool   { return true }
func (e *timeoutNetErr) Temporary() bool { return false }
