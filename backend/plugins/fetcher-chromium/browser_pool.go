package chromiumfetch

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/chromedp/chromedp"
)

// sessionKey はブラウザセッションを識別するキー。
// browserPath / headless / userAgent の組み合わせで一意になる。
type sessionKey string

// makeSessionKey はブラウザ起動条件から sessionKey を組み立てる。
func makeSessionKey(browserPath string, headless bool, userAgent string) sessionKey {
	return sessionKey(browserPath + "|" + strconv.FormatBool(headless) + "|" + userAgent)
}

// browserPool は sessionKey ごとに共有ブラウザセッションを管理する。
type browserPool struct {
	// mu は entries の読み書きを保護する。
	mu sync.Mutex
	// entries は sessionKey ごとのブラウザセッション。
	entries map[sessionKey]*browserSession
}

// defaultBrowserPool はプロセス全体で共有するブラウザプール。
var defaultBrowserPool browserPool = browserPool{
	entries: map[sessionKey]*browserSession{},
}

// addClient は key に対応するセッションへクライアントを参加させる。
// 未作成なら新規セッションを起動する。
func (p *browserPool) addClient(ctx context.Context, key sessionKey, opts []chromedp.ExecAllocatorOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ent, ok := p.entries[key]; ok {
		ent.clients++
		return nil
	}
	sess, err := newBrowserSession(ctx, key, opts)
	if err != nil {
		return err
	}
	p.entries[key] = sess
	return nil
}

// removeClient は key のセッションからクライアントを離脱させる。
// 参照が 0 になったらセッションを閉じる。
func (p *browserPool) removeClient(key sessionKey) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ent, ok := p.entries[key]
	if !ok {
		return
	}
	ent.clients--
	if ent.clients <= 0 {
		ent.close()
		delete(p.entries, key)
	}
}

// openTab は key の共有セッション上に新しいタブを開く。
func (p *browserPool) openTab(ctx context.Context, key sessionKey) (context.Context, context.CancelFunc, error) {
	p.mu.Lock()
	ent, ok := p.entries[key]
	p.mu.Unlock()
	if !ok || ent == nil {
		return nil, nil, fmt.Errorf("chromium browser session not found")
	}
	return ent.openTab(ctx)
}

// joinPool は client を defaultBrowserPool に参加させる。
// 既に参加済みで UA が変わった場合はエラーを返す。
func (c *client) joinPool(ctx context.Context, ua string) error {
	c.poolMu.Lock()
	defer c.poolMu.Unlock()
	key := c.sessionKeyFor(ua)
	if c.poolJoined {
		if c.poolKey != key {
			return fmt.Errorf("inconsistent chromium user-agent in single fetcher client")
		}
		return nil
	}
	if err := defaultBrowserPool.addClient(ctx, key, c.chromedpAllocatorOptions(ua)); err != nil {
		return err
	}
	c.poolKey = key
	c.poolJoined = true
	return nil
}

// leavePool は client を defaultBrowserPool から離脱させる。
func (c *client) leavePool() {
	c.poolMu.Lock()
	defer c.poolMu.Unlock()
	if !c.poolJoined {
		return
	}
	defaultBrowserPool.removeClient(c.poolKey)
	c.poolJoined = false
}

// sessionKeyFor は client 設定と UA から sessionKey を返す。
func (c *client) sessionKeyFor(ua string) sessionKey {
	return makeSessionKey(c.browserPath, c.fetcherCfg.Headless, ua)
}

// runWithTab は共有セッション上でタブを開き、run を実行して閉じる。
func (c *client) runWithTab(ctx context.Context, ua string, run func(context.Context) error) error {
	key := c.sessionKeyFor(ua)
	tabCtx, tabCancel, err := defaultBrowserPool.openTab(ctx, key)
	if err != nil {
		return err
	}
	defer closeTab(tabCtx, tabCancel)
	return run(tabCtx)
}
