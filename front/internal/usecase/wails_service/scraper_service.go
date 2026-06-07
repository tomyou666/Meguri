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
	topicNodeStarted    = "scraper:crawl:nodeStarted"
	topicNodeSucceeded  = "scraper:crawl:nodeSucceeded"
	topicNodeFailed     = "scraper:crawl:nodeFailed"
	topicNodeSkipped    = "scraper:crawl:nodeSkipped"
	topicEdgeDiscovered = "scraper:crawl:edgeDiscovered"
	topicCrawlCompleted = "scraper:crawl:completed"
	topicCrawlError     = "scraper:crawl:error"
)

// ScraperService は backend crawler を駆動し Wails Event で進捗を配信する。
type ScraperService struct {
	app *application.App
	mu  sync.Mutex
	job *activeCrawlJob
}

type activeCrawlJob struct {
	runID  string
	paused bool
	cancel context.CancelFunc
}

// NewScraperService は ScraperService を構築する。
func NewScraperService() *ScraperService {
	return &ScraperService{}
}

// SetApp は Wails App を後から注入する（Event 発火用）。
func (s *ScraperService) SetApp(app *application.App) {
	s.app = app
}

// StartCrawl はクロールを非同期で開始する。
func (s *ScraperService) StartCrawl(req model.StartCrawlRequest) error {
	if s.app == nil {
		return fmt.Errorf("app not initialized")
	}
	if req.RunID == "" || req.WorkspaceID == "" {
		return fmt.Errorf("runId and workspaceId are required")
	}

	s.mu.Lock()
	if s.job != nil {
		s.mu.Unlock()
		return fmt.Errorf("crawl already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.job = &activeCrawlJob{runID: req.RunID, cancel: cancel}
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			s.job = nil
			s.mu.Unlock()
		}()
		if err := s.runCrawl(ctx, req); err != nil {
			s.emit(topicCrawlError, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				Message:     err.Error(),
			})
		}
	}()
	return nil
}

// PauseCrawl は実行中 crawl を一時停止する。
func (s *ScraperService) PauseCrawl(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil || s.job.runID != runID {
		return fmt.Errorf("no active crawl for runId %s", runID)
	}
	s.job.paused = true
	return nil
}

// ResumeCrawl は一時停止中 crawl を再開する。
func (s *ScraperService) ResumeCrawl(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil || s.job.runID != runID {
		return fmt.Errorf("no active crawl for runId %s", runID)
	}
	s.job.paused = false
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

func (s *ScraperService) waitIfPaused(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		s.mu.Lock()
		paused := s.job != nil && s.job.paused
		s.mu.Unlock()
		if !paused {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *ScraperService) emit(topic string, payload model.CrawlEventPayload) {
	if s.app == nil {
		return
	}
	s.app.Event.Emit(topic, payload)
}

func (s *ScraperService) runCrawl(ctx context.Context, req model.StartCrawlRequest) error {
	ws := req.Workspace
	state := newCrawlState(req)

	var (
		enqueued, succeeded, failed, skipped int
		stoppedReason                        = "completed"
	)

	emitSummary := func() {
		s.emit(topicCrawlCompleted, model.CrawlEventPayload{
			WorkspaceID: req.WorkspaceID,
			RunID:       req.RunID,
			Summary: &model.CrawlSummaryDTO{
				Mode:          req.Mode,
				FinishedAt:    time.Now().UTC().Format(time.RFC3339),
				Enqueued:      enqueued,
				Succeeded:     succeeded,
				Failed:        failed,
				Skipped:       skipped,
				StoppedReason: stoppedReason,
			},
		})
	}

	switch req.Mode {
	case 1, 2:
		stats, mainReached, err := s.runMainBFS(ctx, req, state, &enqueued, &succeeded, &failed, &skipped)
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
		if err := s.runManualPass(ctx, req, state, mainReached, &enqueued, &succeeded, &failed, &skipped); err != nil {
			if ctx.Err() != nil {
				stoppedReason = "stopped"
				emitSummary()
				return nil
			}
			return err
		}
	case 3:
		if err := s.runMode3(ctx, req, state, &enqueued, &succeeded, &failed, &skipped); err != nil {
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
		if err := s.runManualPass(ctx, req, state, mainReached, &enqueued, &succeeded, &failed, &skipped); err != nil {
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
	mu          sync.Mutex
	nextNodeSeq int64
	urlToNode   map[string]string
	nodeByID    map[string]model.GraphNodeDTO
	excludeSet  map[string]struct{}
	outEdges    map[string][]string
	appDefaults json.RawMessage
	wsSettings  json.RawMessage
	domainMap   map[string]json.RawMessage
}

func newCrawlState(req model.StartCrawlRequest) *crawlState {
	ws := req.Workspace
	st := &crawlState{
		urlToNode:   map[string]string{},
		nodeByID:    map[string]model.GraphNodeDTO{},
		excludeSet:  map[string]struct{}{},
		outEdges:    map[string][]string{},
		appDefaults: req.AppDefaults,
		wsSettings:  ws.Settings,
		domainMap:   ws.DomainSettings,
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
	}
	for _, e := range ws.Edges {
		st.outEdges[e.Source] = append(st.outEdges[e.Source], e.Target)
	}
	return st
}

func (st *crawlState) nodeIDForURL(rawURL string, create bool) (string, bool) {
	key := rawURL
	if norm, err := domain.NormalizeCrawlURL(rawURL); err == nil {
		key = norm
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.nodeIDForURLLocked(key, create)
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
			childID, _ := st.nodeIDForURLLocked(childKey, true)
			st.mu.Unlock()
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
			s.emit(topicNodeSucceeded, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      nodeID,
				URL:         urlKey,
				Result:      resultToDTO(ev.Result),
			})
		case runner.ProgressFailed:
			nodeID, urlKey := st.resolveNodeID(ev.URL, false)
			if nodeID == "" {
				return
			}
			s.emit(topicNodeFailed, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      nodeID,
				URL:         urlKey,
				Error:       ev.Error,
			})
		case runner.ProgressSkipped:
			nodeID, urlKey := st.resolveNodeID(ev.URL, true)
			if nodeID == "" {
				return
			}
			s.emit(topicNodeSkipped, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      nodeID,
				URL:         urlKey,
				Reason:      ev.SkipReason,
			})
		}
	}

	if err := s.waitIfPaused(ctx); err != nil {
		return nil, mainReached, err
	}

	stats, err := runner.CrawlWithProgress(ctx, cfg, []string{seedURL}, progress)
	return stats, mainReached, err
}

func (s *ScraperService) runMode3(
	ctx context.Context,
	req model.StartCrawlRequest,
	st *crawlState,
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
		if err := s.waitIfPaused(ctx); err != nil {
			return err
		}
		node, ok := st.nodeByID[nodeID]
		if !ok {
			continue
		}
		if _, ex := st.excludeSet[node.URLNormalized]; ex || node.CrawlExclude {
			*skipped++
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
		if err := s.scrapeOneNode(ctx, req, st, node); err != nil {
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
	enqueued, succeeded, failed, skipped *int,
) error {
	for _, node := range st.nodeByID {
		if node.Origin != "manual" {
			continue
		}
		if _, reached := mainReached[node.ID]; reached {
			continue
		}
		if node.Status == "success" {
			continue
		}
		if _, ex := st.excludeSet[node.URLNormalized]; ex || node.CrawlExclude {
			*skipped++
			s.emit(topicNodeSkipped, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      node.ID,
				URL:         node.URLNormalized,
				Reason:      "exclude_urls",
			})
			continue
		}
		if err := s.waitIfPaused(ctx); err != nil {
			return err
		}
		*enqueued++
		if err := s.scrapeOneNode(ctx, req, st, node); err != nil {
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
			s.emit(topicNodeStarted, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      node.ID,
				URL:         ev.URL,
			})
		case runner.ProgressSucceeded:
			s.emit(topicNodeSucceeded, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      node.ID,
				URL:         ev.URL,
				Result:      resultToDTO(ev.Result),
			})
		case runner.ProgressFailed:
			failed = true
			s.emit(topicNodeFailed, model.CrawlEventPayload{
				WorkspaceID: req.WorkspaceID,
				RunID:       req.RunID,
				NodeID:      node.ID,
				URL:         ev.URL,
				Error:       ev.Error,
			})
		}
	}

	_, err = runner.ScrapeWithConfig(ctx, node.URLNormalized, cfg, progress)
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
