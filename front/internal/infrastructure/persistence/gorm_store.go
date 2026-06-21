package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"scraperbot-front/internal/model"
	"scraperbot-front/internal/query"

	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Store は GORM Gen query による永続化実装。
type Store struct {
	q *query.Query
}

// NewStore は Store を構築する。
func NewStore(db *gorm.DB) *Store {
	return &Store{q: query.Use(db)}
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// GetAppConfig は app_config を取得する。
func (s *Store) GetAppConfig(ctx context.Context) (*model.AppConfig, error) {
	ac := s.q.AppConfig
	row, err := ac.WithContext(ctx).Where(ac.ID.Eq(1)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return row, err
}

// SaveAppConfig は app_config を保存する。
func (s *Store) SaveAppConfig(ctx context.Context, defaultsJSON string) error {
	row := &model.AppConfig{
		ID:           model.Int32Ptr(1),
		DefaultsJSON: defaultsJSON,
		UpdatedAt:    nowISO(),
	}
	return s.q.AppConfig.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"defaults_json", "updated_at"}),
	}).Create(row)
}

// BootstrapAppConfig は初回 app_config を挿入する。
func (s *Store) BootstrapAppConfig(ctx context.Context) error {
	row, err := s.GetAppConfig(ctx)
	if err != nil {
		return err
	}
	if row != nil {
		return nil
	}
	return s.SaveAppConfig(ctx, model.DefaultAppConfigJSON)
}

// ListWorkspaces は WS 一覧を返す。
func (s *Store) ListWorkspaces(ctx context.Context) ([]model.WorkspaceListItem, error) {
	ws := s.q.Workspace
	rows, err := ws.WithContext(ctx).Order(ws.UpdatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	out := make([]model.WorkspaceListItem, len(rows))
	for i, r := range rows {
		out[i] = model.WorkspaceListItem{
			ID:        model.StrVal(r.ID),
			Name:      r.Name,
			UpdatedAt: r.UpdatedAt,
		}
	}
	return out, nil
}

// LoadWorkspaceBundle は WS バンドルを読み込む。
func (s *Store) LoadWorkspaceBundle(ctx context.Context, id string) (*model.WorkspaceBundle, error) {
	wsQ := s.q.Workspace
	ws, err := wsQ.WithContext(ctx).Where(wsQ.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	gn := s.q.GraphNode
	nodes, err := gn.WithContext(ctx).Where(gn.WorkspaceID.Eq(id)).Find()
	if err != nil {
		return nil, err
	}
	ge := s.q.GraphEdge
	edges, err := ge.WithContext(ctx).Where(ge.WorkspaceID.Eq(id)).Find()
	if err != nil {
		return nil, err
	}

	var uiPtr *model.GraphUIState
	gu := s.q.GraphUIState
	ui, err := gu.WithContext(ctx).Where(gu.WorkspaceID.Eq(id)).First()
	if err == nil {
		uiPtr = ui
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &model.WorkspaceBundle{
		Workspace: derefWorkspace(ws),
		Nodes:     derefGraphNodes(nodes),
		Edges:     derefGraphEdges(edges),
		UIState:   uiPtr,
	}, nil
}

// SaveWorkspaceBundle は WS バンドルを upsert する。
func (s *Store) SaveWorkspaceBundle(ctx context.Context, bundle model.WorkspaceBundle) error {
	return s.q.Transaction(func(txQ *query.Query) error {
		ws := bundle.Workspace
		ws.UpdatedAt = nowISO()
		if ws.CreatedAt == "" {
			ws.CreatedAt = ws.UpdatedAt
		}
		wsPtr := ws
		if err := txQ.Workspace.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "seed_url", "settings_json", "exclude_urls_json",
				"graph_layout_direction", "baseline_run_id", "updated_at",
			}),
		}).Create(&wsPtr); err != nil {
			return err
		}

		wid := model.StrVal(ws.ID)
		gn := txQ.GraphNode
		if _, err := gn.WithContext(ctx).Where(gn.WorkspaceID.Eq(wid)).Delete(); err != nil {
			return err
		}
		ge := txQ.GraphEdge
		if _, err := ge.WithContext(ctx).Where(ge.WorkspaceID.Eq(wid)).Delete(); err != nil {
			return err
		}

		if len(bundle.Nodes) > 0 {
			if err := txQ.GraphNode.WithContext(ctx).Create(ptrGraphNodes(bundle.Nodes)...); err != nil {
				return err
			}
		}
		if len(bundle.Edges) > 0 {
			if err := txQ.GraphEdge.WithContext(ctx).Create(ptrGraphEdges(bundle.Edges)...); err != nil {
				return err
			}
		}
		if bundle.UIState != nil {
			ui := *bundle.UIState
			ui.WorkspaceID = model.StrPtr(wid)
			if err := txQ.GraphUIState.WithContext(ctx).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "workspace_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"collapsed_node_ids_json"}),
			}).Create(&ui); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteWorkspace は WS を削除する。
func (s *Store) DeleteWorkspace(ctx context.Context, id string) error {
	ws := s.q.Workspace
	_, err := ws.WithContext(ctx).Where(ws.ID.Eq(id)).Delete()
	return err
}

// GetNodeResults は WS の全 node_results を返す。
func (s *Store) GetNodeResults(ctx context.Context, workspaceID string) ([]model.NodeResult, error) {
	nr := s.q.NodeResult
	rows, err := nr.WithContext(ctx).
		Where(nr.WorkspaceID.Eq(workspaceID)).
		Order(nr.FetchedAt.Desc()).
		Find()
	return derefNodeResults(rows), err
}

// AppendNodeResult は結果行を追加する。
func (s *Store) AppendNodeResult(ctx context.Context, row model.NodeResult) error {
	ptr := row
	if err := s.q.NodeResult.WithContext(ctx).Create(&ptr); err != nil {
		return err
	}
	return s.TrimNodeResults(ctx, row.WorkspaceID, row.NodeID, model.MaxNodeResultsPerNode)
}

// DeleteLatestResults は各ノードの最新 1 行を削除する。
func (s *Store) DeleteLatestResults(ctx context.Context, workspaceID string, nodeIDs []string) error {
	nr := s.q.NodeResult
	for _, nodeID := range nodeIDs {
		row, err := nr.WithContext(ctx).
			Where(nr.WorkspaceID.Eq(workspaceID), nr.NodeID.Eq(nodeID)).
			Order(nr.FetchedAt.Desc()).
			First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if _, err := nr.WithContext(ctx).Delete(row); err != nil {
			return err
		}
	}
	return nil
}

// TrimNodeResults はノードごとの履歴を keep 件に切り詰める。
func (s *Store) TrimNodeResults(ctx context.Context, workspaceID, nodeID string, keep int) error {
	nr := s.q.NodeResult
	rows, err := nr.WithContext(ctx).
		Where(nr.WorkspaceID.Eq(workspaceID), nr.NodeID.Eq(nodeID)).
		Order(nr.FetchedAt.Desc()).
		Find()
	if err != nil {
		return err
	}
	if len(rows) <= keep {
		return nil
	}
	_, err = nr.WithContext(ctx).Delete(rows[keep:]...)
	return err
}

// GetCrawlRuns は crawl_runs を返す。
func (s *Store) GetCrawlRuns(ctx context.Context, workspaceID string) ([]model.CrawlRun, error) {
	cr := s.q.CrawlRun
	rows, err := cr.WithContext(ctx).
		Where(cr.WorkspaceID.Eq(workspaceID)).
		Order(cr.StartedAt.Desc()).
		Find()
	return derefCrawlRuns(rows), err
}

// BeginCrawlRun は crawl run を開始する。
func (s *Store) BeginCrawlRun(ctx context.Context, run model.CrawlRun) error {
	ptr := run
	if err := s.q.CrawlRun.WithContext(ctx).Create(&ptr); err != nil {
		return err
	}
	return s.TrimCrawlRuns(ctx, run.WorkspaceID, model.MaxCrawlRunHistory)
}

// FinishCrawlRun は crawl run を終了する。
func (s *Store) FinishCrawlRun(ctx context.Context, runID, status, finishedAt string, summaryJSON, errorMessage *string) error {
	cr := s.q.CrawlRun
	assigns := []field.AssignExpr{
		cr.Status.Value(status),
		cr.FinishedAt.Value(finishedAt),
	}
	if summaryJSON != nil {
		assigns = append(assigns, cr.SummaryJSON.Value(*summaryJSON))
	}
	if errorMessage != nil {
		assigns = append(assigns, cr.ErrorMessage.Value(*errorMessage))
	}
	info, err := cr.WithContext(ctx).Where(cr.ID.Eq(runID)).UpdateSimple(assigns...)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return fmt.Errorf("crawl run not found: %s", runID)
	}
	return nil
}

// TrimCrawlRuns は WS の run 履歴を keep 件に切り詰める。
func (s *Store) TrimCrawlRuns(ctx context.Context, workspaceID string, keep int) error {
	cr := s.q.CrawlRun
	rows, err := cr.WithContext(ctx).
		Where(cr.WorkspaceID.Eq(workspaceID)).
		Order(cr.StartedAt.Desc()).
		Find()
	if err != nil {
		return err
	}
	if len(rows) <= keep {
		return nil
	}
	_, err = cr.WithContext(ctx).Delete(rows[keep:]...)
	return err
}

// UpsertDiscoveredGraph は crawl 中に発見したノードとエッジを追加する。
func (s *Store) UpsertDiscoveredGraph(ctx context.Context, workspaceID, sourceNodeID, targetNodeID, targetURL string) error {
	if workspaceID == "" || sourceNodeID == "" || targetNodeID == "" || targetURL == "" {
		return fmt.Errorf("upsert discovered graph: missing required fields")
	}
	if sourceNodeID == targetNodeID {
		return nil
	}
	return s.q.Transaction(func(txQ *query.Query) error {
		idle := "idle"
		node := model.GraphNode{
			WorkspaceID:      workspaceID,
			ID:               targetNodeID,
			URLNormalized:    targetURL,
			Label:            targetURL,
			NodeSettingsJSON: "{}",
			Origin:           "crawl",
			Status:           &idle,
		}
		gn := txQ.GraphNode
		if err := gn.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "workspace_id"}, {Name: "id"}},
			DoNothing: true,
		}).Create(&node); err != nil {
			return err
		}

		edgeID := fmt.Sprintf("e-%s-%s", sourceNodeID, targetNodeID)
		edge := model.GraphEdge{
			WorkspaceID:  workspaceID,
			ID:           edgeID,
			SourceNodeID: sourceNodeID,
			TargetNodeID: targetNodeID,
		}
		ge := txQ.GraphEdge
		return ge.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "workspace_id"}, {Name: "id"}},
			DoNothing: true,
		}).Create(&edge)
	})
}

// PatchGraphNodeStatus は graph_nodes.status を更新する。
func (s *Store) PatchGraphNodeStatus(ctx context.Context, workspaceID, nodeID, status string, lastError *string) error {
	gn := s.q.GraphNode
	assigns := []field.AssignExpr{gn.Status.Value(status)}
	if lastError != nil {
		assigns = append(assigns, gn.LastError.Value(*lastError))
	} else {
		assigns = append(assigns, gn.LastError.Null())
	}
	info, err := gn.WithContext(ctx).
		Where(gn.WorkspaceID.Eq(workspaceID), gn.ID.Eq(nodeID)).
		UpdateSimple(assigns...)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return fmt.Errorf("node not found: %s/%s", workspaceID, nodeID)
	}
	return nil
}

// PatchGraphNodePositions は graph_nodes の座標を部分更新する。
func (s *Store) PatchGraphNodePositions(ctx context.Context, workspaceID string, updates []model.NodePositionPatchDTO) error {
	if len(updates) == 0 {
		return nil
	}
	return s.q.Transaction(func(txQ *query.Query) error {
		gn := txQ.GraphNode
		for _, u := range updates {
			userPos := int32(0)
			if u.UserPositioned {
				userPos = 1
			}
			info, err := gn.WithContext(ctx).
				Where(gn.WorkspaceID.Eq(workspaceID), gn.ID.Eq(u.NodeID)).
				UpdateSimple(
					gn.PositionX.Value(u.Position.X),
					gn.PositionY.Value(u.Position.Y),
					gn.UserPositioned.Value(userPos),
				)
			if err != nil {
				return err
			}
			if info.RowsAffected == 0 {
				return fmt.Errorf("node not found: %s/%s", workspaceID, u.NodeID)
			}
		}
		ws := txQ.Workspace
		_, err := ws.WithContext(ctx).
			Where(ws.ID.Eq(workspaceID)).
			UpdateSimple(ws.UpdatedAt.Value(nowISO()))
		return err
	})
}

// SetBaselineRunID は baseline_run_id を更新する。
func (s *Store) SetBaselineRunID(ctx context.Context, workspaceID, runID string) error {
	ws := s.q.Workspace
	_, err := ws.WithContext(ctx).
		Where(ws.ID.Eq(workspaceID)).
		UpdateSimple(ws.BaselineRunID.Value(runID))
	return err
}

func derefWorkspace(p *model.Workspace) model.Workspace {
	if p == nil {
		return model.Workspace{}
	}
	return *p
}

func derefGraphNodes(ptrs []*model.GraphNode) []model.GraphNode {
	out := make([]model.GraphNode, len(ptrs))
	for i, p := range ptrs {
		if p != nil {
			out[i] = *p
		}
	}
	return out
}

func derefGraphEdges(ptrs []*model.GraphEdge) []model.GraphEdge {
	out := make([]model.GraphEdge, len(ptrs))
	for i, p := range ptrs {
		if p != nil {
			out[i] = *p
		}
	}
	return out
}

func derefNodeResults(ptrs []*model.NodeResult) []model.NodeResult {
	out := make([]model.NodeResult, len(ptrs))
	for i, p := range ptrs {
		if p != nil {
			out[i] = *p
		}
	}
	return out
}

func derefCrawlRuns(ptrs []*model.CrawlRun) []model.CrawlRun {
	out := make([]model.CrawlRun, len(ptrs))
	for i, p := range ptrs {
		if p != nil {
			out[i] = *p
		}
	}
	return out
}

func ptrGraphNodes(rows []model.GraphNode) []*model.GraphNode {
	out := make([]*model.GraphNode, len(rows))
	for i := range rows {
		out[i] = &rows[i]
	}
	return out
}

func ptrGraphEdges(rows []model.GraphEdge) []*model.GraphEdge {
	out := make([]*model.GraphEdge, len(rows))
	for i := range rows {
		out[i] = &rows[i]
	}
	return out
}
