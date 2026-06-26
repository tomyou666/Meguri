// Package robots は robots.txt のホスト単位キャッシュと許可判定を提供する。
package robots

import (
	"context"
	"log/slog"
	"net/url"
	"sync"

	"github.com/temoto/robotstxt"

	"meguri/internal/domain/plugin"
)

// Cache はホスト単位で robots.txt を一度だけ取得・キャッシュする。
type Cache struct {
	// mu は hosts マップの排他制御。
	mu sync.Mutex
	// hosts は scheme+host キー→パース済み robots データ。
	hosts map[string]*robotstxt.RobotsData
	// fetcher は robots.txt 取得用 Fetcher。
	fetcher plugin.Fetcher
}

// NewCache は Fetcher から robots キャッシュを構築する。
func NewCache(fetcher plugin.Fetcher) *Cache {
	return &Cache{
		hosts:   map[string]*robotstxt.RobotsData{},
		fetcher: fetcher,
	}
}

// Allowed は与えられた URL と User-Agent に対する許可判定。
// 取得失敗・パース失敗は保守的に「許可」として扱う（設計書 05 章方針）。
func (c *Cache) Allowed(ctx context.Context, u *url.URL, ua string) bool {
	if ua == "" {
		ua = "*"
	}
	data := c.get(ctx, u)
	if data == nil {
		return true
	}
	return data.TestAgent(u.Path, ua)
}

// get はホスト単位で robots.txt を取得・キャッシュし、パース結果を返す。
func (c *Cache) get(ctx context.Context, u *url.URL) *robotstxt.RobotsData {
	host := u.Scheme + "://" + u.Host
	c.mu.Lock()
	defer c.mu.Unlock()
	if d, ok := c.hosts[host]; ok {
		return d
	}

	robotsURL, err := url.Parse(host + "/robots.txt")
	if err != nil {
		c.hosts[host] = nil
		return nil
	}
	res, err := c.fetcher.Get(ctx, robotsURL, nil)
	if err != nil {
		slog.Warn("robots.txt fetch failed (treat as allow)", "host", host, "err", err.Error())
		c.hosts[host] = nil
		return nil
	}
	if res.StatusCode == 404 || res.StatusCode >= 500 {
		c.hosts[host] = nil
		return nil
	}
	data, err := robotstxt.FromBytes(res.Body)
	if err != nil {
		slog.Warn("robots.txt parse failed (treat as allow)", "host", host, "err", err.Error())
		c.hosts[host] = nil
		return nil
	}
	c.hosts[host] = data
	return data
}
