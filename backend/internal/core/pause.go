package core

import (
	"context"
	"sync"
)

// PauseController はクロール実行の一時停止を制御する。
type PauseController struct {
	mu     sync.Mutex
	paused bool
	cond   *sync.Cond
}

// NewPauseController は PauseController を構築する。
func NewPauseController() *PauseController {
	p := &PauseController{}
	p.cond = sync.NewCond(&p.mu)
	return p
}

// Pause は新規ジョブ開始をブロックする。
func (p *PauseController) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = true
}

// Resume は一時停止を解除し、待機中の worker を再開する。
func (p *PauseController) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = false
	p.cond.Broadcast()
}

// WaitIfPaused は一時停止中なら解除または ctx キャンセルまでブロックする。
//
// ctx がキャンセルされた場合は ctx.Err() を返す。
func (p *PauseController) WaitIfPaused(ctx context.Context) error {
	if p == nil {
		return nil
	}
	wake := make(chan struct{})
	defer close(wake)
	go func() {
		select {
		case <-ctx.Done():
			p.cond.Broadcast()
		case <-wake:
		}
	}()

	p.mu.Lock()
	defer p.mu.Unlock()
	for p.paused {
		if err := ctx.Err(); err != nil {
			return err
		}
		p.cond.Wait()
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

// IsPaused は一時停止中かどうかを返す。
func (p *PauseController) IsPaused() bool {
	if p == nil {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.paused
}
