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
func newBrowserSession(parent context.Context, key sessionKey, opts []chromedp.ExecAllocatorOption) (*browserSession, error) {
	job, jobOpt := newBrowserJob()
	opts = append(opts, jobOpt)

	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, opts...)
	rootCtx, rootCancel := chromedp.NewContext(allocCtx)

	if err := chromedp.Run(rootCtx); err != nil {
		rootCancel()
		allocCancel()
		job.Close()
		return nil, fmt.Errorf("start chromium browser: %w", err)
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
func (s *browserSession) openTab(ctx context.Context) (context.Context, context.CancelFunc, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rootCtx == nil {
		return nil, nil, fmt.Errorf("chromium browser session closed")
	}
	tabCtx, tabCancel := chromedp.NewContext(s.rootCtx)
	return tabCtx, tabCancel, nil
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
