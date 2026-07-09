package chromiumfetch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// browserSession は 1 つの Chromium プロセスとルートコンテキストを保持する。
type browserSession struct {
	// key はこのセッションを識別する sessionKey。
	key sessionKey

	// mu はセッション状態の読み書きを保護する。
	mu sync.Mutex

	// allocCtx は ExecAllocator のコンテキスト。
	allocCtx context.Context
	// allocCancel は allocCtx をキャンセルする。
	allocCancel context.CancelFunc
	// rootCtx はタブ作成の親となるルート chromedp コンテキスト。
	rootCtx context.Context
	// rootCancel は rootCtx をキャンセルする。
	rootCancel context.CancelFunc
	// job はブラウザプロセスのライフサイクル管理。
	job *browserJob
	// clients はこのセッションを参照している client 数。
	clients int
}

// newBrowserSession は Chromium を起動し、共有用の browserSession を返す。
//
// ブラウザ寿命は parent のキャンセルから切り離す（共有プールが最初のリクエスト終了で死なないようにする）。
// 起動自体は parent のキャンセルで中断する。
func newBrowserSession(parent context.Context, key sessionKey, opts []chromedp.ExecAllocatorOption) (*browserSession, error) {
	if err := parent.Err(); err != nil {
		return nil, fmt.Errorf("start chromium browser: %w", err)
	}

	job, jobOpt := newBrowserJob()
	opts = append(opts, jobOpt)

	// 共有セッションはリクエスト context に紐づけない。
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.WithoutCancel(parent), opts...)
	rootCtx, rootCancel := chromedp.NewContext(allocCtx)

	errCh := make(chan error, 1)
	go func() {
		errCh <- chromedp.Run(rootCtx)
	}()

	var runErr error
	select {
	case runErr = <-errCh:
	case <-parent.Done():
		rootCancel()
		allocCancel()
		job.Close()
		return nil, fmt.Errorf("start chromium browser: %w", parent.Err())
	}
	if runErr != nil {
		rootCancel()
		allocCancel()
		job.Close()
		return nil, fmt.Errorf("start chromium browser: %w", runErr)
	}

	return &browserSession{
		key:         key,
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		rootCtx:     rootCtx,
		rootCancel:  rootCancel,
		job:         job,
		clients:     1,
	}, nil
}

// openTab はルートコンテキスト上に新しいタブコンテキストを作る。
//
// ctx がキャンセルされたらタブだけを閉じる（共有ブラウザは維持する）。
func (s *browserSession) openTab(ctx context.Context) (context.Context, context.CancelFunc, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rootCtx == nil {
		return nil, nil, fmt.Errorf("chromium browser session closed")
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	tabCtx, tabCancel := chromedp.NewContext(s.rootCtx)
	stop := context.AfterFunc(ctx, tabCancel)
	cancel := func() {
		stop()
		tabCancel()
	}
	return tabCtx, cancel, nil
}

// close はブラウザプロセスと関連コンテキストを解放する。
func (s *browserSession) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rootCtx == nil {
		return
	}

	if s.rootCancel != nil {
		tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = chromedp.Cancel(tctx)
		cancel()
		s.rootCancel()
		s.rootCancel = nil
	}
	if s.allocCancel != nil {
		s.allocCancel()
		s.allocCancel = nil
	}
	if s.job != nil {
		s.job.Close()
		s.job = nil
	}
	s.rootCtx = nil
	s.allocCtx = nil
}

// closeTab はタブコンテキストをキャンセルし、chromedp 側の後始末を行う。
func closeTab(tabCtx context.Context, tabCancel context.CancelFunc) {
	if tabCancel != nil {
		tabCancel()
	}
	if tabCtx != nil {
		tctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = chromedp.Cancel(tctx)
		cancel()
	}
}
