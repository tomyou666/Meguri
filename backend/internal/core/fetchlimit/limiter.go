package fetchlimit

import (
	"context"
	"sync"

	"scraperbot/internal/domain/model"
)

// FetchLimiter はフェッチャ種別ごとの同時取得数を制限する。
type FetchLimiter struct {
	mu sync.Mutex

	httpSem        *kindLimiter
	chromiumSem    *kindLimiter
	chromiumCap    int
	chromiumMax    int
	staticChromium int

	stopDynamic chan struct{}
	dynamicDone chan struct{}
}

// NewFromConfig は設定から FetchLimiter を構築する。
func NewFromConfig(cfg model.FetchLimitsConfig) *FetchLimiter {
	httpCap := cfg.EffectiveHTTPMaxInflight()
	chromiumCap := cfg.EffectiveChromiumMaxInflight()
	return &FetchLimiter{
		httpSem:        newKindLimiter(httpCap),
		chromiumSem:    newKindLimiter(chromiumCap),
		chromiumCap:    chromiumCap,
		chromiumMax:    chromiumCap,
		staticChromium: chromiumCap,
	}
}

// Acquire は指定フェッチャ種別の取得スロットを確保する。
func (l *FetchLimiter) Acquire(ctx context.Context, kind model.FetcherKind) error {
	if kind == "" {
		kind = model.FetcherHTTP
	}
	switch kind {
	case model.FetcherChromium:
		return l.chromiumSem.acquire(ctx)
	default:
		return l.httpSem.acquire(ctx)
	}
}

// Release は Acquire で確保したスロットを解放する。
func (l *FetchLimiter) Release(kind model.FetcherKind) {
	if kind == "" {
		kind = model.FetcherHTTP
	}
	switch kind {
	case model.FetcherChromium:
		l.chromiumSem.release()
	default:
		l.httpSem.release()
	}
}

// SetChromiumCapacity は Chromium の同時実行上限を変更する（1〜8 にクランプ）。
func (l *FetchLimiter) SetChromiumCapacity(n int) {
	if n < 1 {
		n = 1
	}
	if n > model.MaxChromiumMaxInflight {
		n = model.MaxChromiumMaxInflight
	}
	l.mu.Lock()
	l.chromiumCap = n
	if n > l.chromiumMax {
		l.chromiumMax = n
	}
	l.chromiumSem.setCapacity(n)
	l.mu.Unlock()
}

// ChromiumCapacity は現在の Chromium 同時実行上限を返す。
func (l *FetchLimiter) ChromiumCapacity() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.chromiumCap
}

// StaticChromiumLimit は設定由来の Chromium 静的上限を返す。
func (l *FetchLimiter) StaticChromiumLimit() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.staticChromium
}

// Close は動的監視 goroutine を停止する。
func (l *FetchLimiter) Close() {
	if l.stopDynamic == nil {
		return
	}
	close(l.stopDynamic)
	<-l.dynamicDone
	l.stopDynamic = nil
	l.dynamicDone = nil
}

type kindLimiter struct {
	mu       sync.Mutex
	cond     *sync.Cond
	capacity int
	inFlight int
}

func newKindLimiter(capacity int) *kindLimiter {
	k := &kindLimiter{capacity: capacity}
	k.cond = sync.NewCond(&k.mu)
	return k
}

func (k *kindLimiter) setCapacity(n int) {
	k.mu.Lock()
	k.capacity = n
	k.cond.Broadcast()
	k.mu.Unlock()
}

func (k *kindLimiter) acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	cancelled := make(chan struct{})
	defer close(cancelled)
	go func() {
		select {
		case <-ctx.Done():
			k.cond.Broadcast()
		case <-cancelled:
		}
	}()

	k.mu.Lock()
	defer k.mu.Unlock()
	for k.inFlight >= k.capacity {
		if err := ctx.Err(); err != nil {
			return err
		}
		k.cond.Wait()
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	k.inFlight++
	return nil
}

func (k *kindLimiter) release() {
	k.mu.Lock()
	if k.inFlight > 0 {
		k.inFlight--
	}
	k.mu.Unlock()
	k.cond.Broadcast()
}
