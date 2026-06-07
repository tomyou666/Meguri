package domain

import (
	"net/url"
	"strings"
)

// NormalizeCrawlURL は backend crawler と同じ規則で URL を正規化する。
func NormalizeCrawlURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	cp := *u
	cp.Scheme = strings.ToLower(cp.Scheme)
	cp.Host = strings.ToLower(cp.Host)
	cp.Fragment = ""
	host := cp.Host
	switch {
	case cp.Scheme == "http" && strings.HasSuffix(host, ":80"):
		cp.Host = strings.TrimSuffix(host, ":80")
	case cp.Scheme == "https" && strings.HasSuffix(host, ":443"):
		cp.Host = strings.TrimSuffix(host, ":443")
	}
	return cp.String(), nil
}
