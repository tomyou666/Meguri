package wails_service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"scraperbot-front/internal/domain"
	"scraperbot-front/internal/model"
	"scraperbot/pkg/runner"

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

// ScraperService は backend crawler を駆動し Wails Event で進捗を配信する。
type ScraperService struct {
	app     *application.App
	persist *domain.CrawlPersistService
	mu      sync.Mutex
	job     *activeCrawlJob
}

type activeCrawlJob struct {
	runID  string
	pause  *runner.PauseController
	cache  *runner.RunnerCache
	cancel context.CancelFunc
}

// NewScraperService は ScraperService を構築する。
func NewScraperService(persist *domain.CrawlPersistService) *ScraperService {
	return &ScraperService{persist: persist}
}

// SetApp は Wails App を後から注入する（Event 発火用）。
func (s *ScraperService) SetApp(app *application.App) {
	s.app = app
}

// StartCrawl はクロールを非同期で開始し runId を返す。
func (s *ScraperService) StartCrawl(req model.StartCrawlRequest) (string, error) {
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
	s.job = &activeCrawlJob{
		runID:  runID,
		pause:  runner.NewPauseController(),
		cache:  runner.NewRunnerCache(),
		cancel: cancel,
	}
	s.mu.Unlock()

	s.emit(topicRunStarted, model.CrawlEventPayload{
		WorkspaceID: req.WorkspaceID,
		RunID:       runID,
	})

	go func() {
		defer func() {
			s.mu.Lock()
			if s.job != nil && s.job.cache != nil {
				s.job.cache.CloseAll()
			}
			s.job = nil
			s.mu.Unlock()
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
func (s *ScraperService) PauseCrawl(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil || s.job.runID != runID {
		return fmt.Errorf("no active crawl for runId %s", runID)
	}
	s.job.pause.Pause()
	return nil
}

// ResumeCrawl は一時停止中 crawl を再開する。
func (s *ScraperService) ResumeCrawl(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil || s.job.runID != runID {
		return fmt.Errorf("no active crawl for runId %s", runID)
	}
	s.job.pause.Resume()
	return nil
}

// StopCrawl は実行中 crawl をキャンセルする。
func (s *ScraperService) StopCrawl(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil || s.job.runID != runID {
		return nil
	}
	s.job.cancel()
	return nil
}

func (s *ScraperService) runOptions() *runner.RunOptions {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil {
		return nil
	}
	return &runner.RunOptions{
		Pause: s.job.pause,
		Cache: s.job.cache,
	}
}

func (s *ScraperService) emit(topic string, payload model.CrawlEventPayload) {
	if s.app == nil {
		return
	}
	s.app.Event.Emit(topic, payload)
}

func (s *ScraperService) finishCrawlRun(
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

func (s *ScraperService) persistNodeStarted(ctx context.Context, req model.StartCrawlRequest, nodeID string) {
	if s.persist == nil || nodeID == "" {
		return
	}
	_ = s.persist.PatchGraphNodeStatus(ctx, model.PatchGraphNodeStatusRequest{
		WorkspaceID: req.WorkspaceID,
		NodeID:      nodeID,
		Status:      "running",
	})
}

func (s *ScraperService) persistNodeSucceeded(
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
	_ = s.persist.AppendNodeResult(ctx, model.AppendNodeResultRequest{
		WorkspaceID:  req.WorkspaceID,
		RunID:        req.RunID,
		NodeID:       nodeID,
		URL:          url,
		Markdown:     markdown,
		LinksJSON:    linksJSON,
		MetadataJSON: metadataJSON,
		FetchedAt:    domain.NowISO(),
		ContentHash:  contentHash,
	})
	_ = s.persist.PatchGraphNodeStatus(ctx, model.PatchGraphNodeStatusRequest{
		WorkspaceID: req.WorkspaceID,
		NodeID:      nodeID,
		Status:      "success",
	})
}

func (s *ScraperService) persistNodeFailed(
	ctx context.Context,
	req model.StartCrawlRequest,
	nodeID, url, errMsg string,
) {
	if s.persist == nil || nodeID == "" {
		return
	}
	_ = s.persist.AppendNodeResult(ctx, model.AppendNodeResultRequest{
		WorkspaceID: req.WorkspaceID,
		RunID:       req.RunID,
		NodeID:      nodeID,
		URL:         url,
		Error:       errMsg,
		FetchedAt:   domain.NowISO(),
	})
	_ = s.persist.PatchGraphNodeStatus(ctx, model.PatchGraphNodeStatusRequest{
		WorkspaceID: req.WorkspaceID,
		NodeID:      nodeID,
		Status:      "error",
		LastError:   errMsg,
	})
}

func (s *ScraperService) persistNodeSkipped(ctx context.Context, req model.StartCrawlRequest, nodeID string) {
	if s.persist == nil || nodeID == "" {
		return
	}
	_ = s.persist.PatchGraphNodeStatus(ctx, model.PatchGraphNodeStatusRequest{
		WorkspaceID: req.WorkspaceID,
		NodeID:      nodeID,
		Status:      "skipped",
	})
}

func (s *ScraperService) persistEdgeDiscovered(
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

func (s *ScraperService) runCrawl(ctx context.Context, req model.StartCrawlRequest) error {
	ws := req.Workspace
	state := newCrawlState(req)

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
		stats, mainReached, err := s.runMainBFS(ctx, req, state, s.runOptions(), &enqueued, &succeeded, &failed, &skipped)
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
		if err := s.runManualPass(ctx, req, state, mainReached, s.runOptions(), &enqueued, &succeeded, &failed, &skipped); err != nil {
			if ctx.Err() != nil {
				stoppedReason = "stopped"
				emitSummary()
				return nil
			}
			return err
		}
	case 3:
		if err := s.runMode3(ctx, req, state, s.runOptions(), &enqueued, &succeeded, &failed, &skipped); err != nil {
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
		if err := s.runManualPass(ctx, req, state, mainReached, s.runOptions(), &enqueued, &succeeded, &failed, &skipped); err != nil {
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
	domainMap        map[string]json.RawMessage
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
		domainMap:        ws.DomainSettings,
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

func (s *ScraperService) emitLinkSkipped(
	req model.StartCrawlRequest,
	st *crawlState,
	parentURL, childURL, reason string,
) {
	st.mu.Lock()
	st.linkSkippedCount++
	st.mu.Unlock()
	s.emit(topicLinkSkipped, model.CrawlEventPayload{
		WorkspaceID: req.WorkspaceID,
		RunID:       req.RunID,
		URL:         parentURL,
		TargetURL:   childURL,
		Reason:      reason,
	})
}

func (st *crawlState) hostFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

func (st *crawlState) mergedConfig(mode int32, node model.GraphNodeDTO) (*runner.Config, error) {
	layers := []json.RawMessage{st.appDefaults}
	if mode != 2 {
		layers = append(layers, st.wsSettings)
		if host := st.hostFromURL(node.URLNormalized); host != "" {
			if d, ok := st.domainMap[host]; ok {
				layers = append(layers, d)
			}
		}
		if mode != 2 {
			layers = append(layers, node.NodeSettings)
		}
	}
	merged, err := runner.MergeUIConfigLayers(layers...)
	if err != nil {
		return nil, err
	}
	cfg, err := runner.ParseUIConfig(merged)
	if err != nil {
		return nil, err
	}
	exclude := make([]string, 0, len(st.excludeSet))
	for u := range st.excludeSet {
		exclude = append(exclude, u)
	}
	cfg.Crawl.ExcludeURLs = exclude
	cfg.Targets = []string{node.URLNormalized}
	return cfg, nil
}

func (s *ScraperService) runMainBFS(
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

	seedNode, ok := st.nodeByID[st.urlToNode[seedURL]]
	if !ok {
		for _, n := range ws.Nodes {
			if n.URLNormalized == seedURL {
				seedNode = n
				ok = true
				break
			}
		}
	}
	if !ok {
		return nil, mainReached, fmt.Errorf("seed node not found for %s", seedURL)
	}

	cfg, err := st.mergedConfig(req.Mode, seedNode)
	if err != nil {
		return nil, mainReached, err
	}
	cfg.Crawl.Enabled = true
	cfg.Crawl.SkipScrapeURLs = st.skipScrapeURLs()

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
				Result:      dto,
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
				s.emitLinkSkipped(req, st, ev.ParentURL, urlKey, "duplicate_existing")
				return
			}
			nodeID, _ := st.resolveNodeID(ev.URL, false)
			if nodeID == "" {
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

func (s *ScraperService) runMode3(
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

func (s *ScraperService) runManualPass(
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

func (s *ScraperService) scrapeOneNode(
	ctx context.Context,
	req model.StartCrawlRequest,
	st *crawlState,
	node model.GraphNodeDTO,
	opts *runner.RunOptions,
) error {
	cfg, err := st.mergedConfig(req.Mode, node)
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
				Result:      dto,
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
