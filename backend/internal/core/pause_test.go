package core_test

import (
	"context"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/internal/core"
)

func TestPauseControllerWaitAndResume(t *testing.T) {
	p := core.NewPauseController()
	ctx := context.Background()

	p.Pause()
	done := make(chan struct{})
	go func() {
		err := p.WaitIfPaused(ctx)
		assert.NoError(t, err)
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("should block while paused")
	case <-time.After(50 * time.Millisecond):
	}

	p.Resume()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("should unblock after resume")
	}
}

func TestPauseControllerContextCancel(t *testing.T) {
	p := core.NewPauseController()
	ctx, cancel := context.WithCancel(context.Background())

	p.Pause()
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.WaitIfPaused(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("should return on ctx cancel")
	}
}

func TestCrawlerPauseBlocksNewFetch(t *testing.T) {
	srv := newTestWebServer(t)
	defer srv.Close()

	cfg := baseConfig()
	cfg.Crawl.Enabled = true
	cfg.Crawl.MaxDepth = 1
	cfg.Crawl.MaxPages = 10
	cfg.Crawl.MaxConcurrency = 2
	cfg.Crawl.RequestDelay = 50 * time.Millisecond

	pause := core.NewPauseController()
	var started atomic.Int32

	progress := func(ev core.ProgressEvent) {
		if ev.Kind == core.ProgressStarted {
			started.Add(1)
		}
	}

	k := setupKernel(t, cfg)
	c := core.NewCrawler(k, core.NewPipeline(k), nil, nil, progress)
	c.SetPauseController(pause)

	seed, err := url.Parse(srv.URL + "/links_with_pdf.html")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		_, _ = c.Run(ctx, []*url.URL{seed})
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	before := started.Load()
	pause.Pause()
	time.Sleep(200 * time.Millisecond)
	afterPause := started.Load()

	pause.Resume()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("crawl should complete")
	}

	assert.GreaterOrEqual(t, before, int32(1))
	assert.Equal(t, before, afterPause, "no new fetch should start while paused")
}
