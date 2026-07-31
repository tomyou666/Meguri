package core

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
)

// TestCrawler_skipReason_hosts は include_hosts / exclude_hosts によるスキップ判定を検証する。
func TestCrawler_skipReason_hosts(t *testing.T) {
	mustURL := func(raw string) *url.URL {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		return u
	}

	newCrawler := func(include, exclude []string) *Crawler {
		cfg := model.Default()
		cfg.Targets = []string{"https://example.com/"}
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 5
		cfg.Crawl.IncludeHosts = include
		cfg.Crawl.ExcludeHosts = exclude
		cfg.Crawl.AllowExternal = true
		c := &Crawler{cfg: &cfg}
		if len(include) > 0 {
			c.includeHosts = make(map[string]struct{}, len(include))
			for _, h := range include {
				c.includeHosts[h] = struct{}{}
			}
		}
		if len(exclude) > 0 {
			c.excludeHosts = make(map[string]struct{}, len(exclude))
			for _, h := range exclude {
				c.excludeHosts[h] = struct{}{}
			}
		}
		return c
	}

	t.Run("正常系: exclude_hosts 一致は exclude_hosts でスキップ", func(t *testing.T) {
		c := newCrawler(nil, []string{"blocked.example"})
		got := c.skipReason(mustURL("https://blocked.example/a"), 0, nil)
		assert.Equal(t, "exclude_hosts", got)
	})

	t.Run("正常系: ポート違いでは exclude_hosts に一致しない", func(t *testing.T) {
		c := newCrawler(nil, []string{"example.com"})
		got := c.skipReason(mustURL("https://example.com:8080/a"), 0, nil)
		assert.Equal(t, "", got)
	})

	t.Run("正常系: include_hosts 非空で不一致は include_hosts でスキップ", func(t *testing.T) {
		c := newCrawler([]string{"allowed.example"}, nil)
		got := c.skipReason(mustURL("https://other.example/a"), 0, nil)
		assert.Equal(t, "include_hosts", got)
	})

	t.Run("正常系: include_hosts 空ならホスト理由ではスキップしない", func(t *testing.T) {
		c := newCrawler(nil, nil)
		got := c.skipReason(mustURL("https://any.example/a"), 0, nil)
		assert.Equal(t, "", got)
	})

	t.Run("正常系: include 許可かつ exclude 対象は exclude_hosts 優先", func(t *testing.T) {
		c := newCrawler([]string{"both.example"}, []string{"both.example"})
		got := c.skipReason(mustURL("https://both.example/a"), 0, nil)
		assert.Equal(t, "exclude_hosts", got)
	})
}
