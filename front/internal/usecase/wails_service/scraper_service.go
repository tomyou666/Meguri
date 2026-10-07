package wails_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"meguri-app/internal/domain"
	"meguri-app/internal/model"
	"meguri/pkg/runner"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	topicRunStarted     = "scraper:crawl:runStarted"
	topicNodeStarted    = "scraper:crawl:nodeStarted"
	topicNodeSucceeded  = "scraper:crawl:nodeSucceeded"
	topicNodeFailed     = "scraper:crawl:nodeFailed"
	topicNodeSkipped    = "scraper:crawl:nodeSkipped"
	topicLinkSkipped    = "scraper:crawl:linkSkipped"
	topicEdgeDiscovered = "scraper:crawl:edgeDiscovered"
	topicCrawlCompleted = "scraper:crawl:completed"
	topicCrawlError     = "scraper:crawl:error"
)

const crawlShutdownTimeout = 5 * time.Second

//go:generate go tool gowrap gen -p meguri-app/internal/usecase/wails_service -i ScraperServiceAPI -t templates/slog_debug_slow.gotmpl -o scraper_service_with_debug_log.go -v ServiceName=ScraperService -v DecoratorName=ScraperServiceWithDebugLog

// ScraperServiceAPI は ScraperService の公開メソッド面。
type ScraperServiceAPI interface {
	// SetApp は Wails App を後から注入する（Event 発火用）。
	SetApp(app *application.App)
	// ServiceShutdown は Wails アプリ終了時に active crawl を止め chromium プールを閉じる。
	ServiceShutdown() error
	// StartCrawl はクロールを非同期で開始し runId を返す。
	StartCrawl(req model.StartCrawlRequest) (string, error)
	// PauseCrawl は実行中 crawl を一時停止する。
	PauseCrawl(runID string) error
	// ResumeCrawl は一時停止中 crawl を再開する。
	ResumeCrawl(runID string) error
	// StopCrawl は実行中 crawl をキャンセルする。
	StopCrawl(runID string) error
	// FetchRobotsTxt は robots.txt を取得する。
	FetchRobotsTxt(host, baseURL string, appDefaults, wsSettings json.RawMessage) (model.RobotsTxtInfoDTO, error)
}

// ScraperService は backend crawler を駆動し Wails Event で進捗を配信する（debug ログ付き殻）。
type ScraperService struct {
	ScraperServiceWithDebugLog
}

// scraperService は ScraperService の実装本体。
type scraperService struct {
	// app は Wails アプリ（Event 発火用）。
	app *application.App
	// persist はクロール永続化サービス。
	persist *domain.CrawlPersistService
	// mu は active crawl job の排他制御。
	mu sync.Mutex
	// job は実行中クロールジョブ。
	job *activeCrawlJob
}

type activeCrawlJob struct {
	// runID はクロール実行 ID。
	runID string
	// pause は一時停止コントローラ。
	pause *runner.PauseController
	// cache は runner キャッシュ。
	cache *runner.RunnerCache
	// opts は runner 実行オプション。
	opts *runner.RunOptions
	// cancel はクロール context の cancel。
	cancel context.CancelFunc
	// done はクロール goroutine 終了を通知する（close でシグナル）。
	done chan struct{}
}

// NewScraperService は ScraperService を構築する。
func NewScraperService(persist *domain.CrawlPersistService) *ScraperService {
	impl := &scraperService{persist: persist}
	return &ScraperService{ScraperServiceWithDebugLog: NewScraperServiceWithDebugLog(impl)}
}

// SetApp は Wails App を後から注入する（Event 発火用）。
func (s *scraperService) SetApp(app *application.App) {
	s.app = app
}

// ServiceShutdown は Wails アプリ終了時に active crawl を止め chromium プールを閉じる。
func (s *scraperService) ServiceShutdown() error {
	s.mu.Lock()
	var done <-chan struct{}
	if s.job != nil {
		done = s.job.done
		s.job.cancel()
	}
	s.mu.Unlock()

	if done != nil {
		select {
		case <-done:
		case <-time.After(crawlShutdownTimeout):
			slog.Warn("scraper crawl shutdown timed out")
			s.releaseActiveJobResources()
		}
	}

	runner.CloseChromiumBrowsers()
	return nil
}

// StartCrawl はクロールを非同期で開始し runId を返す。
func (s *scraperService) StartCrawl(req model.StartCrawlRequest) (string, error) {
	if s.app == nil {
		return "", fmt.Errorf("app not initialized")
	}
	if req.WorkspaceID == "" {
		return "", fmt.Errorf("workspaceId is required")
	}
	if s.persist == nil {
		return "", fmt.Errorf("persist service not initialized")
	}

	runID := domain.NewRunID()
	req.RunID = runID
	startedAt := domain.NowISO()

	if err := s.persist.BeginCrawlRun(context.Background(), model.BeginCrawlRunRequest{
		WorkspaceID: req.WorkspaceID,
		RunID:       runID,
		Mode:        req.Mode,
		StartedAt:   startedAt,
	}); err != nil {
		return "", err
	}

	s.mu.Lock()
	if s.job != nil {
		s.mu.Unlock()
		return "", fmt.Errorf("crawl already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	pause := runner.NewPauseController()
	cache := runner.NewRunnerCache()
	done := make(chan struct{})
	s.job = &activeCrawlJob{
		runID:  runID,
		pause:  pause,
		cache:  cache,
		opts:   &runner.RunOptions{Pause: pause, Cache: cache},
		cancel: cancel,
		done:   done,
	}
	s.mu.Unlock()

	s.emit(topicRunStarted, model.CrawlEventPayload{
		WorkspaceID: req.WorkspaceID,
		RunID:       runID,
	})

	go func() {
		defer close(done)
		defer func() {
			s.releaseActiveJobResources()
		}()
		if err := s.runCrawl(ctx, req); err != nil {
			s.finishCrawlRun(ctx, req, "error", nil, err.Error())
			s.emit(topicCrawlError, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       runID,
				Message:     err.Error(),
			})
		}
	}()
	return runID, nil
}

// PauseCrawl は実行中 crawl を一時停止する。
func (s *scraperService) PauseCrawl(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil || s.job.runID != runID {
		return fmt.Errorf("no active crawl for runId %s", runID)
	}
	s.job.pause.Pause()
	return nil
}

// ResumeCrawl は一時停止中 crawl を再開する。
func (s *scraperService) ResumeCrawl(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil || s.job.runID != runID {
		return fmt.Errorf("no active crawl for runId %s", runID)
	}
	s.job.pause.Resume()
	return nil
}

// StopCrawl は実行中 crawl をキャンセルする。
func (s *scraperService) StopCrawl(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil || s.job.runID != runID {
		return nil
	}
	s.job.cancel()
	return nil
}

// releaseActiveJobResources は active job の RunnerCache と FetchLimiter を解放する。
func (s *scraperService) releaseActiveJobResources() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil {
		return
	}
	if s.job.cache != nil {
		s.job.cache.CloseAll()
	}
	if s.job.opts != nil && s.job.opts.FetchLimiter != nil {
		s.job.opts.FetchLimiter.Close()
	}
	s.job = nil
}

func (s *scraperService) runOptions() *runner.RunOptions {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil {
		return nil
	}
	return s.job.opts
}

func (s *scraperService) emit(topic string, payload model.CrawlEventPayload) {
	if s.app == nil {
		return
	}
	s.app.Event.Emit(topic, payload)
}

func (s *scraperService) finishCrawlRun(
	ctx context.Context,
	req model.StartCrawlRequest,
	status string,
	summary *model.CrawlSummaryDTO,
	errMsg string,
) {
	if s.persist == nil {
		return
	}
	var summaryJSON json.RawMessage
	if summary != nil {
		if b, err := json.Marshal(summary); err == nil {
			summaryJSON = b
		}
	}
	_ = s.persist.FinishCrawlRun(ctx, model.FinishCrawlRunRequest{
		WorkspaceID:  req.WorkspaceID,
		RunID:        req.RunID,
		Status:       status,
		FinishedAt:   domain.NowISO(),
		SummaryJSON:  summaryJSON,
		ErrorMessage: errMsg,
	})
}

func (s *scraperService) persistNodeStarted(ctx context.Context, req model.StartCrawlRequest, nodeID string) {
	if s.persist == nil || nodeID == "" {
		return
	}
	err := s.persist.PatchGraphNodeStatus(ctx, model.PatchGraphNodeStatusRequest{
		WorkspaceID: req.WorkspaceID,
		NodeID:      nodeID,
		Status:      "running",
	})
	s.logPersistError(ctx, "patchGraphNodeStatus", req, nodeID, "", err)
}

func (s *scraperService) logPersistError(
	ctx context.Context,
	op string,
	req model.StartCrawlRequest,
	nodeID, url string,
	err error,
) {
	if err == nil {
		return
	}
	slog.ErrorContext(ctx, "crawl persist failed",
		"op", op,
		"workspaceId", req.WorkspaceID,
		"runId", req.RunID,
		"nodeId", nodeID,
		"url", url,
		"err", err,
	)
}

func (s *scraperService) persistNodeSucceeded(
	ctx context.Context,
	req model.StartCrawlRequest,
	nodeID, url string,
	result *model.CrawlNodeResultDTO,
) {
	if s.persist == nil || nodeID == "" || result == nil {
		return
	}
	markdown := result.Markdown
	contentHash := ""
	if markdown != "" {
		contentHash = domain.ContentHashFromMarkdown(markdown)
	}
	var linksJSON, metadataJSON string
	if len(result.Links) > 0 {
		if b, err := json.Marshal(result.Links); err == nil {
			linksJSON = string(b)
		}
	}
	if len(result.Metadata) > 0 {
		if b, err := json.Marshal(result.Metadata); err == nil {
			metadataJSON = string(b)
		}
	}
	linksHash := domain.LinksHashFromLinks(result.Links)
	err := s.persist.AppendNodeResult(ctx, model.AppendNodeResultRequest{
		WorkspaceID:  req.WorkspaceID,
		RunID:        req.RunID,
		NodeID:       nodeID,
		URL:          url,
		Markdown:     markdown,
		HTML:         result.HTML,
		RawHTML:      result.RawHTML,
		LinksJSON:    linksJSON,
		MetadataJSON: metadataJSON,
		FetchedAt:    domain.NowISO(),
		ContentHash:  contentHash,
		LinksHash:    linksHash,
	})
	s.logPersistError(ctx, "appendNodeResult", req, nodeID, url, err)
	err = s.persist.PatchGraphNodeStatus(ctx, model.PatchGraphNodeStatusRequest{
		WorkspaceID: req.WorkspaceID,
		NodeID:      nodeID,
		Status:      "success",
	})
	s.logPersistError(ctx, "patchGraphNodeStatus", req, nodeID, url, err)
}

func (s *scraperService) persistNodeFailed(
	ctx context.Context,
	req model.StartCrawlRequest,
	nodeID, url, errMsg string,
) {
	if s.persist == nil || nodeID == "" {
		return
	}
	err := s.persist.AppendNodeResult(ctx, model.AppendNodeResultRequest{
		WorkspaceID: req.WorkspaceID,
		RunID:       req.RunID,
		NodeID:      nodeID,
		URL:         url,
		Error:       errMsg,
		FetchedAt:   domain.NowISO(),
	})
	s.logPersistError(ctx, "appendNodeResult", req, nodeID, url, err)
	err = s.persist.PatchGraphNodeStatus(ctx, model.PatchGraphNodeStatusRequest{
		WorkspaceID: req.WorkspaceID,
		NodeID:      nodeID,
		Status:      "error",
		LastError:   errMsg,
	})
	s.logPersistError(ctx, "patchGraphNodeStatus", req, nodeID, url, err)
}

func (s *scraperService) persistNodeSkipped(ctx context.Context, req model.StartCrawlRequest, nodeID string) {
	if s.persist == nil || nodeID == "" {
		return
	}
	err := s.persist.PatchGraphNodeStatus(ctx, model.PatchGraphNodeStatusRequest{
		WorkspaceID: req.WorkspaceID,
		NodeID:      nodeID,
		Status:      "skipped",
	})
	s.logPersistError(ctx, "patchGraphNodeStatus", req, nodeID, "", err)
}

func (s *scraperService) persistEdgeDiscovered(
	ctx context.Context,
	req model.StartCrawlRequest,
	sourceID, targetID, targetURL string,
) {
	if s.persist == nil {
		return
	}
	_ = s.persist.UpsertDiscoveredGraph(ctx, model.UpsertDiscoveredGraphRequest{
		WorkspaceID: req.WorkspaceID,
		SourceID:    sourceID,
		TargetID:    targetID,
		TargetURL:   targetURL,
	})
}

func (s *scraperService) runCrawl(ctx context.Context, req model.StartCrawlRequest) error {
	ws := req.Workspace
	state := newCrawlState(req)

	opts := s.runOptions()
	if opts != nil {
		baseCfg, err := state.limiterConfig()
		if err != nil {
			return err
		}
		lim := runner.PrepareFetchLimiter(ctx, baseCfg, opts)
		if opts.Cache != nil {
			opts.Cache.SetFetchLimiter(lim)
		}
	}

	var (
		enqueued, succeeded, failed, skipped int
		stoppedReason                        = "completed"
	)

	emitSummary := func() {
		status := "completed"
		if stoppedReason == "stopped" {
			status = "stopped"
		}
		summary := &model.CrawlSummaryDTO{
			Mode:                  req.Mode,
			FinishedAt:            time.Now().UTC().Format(time.RFC3339),
			Enqueued:              enqueued,
			Succeeded:             succeeded,
			Failed:                failed,
			Skipped:               skipped,
			SkippedDuplicateLinks: state.linkSkippedCount,
			StoppedReason:         stoppedReason,
		}
		s.finishCrawlRun(ctx, req, status, summary, "")
		s.emit(topicCrawlCompleted, model.CrawlEventPayload{
			WorkspaceID: req.WorkspaceID,
			RunID:       req.RunID,
			Summary:     summary,
		})
	}

	switch req.Mode {
	case 1, 2:
		stats, mainReached, err := s.runMainBFS(ctx, req, state, opts, &enqueued, &succeeded, &failed, &skipped)
		if err != nil {
			if ctx.Err() != nil {
				stoppedReason = "stopped"
				emitSummary()
				return nil
			}
			return err
		}
		if stats != nil {
			enqueued = stats.Enqueued
			succeeded = stats.Succeeded
			failed = stats.Failed
			skipped = stats.Skipped
		}
		if err := s.runManualPass(ctx, req, state, mainReached, opts, &enqueued, &succeeded, &failed, &skipped); err != nil {
			if ctx.Err() != nil {
				stoppedReason = "stopped"
				emitSummary()
				return nil
			}
			return err
		}
	case 3:
		if err := s.runMode3(ctx, req, state, opts, &enqueued, &succeeded, &failed, &skipped); err != nil {
			if ctx.Err() != nil {
				stoppedReason = "stopped"
				emitSummary()
				return nil
			}
			return err
		}
		mainReached := map[string]struct{}{}
		for _, n := range ws.Nodes {
			mainReached[n.ID] = struct{}{}
		}
		if err := s.runManualPass(ctx, req, state, mainReached, opts, &enqueued, &succeeded, &failed, &skipped); err != nil {
			if ctx.Err() != nil {
				stoppedReason = "stopped"
				emitSummary()
				return nil
			}
			return err
		}
	case 4:
		if err := s.runMode4(ctx, req, state, opts, &enqueued, &succeeded, &failed, &skipped); err != nil {
			if ctx.Err() != nil {
				stoppedReason = "stopped"
				emitSummary()
				return nil
			}
			return err
		}
	default:
		return fmt.Errorf("unsupported mode %d", req.Mode)
	}

	if ctx.Err() != nil {
		stoppedReason = "stopped"
	}
	emitSummary()
	return nil
}

type crawlState struct {
	mu               sync.Mutex
	nextNodeSeq      int64
	urlToNode        map[string]string
	nodeByID         map[string]model.GraphNodeDTO
	initialURLs      map[string]struct{}
	materializedURLs map[string]struct{}
	excludeSet       map[string]struct{}
	outEdges         map[string][]string
	appDefaults      json.RawMessage
	wsSettings       json.RawMessage
	rescrapeExisting bool
	linkSkippedCount int
}

func newCrawlState(req model.StartCrawlRequest) *crawlState {
	ws := req.Workspace
	st := &crawlState{
		urlToNode:        map[string]string{},
		nodeByID:         map[string]model.GraphNodeDTO{},
		initialURLs:      map[string]struct{}{},
		materializedURLs: map[string]struct{}{},
		excludeSet:       map[string]struct{}{},
		outEdges:         map[string][]string{},
		appDefaults:      req.AppDefaults,
		wsSettings:       ws.Settings,
		rescrapeExisting: req.RescrapeExisting,
	}
	for _, u := range ws.ExcludeURLs {
		st.excludeSet[u] = struct{}{}
	}
	for _, n := range ws.Nodes {
		key := n.URLNormalized
		if norm, err := domain.NormalizeCrawlURL(n.URLNormalized); err == nil {
			key = norm
		}
		st.urlToNode[key] = n.ID
		st.nodeByID[n.ID] = n
		st.initialURLs[key] = struct{}{}
		st.materializedURLs[key] = struct{}{}
	}
	for _, e := range ws.Edges {
		st.outEdges[e.Source] = append(st.outEdges[e.Source], e.Target)
	}
	return st
}

func (st *crawlState) crawlURLKey(rawURL string) string {
	if norm, err := domain.NormalizeCrawlURL(rawURL); err == nil {
		return norm
	}
	return rawURL
}

func (st *crawlState) resolveNodeID(rawURL string, create bool) (nodeID, urlKey string) {
	key := st.crawlURLKey(rawURL)
	st.mu.Lock()
	defer st.mu.Unlock()
	id, _ := st.nodeIDForURLLocked(key, create)
	return id, key
}

func (st *crawlState) nodeIDForURLLocked(key string, create bool) (string, bool) {
	if id, ok := st.urlToNode[key]; ok {
		return id, true
	}
	if !create {
		return "", false
	}
	st.nextNodeSeq++
	id := fmt.Sprintf("n-%d-%d", time.Now().UnixMilli(), st.nextNodeSeq)
	st.urlToNode[key] = id
	st.nodeByID[id] = model.GraphNodeDTO{
		ID:            id,
		URLNormalized: key,
		Label:         key,
		Origin:        "crawl",
		Status:        "idle",
	}
	return id, false
}

// skipScrapeURLs は再取得トグル OFF 時に fetch しない success ノード URL 一覧を返す。
func (st *crawlState) skipScrapeURLs() []string {
	if st.rescrapeExisting {
		return nil
	}
	urls := make([]string, 0)
	for _, n := range st.nodeByID {
		if n.Status == "success" {
			urls = append(urls, n.URLNormalized)
		}
	}
	return urls
}

// mergeSkipScrapeLinkMap は DB マップに LastResult.Links を不足分だけ足す。
func (st *crawlState) mergeSkipScrapeLinkMap(fromDB map[string][]string, skipURLs []string) map[string][]string {
	if len(skipURLs) == 0 {
		return nil
	}
	want := make(map[string]struct{}, len(skipURLs))
	for _, u := range skipURLs {
		want[u] = struct{}{}
	}
	out := make(map[string][]string)
	for k, v := range fromDB {
		if _, ok := want[k]; !ok {
			continue
		}
		if len(v) == 0 {
			continue
		}
		out[k] = append([]string(nil), v...)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// linkSkipReason は子 URL が素材化済みのときスキップ理由を返す。未登録なら空文字。
func (st *crawlState) linkSkipReason(childKey string) string {
	if _, ok := st.initialURLs[childKey]; ok {
		return "duplicate_existing"
	}
	if _, ok := st.materializedURLs[childKey]; ok {
		return "duplicate_in_run"
	}
	return ""
}

func (st *crawlState) markMaterialized(childKey string) {
	st.materializedURLs[childKey] = struct{}{}
}

// shouldSuppressNodeSkipped は今回 run で到達済み（Started/Succeeded）のノードに nodeSkipped を出さない。
func shouldSuppressNodeSkipped(mainReached map[string]struct{}, nodeID string) bool {
	_, ok := mainReached[nodeID]
	return ok
}

// noteLinkSkipped はリンクスキップ件数を加算する（UI へは出さない）。
func (s *scraperService) noteLinkSkipped(st *crawlState) {
	st.mu.Lock()
	st.linkSkippedCount++
	st.mu.Unlock()
}

// emitLinkSkipped はリンクスキップを集計し UI へ通知する。
func (s *scraperService) emitLinkSkipped(
	req model.StartCrawlRequest,
	st *crawlState,
	parentURL, childURL, reason string,
) {
	s.noteLinkSkipped(st)
	s.emit(topicLinkSkipped, model.CrawlEventPayload{
		WorkspaceID: req.WorkspaceID,
		RunID:       req.RunID,
		URL:         parentURL,
		TargetURL:   childURL,
		Reason:      reason,
	})
}

// limiterConfig は fetch limiter 用に app → workspace をマージする。ノード層は含めない。
func (st *crawlState) limiterConfig() (*runner.Config, error) {
	merged, err := runner.MergeUIConfigLayers(st.appDefaults, st.wsSettings)
	if err != nil {
		return nil, err
	}
	return runner.ParseUIConfig(merged)
}

// runConfig は探索全体の設定。app → workspace に、ノード content は載せない。
func (st *crawlState) runConfig() (*runner.Config, error) {
	return st.mergedConfig(model.GraphNodeDTO{})
}

// pageConfig は既存ノードに content 上書きがあるときだけ、その URL の設定を返す。
// 未登録 URL と content のないノードは nil（ラン共通パイプライン）。
func (st *crawlState) pageConfig(rawURL string) (*runner.Config, error) {
	key := st.crawlURLKey(rawURL)
	st.mu.Lock()
	id, ok := st.urlToNode[key]
	var node model.GraphNodeDTO
	if ok {
		node = st.nodeByID[id]
	}
	st.mu.Unlock()
	if !ok {
		return nil, nil
	}
	filtered, err := runner.FilterNodeUIConfigLayer(node.NodeSettings)
	if err != nil {
		return nil, err
	}
	if len(filtered) == 0 || string(filtered) == "{}" || string(filtered) == "null" {
		return nil, nil
	}
	return st.mergedConfig(node)
}

// mergedConfig は app → workspace → node content をマージする。モードでは分岐しない。
func (st *crawlState) mergedConfig(node model.GraphNodeDTO) (*runner.Config, error) {
	nodeLayer, err := runner.FilterNodeUIConfigLayer(node.NodeSettings)
	if err != nil {
		return nil, err
	}
	layers := []json.RawMessage{st.appDefaults, st.wsSettings, nodeLayer}
	merged, err := runner.MergeUIConfigLayers(layers...)
	if err != nil {
		return nil, err
	}
	cfg, err := runner.ParseUIConfig(merged)
	if err != nil {
		return nil, err
	}
	runner.DeriveContentFormats(cfg)
	exclude := make([]string, 0, len(st.excludeSet))
	for u := range st.excludeSet {
		exclude = append(exclude, u)
	}
	cfg.Crawl.ExcludeURLs = exclude
	cfg.Targets = []string{node.URLNormalized}
	return cfg, nil
}

func (s *scraperService) runMainBFS(
	ctx context.Context,
	req model.StartCrawlRequest,
	st *crawlState,
	opts *runner.RunOptions,
	enqueued, succeeded, failed, skipped *int,
) (*runner.CrawlStats, map[string]struct{}, error) {
	ws := req.Workspace
	mainReached := map[string]struct{}{}

	var seedURL string
	if req.Mode == 2 {
		if req.StartNodeID == "" {
			return nil, mainReached, fmt.Errorf("mode 2 requires startNodeId")
		}
		node, ok := st.nodeByID[req.StartNodeID]
		if !ok {
			return nil, mainReached, fmt.Errorf("start node not found")
		}
		seedURL = node.URLNormalized
	} else {
		seedURL = ws.SeedURL
	}

	_, ok := st.nodeByID[st.urlToNode[seedURL]]
	if !ok {
		for _, n := range ws.Nodes {
			if n.URLNormalized == seedURL {
				ok = true
				break
			}
		}
	}
	if !ok {
		return nil, mainReached, fmt.Errorf("seed node not found for %s", seedURL)
	}

	cfg, err := st.runConfig()
	if err != nil {
		return nil, mainReached, err
	}
	cfg.Targets = []string{seedURL}
	cfg.Crawl.Enabled = true
	skipURLs := st.skipScrapeURLs()
	cfg.Crawl.SkipScrapeURLs = skipURLs
	if len(skipURLs) > 0 {
		var fromDB map[string][]string
		if s.persist != nil {
			m, err := s.persist.BuildSkipScrapeLinkMap(ctx, req.WorkspaceID, skipURLs)
			if err != nil {
				slog.WarnContext(ctx, "skip scrape link map load failed",
					"workspaceId", req.WorkspaceID,
					"err", err.Error(),
				)
			} else {
				fromDB = m
			}
		}
		cfg.Crawl.SkipScrapeLinkMap = st.mergeSkipScrapeLinkMap(fromDB, skipURLs)
	}

	if opts != nil {
		prev := opts.PageConfig
		opts.PageConfig = st.pageConfig
		defer func() { opts.PageConfig = prev }()
	}

	progress := func(ev runner.ProgressEvent) {
		switch ev.Kind {
		case runner.ProgressLinkDiscovered:
			st.mu.Lock()
			parentKey := st.crawlURLKey(ev.ParentURL)
			parentID, parentOK := st.nodeIDForURLLocked(parentKey, false)
			if !parentOK {
				st.mu.Unlock()
				return
			}
			childKey := st.crawlURLKey(ev.URL)
			if reason := st.linkSkipReason(childKey); reason != "" {
				st.mu.Unlock()
				// 既存グラフ上の再発見は集計のみ（UI 連打を避ける）。
				if reason == "duplicate_existing" {
					s.noteLinkSkipped(st)
					return
				}
				s.emitLinkSkipped(req, st, parentKey, childKey, reason)
				return
			}
			childID, _ := st.nodeIDForURLLocked(childKey, true)
			st.markMaterialized(childKey)
			st.mu.Unlock()
			s.persistEdgeDiscovered(ctx, req, parentID, childID, childKey)
			s.emit(topicEdgeDiscovered, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				SourceID:    parentID,
				TargetID:    childID,
				TargetURL:   childKey,
			})
		case runner.ProgressStarted:
			st.mu.Lock()
			urlKey := st.crawlURLKey(ev.URL)
			nodeID, _ := st.nodeIDForURLLocked(urlKey, true)
			if nodeID != "" {
				mainReached[nodeID] = struct{}{}
			}
			st.mu.Unlock()
			if nodeID == "" {
				return
			}
			s.persistNodeStarted(ctx, req, nodeID)
			s.emit(topicNodeStarted, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      nodeID,
				URL:         urlKey,
			})
		case runner.ProgressSucceeded:
			st.mu.Lock()
			urlKey := st.crawlURLKey(ev.URL)
			nodeID, _ := st.nodeIDForURLLocked(urlKey, false)
			if nodeID != "" {
				mainReached[nodeID] = struct{}{}
			}
			st.mu.Unlock()
			if nodeID == "" {
				return
			}
			dto := resultToDTO(ev.Result)
			s.persistNodeSucceeded(ctx, req, nodeID, urlKey, dto)
			s.emit(topicNodeSucceeded, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      nodeID,
				URL:         urlKey,
			})
		case runner.ProgressFailed:
			nodeID, urlKey := st.resolveNodeID(ev.URL, false)
			if nodeID == "" {
				return
			}
			s.persistNodeFailed(ctx, req, nodeID, urlKey, ev.Error)
			s.emit(topicNodeFailed, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      nodeID,
				URL:         urlKey,
				Error:       ev.Error,
			})
		case runner.ProgressSkipped:
			urlKey := st.crawlURLKey(ev.URL)
			switch ev.SkipReason {
			case "duplicate":
				s.emitLinkSkipped(req, st, ev.ParentURL, urlKey, "duplicate_in_run")
				return
			case "already_success":
				// 再取得 OFF の既存 success は集計のみ（UI 連打を避ける）。
				s.noteLinkSkipped(st)
				return
			}
			nodeID, _ := st.resolveNodeID(ev.URL, false)
			if nodeID == "" {
				return
			}
			if shouldSuppressNodeSkipped(mainReached, nodeID) {
				s.emitLinkSkipped(req, st, ev.ParentURL, urlKey, "duplicate_in_run")
				return
			}
			s.persistNodeSkipped(ctx, req, nodeID)
			s.emit(topicNodeSkipped, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      nodeID,
				URL:         urlKey,
				Reason:      ev.SkipReason,
			})
		}
	}

	stats, err := runner.CrawlWithProgress(ctx, cfg, []string{seedURL}, progress, opts)
	return stats, mainReached, err
}

func (s *scraperService) runMode3(
	ctx context.Context,
	req model.StartCrawlRequest,
	st *crawlState,
	opts *runner.RunOptions,
	enqueued, succeeded, failed, skipped *int,
) error {
	if req.StartNodeID == "" {
		return fmt.Errorf("mode 3 requires startNodeId")
	}
	nodeIDs := map[string]struct{}{}
	for id := range st.nodeByID {
		nodeIDs[id] = struct{}{}
	}
	order := domain.ForwardReachableExisting(req.StartNodeID, nodeIDs, st.outEdges)
	visit := append([]string{req.StartNodeID}, order...)
	return s.scrapeExistingNodesInOrder(ctx, req, st, visit, opts, enqueued, succeeded, failed, skipped)
}

// runMode4 は明示 nodeIds の既存ノードのみを入力順に scrape する（リンク探索なし）。
func (s *scraperService) runMode4(
	ctx context.Context,
	req model.StartCrawlRequest,
	st *crawlState,
	opts *runner.RunOptions,
	enqueued, succeeded, failed, skipped *int,
) error {
	if len(req.NodeIDs) == 0 {
		return fmt.Errorf("mode 4 requires nodeIds")
	}
	visit := filterExistingNodeIDs(req.NodeIDs, st.nodeByID)
	return s.scrapeExistingNodesInOrder(ctx, req, st, visit, opts, enqueued, succeeded, failed, skipped)
}

// filterExistingNodeIDs は nodeByID に存在する ID のみを入力順で返す。
func filterExistingNodeIDs(nodeIDs []string, nodeByID map[string]model.GraphNodeDTO) []string {
	visit := make([]string, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		if _, ok := nodeByID[nodeID]; ok {
			visit = append(visit, nodeID)
		}
	}
	return visit
}

// scrapeExistingNodesInOrder は visit 順に既存ノードを scrape する（mode 3 / 4 共通）。
func (s *scraperService) scrapeExistingNodesInOrder(
	ctx context.Context,
	req model.StartCrawlRequest,
	st *crawlState,
	visit []string,
	opts *runner.RunOptions,
	enqueued, succeeded, failed, skipped *int,
) error {
	for _, nodeID := range visit {
		node, ok := st.nodeByID[nodeID]
		if !ok {
			continue
		}
		if !st.rescrapeExisting && node.Status == "success" {
			s.emitLinkSkipped(req, st, "", node.URLNormalized, "duplicate_existing")
			continue
		}
		if _, ex := st.excludeSet[node.URLNormalized]; ex || node.CrawlExclude {
			*skipped++
			s.persistNodeSkipped(ctx, req, nodeID)
			s.emit(topicNodeSkipped, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      nodeID,
				URL:         node.URLNormalized,
				Reason:      "exclude_urls",
			})
			continue
		}
		*enqueued++
		if err := s.scrapeOneNode(ctx, req, st, node, opts); err != nil {
			if ctx.Err() != nil {
				return err
			}
			*failed++
		} else {
			*succeeded++
		}
	}
	return nil
}

func (s *scraperService) runManualPass(
	ctx context.Context,
	req model.StartCrawlRequest,
	st *crawlState,
	mainReached map[string]struct{},
	opts *runner.RunOptions,
	enqueued, succeeded, failed, skipped *int,
) error {
	for _, node := range st.nodeByID {
		if node.Origin != "manual" {
			continue
		}
		if _, reached := mainReached[node.ID]; reached {
			continue
		}
		if !st.rescrapeExisting && node.Status == "success" {
			s.emitLinkSkipped(req, st, "", node.URLNormalized, "duplicate_existing")
			continue
		}
		if _, ex := st.excludeSet[node.URLNormalized]; ex || node.CrawlExclude {
			*skipped++
			s.persistNodeSkipped(ctx, req, node.ID)
			s.emit(topicNodeSkipped, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      node.ID,
				URL:         node.URLNormalized,
				Reason:      "exclude_urls",
			})
			continue
		}
		*enqueued++
		if err := s.scrapeOneNode(ctx, req, st, node, opts); err != nil {
			if ctx.Err() != nil {
				return err
			}
			*failed++
		} else {
			*succeeded++
		}
	}
	return nil
}

func (s *scraperService) scrapeOneNode(
	ctx context.Context,
	req model.StartCrawlRequest,
	st *crawlState,
	node model.GraphNodeDTO,
	opts *runner.RunOptions,
) error {
	cfg, err := st.mergedConfig(node)
	if err != nil {
		return err
	}
	cfg.Crawl.Enabled = false

	var failed bool
	progress := func(ev runner.ProgressEvent) {
		switch ev.Kind {
		case runner.ProgressStarted:
			s.persistNodeStarted(ctx, req, node.ID)
			s.emit(topicNodeStarted, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      node.ID,
				URL:         ev.URL,
			})
		case runner.ProgressSucceeded:
			dto := resultToDTO(ev.Result)
			s.persistNodeSucceeded(ctx, req, node.ID, ev.URL, dto)
			s.emit(topicNodeSucceeded, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      node.ID,
				URL:         ev.URL,
			})
		case runner.ProgressFailed:
			failed = true
			s.persistNodeFailed(ctx, req, node.ID, ev.URL, ev.Error)
			s.emit(topicNodeFailed, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      node.ID,
				URL:         ev.URL,
				Error:       ev.Error,
			})
		}
	}

	_, err = runner.ScrapeWithConfig(ctx, node.URLNormalized, cfg, progress, opts)
	if err != nil {
		return err
	}
	if failed {
		return fmt.Errorf("scrape failed for %s", node.URLNormalized)
	}
	return nil
}

func resultToDTO(res *runner.Result) *model.CrawlNodeResultDTO {
	if res == nil || res.URL == nil {
		return nil
	}
	dto := &model.CrawlNodeResultDTO{URL: res.URL.String()}
	if res.Markdown != "" {
		dto.Markdown = res.Markdown
	}
	if res.HTML != "" {
		dto.HTML = res.HTML
	}
	if res.RawHTML != "" {
		dto.RawHTML = res.RawHTML
	}
	if len(res.Links) > 0 {
		dto.Links = make([]string, len(res.Links))
		for i, u := range res.Links {
			dto.Links[i] = u.String()
		}
	}
	if len(res.Metadata) > 0 {
		dto.Metadata = res.Metadata
	}
	return dto
}

// FetchRobotsTxt は host の robots.txt を取得する。
//
// baseURL は scheme 推定用（ノード URL）。空の場合は https を使用する。
// appDefaults と wsSettings は MergeUIConfigLayers 用の PartialConfig JSON。
func (s *scraperService) FetchRobotsTxt(host, baseURL string, appDefaults, wsSettings json.RawMessage) (model.RobotsTxtInfoDTO, error) {
	res, err := runner.FetchRobotsTxt(context.Background(), host, baseURL, appDefaults, wsSettings)
	if err != nil {
		return model.RobotsTxtInfoDTO{}, err
	}
	return model.RobotsTxtInfoDTO{
		Host:       res.Host,
		Status:     res.Status,
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Error:      res.Error,
	}, nil
}
