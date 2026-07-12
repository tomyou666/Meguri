// Package robots は robots.txt のホスト単位キャッシュと許可判定を提供する。
package robots

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"sync"

	"github.com/temoto/robotstxt"
	"golang.org/x/sync/singleflight"

	"meguri/internal/domain/plugin"
)

// Cache はホスト単位で robots.txt を一度だけ取得・キャッシュする。
//
// キャッシュキーは host（非デフォルト port 含む）のみとし、http/https で共有する。
// 取得 URL の scheme は Allowed 呼び出し側 URL の scheme を使い、同一ホストの
// 同時ミスは singleflight で 1 回にまとめる（先に Do した呼び出しの scheme が勝つ）。
// scheme 違いで robots.txt 内容が異なる稀ケースは、先勝ちキャッシュを許容する。
// サブドメインは別ホストとして別キーのまま（統合しない）。
type Cache struct {
	// mu は hosts マップの排他制御（取得 I/O 中は保持しない）。
	mu sync.Mutex
	// hosts は host キー→パース済み robots データ（失敗時は nil エントリで許可扱いをキャッシュ）。
	hosts map[string]*robotstxt.RobotsData
	// flight は同一ホストの同時ミスを 1 回の fetch にまとめる。
	flight singleflight.Group
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
// ua は robots.txt 取得時の User-Agent ヘッダにも使う（ページ取得と揃える）。
func (c *Cache) Allowed(ctx context.Context, u *url.URL, ua string) bool {
	agent := ua
	if agent == "" {
		agent = "*"
	}
	data := c.get(ctx, u, ua)
	if data == nil {
		return true
	}
	return data.TestAgent(u.Path, agent)
}

// cacheKey は robots キャッシュのキーを返す（host。非デフォルト port 含む）。
func cacheKey(u *url.URL) string {
	return u.Host
}

// get はホスト単位で robots.txt を取得・キャッシュし、パース結果を返す。
// requestUA が空でない場合は User-Agent ヘッダとして付与する。
func (c *Cache) get(ctx context.Context, u *url.URL, requestUA string) *robotstxt.RobotsData {
	key := cacheKey(u)
	c.mu.Lock()
	if d, ok := c.hosts[key]; ok {
		c.mu.Unlock()
		return d
	}
	c.mu.Unlock()

	scheme := u.Scheme
	v, _, _ := c.flight.Do(key, func() (interface{}, error) {
		c.mu.Lock()
		if d, ok := c.hosts[key]; ok {
			c.mu.Unlock()
			return d, nil
		}
		c.mu.Unlock()

		data, cacheable := c.fetchAndParse(ctx, scheme, key, requestUA)
		if cacheable {
			c.mu.Lock()
			c.hosts[key] = data
			c.mu.Unlock()
		}
		return data, nil
	})
	if v == nil {
		return nil
	}
	return v.(*robotstxt.RobotsData)
}

// fetchAndParse は scheme://host/robots.txt を取得してパースする。
// 戻り値の bool はキャッシュしてよいか。context.Canceled 時のみ false（未取得のまま）。
// その他の失敗は nil + true（許可扱いでキャッシュ）。
func (c *Cache) fetchAndParse(ctx context.Context, scheme, host, requestUA string) (*robotstxt.RobotsData, bool) {
	robotsURL, err := url.Parse(scheme + "://" + host + "/robots.txt")
	if err != nil {
		return nil, true
	}
	var headers map[string]string
	if ua := strings.TrimSpace(requestUA); ua != "" {
		headers = map[string]string{"User-Agent": ua}
	}
	res, err := c.fetcher.Get(ctx, robotsURL, headers)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			slog.Warn("robots.txt fetch canceled (not cached)", "host", host, "err", err.Error())
			return nil, false
		}
		slog.Warn("robots.txt fetch failed (treat as allow)", "host", host, "err", err.Error())
		return nil, true
	}
	if res.StatusCode == 404 || res.StatusCode >= 500 {
		return nil, true
	}
	data, err := robotstxt.FromBytes(res.Body)
	if err != nil {
		slog.Warn("robots.txt parse failed (treat as allow)", "host", host, "err", err.Error())
		return nil, true
	}
	return data, true
}
