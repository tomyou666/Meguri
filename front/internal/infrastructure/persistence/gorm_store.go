package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"meguri-app/internal/model"
	"meguri-app/internal/query"

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

// nowISO は現在時刻の UTC RFC3339 文字列を返す。
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
		if len(bundle.Nodes) > 0 {
			if err := gn.WithContext(ctx).Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "workspace_id"}, {Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"url_normalized", "label", "position_x", "position_y", "user_positioned",
					"node_settings_json", "crawl_exclude", "status", "last_error", "origin",
				}),
			}).Create(ptrGraphNodes(bundle.Nodes)...); err != nil {
				return err
			}
			nodeIDs := graphNodeIDs(bundle.Nodes)
			if _, err := gn.WithContext(ctx).
				Where(gn.WorkspaceID.Eq(wid), gn.ID.NotIn(nodeIDs...)).
				Delete(); err != nil {
				return err
			}
		} else if _, err := gn.WithContext(ctx).Where(gn.WorkspaceID.Eq(wid)).Delete(); err != nil {
			return err
		}

		ge := txQ.GraphEdge
		if len(bundle.Edges) > 0 {
			edgeIDs := graphEdgeIDs(bundle.Edges)
			if _, err := ge.WithContext(ctx).
				Where(ge.WorkspaceID.Eq(wid), ge.ID.NotIn(edgeIDs...)).
				Delete(); err != nil {
				return err
			}
			if err := ge.WithContext(ctx).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "workspace_id"}, {Name: "id"}},
				DoNothing: true,
			}).Create(ptrGraphEdges(bundle.Edges)...); err != nil {
				return err
			}
		} else if _, err := ge.WithContext(ctx).Where(ge.WorkspaceID.Eq(wid)).Delete(); err != nil {
			return err
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

// GetNodeResultMetas は WS の全 node_results メタを返す。
func (s *Store) GetNodeResultMetas(ctx context.Context, workspaceID string) ([]model.NodeResult, error) {
	nr := s.q.NodeResult
	rows, err := nr.WithContext(ctx).
		Where(nr.WorkspaceID.Eq(workspaceID)).
		Order(nr.FetchedAt.Desc()).
		Find()
	return derefNodeResults(rows), err
}

// GetNodeResultMetasByNodeIDs は指定ノードのメタを fetched_at 降順で返す。
//
// nodeIDs が空のときは空スライスを返す。
// 成功／失敗の選定は呼び出し側（latestSuccessByNode 等）に委ねる。
func (s *Store) GetNodeResultMetasByNodeIDs(ctx context.Context, workspaceID string, nodeIDs []string) ([]model.NodeResult, error) {
	if len(nodeIDs) == 0 {
		return []model.NodeResult{}, nil
	}
	nr := s.q.NodeResult
	rows, err := nr.WithContext(ctx).
		Where(nr.WorkspaceID.Eq(workspaceID), nr.NodeID.In(nodeIDs...)).
		Order(nr.FetchedAt.Desc()).
		Find()
	return derefNodeResults(rows), err
}

// GetNodeResultBodies は result id 指定で本文を返す。
//
// ids が空のときは空スライスを返す。JOIN しない。
func (s *Store) GetNodeResultBodies(ctx context.Context, ids []string) ([]model.NodeResultBody, error) {
	if len(ids) == 0 {
		return []model.NodeResultBody{}, nil
	}
	nb := s.q.NodeResultBody
	rows, err := nb.WithContext(ctx).Where(nb.ID.In(ids...)).Find()
	return derefNodeResultBodies(rows), err
}

// AppendNodeResult はメタを追加し、成功行なら本文も書く。
//
// body は error が空のときのみ書き込む。その後メタ trim と本文 GC を行う。
func (s *Store) AppendNodeResult(ctx context.Context, meta model.NodeResult, body *model.NodeResultBody) error {
	return s.q.Transaction(func(tx *query.Query) error {
		ptr := meta
		if err := tx.NodeResult.WithContext(ctx).Create(&ptr); err != nil {
			return err
		}
		isSuccess := meta.Error == nil || *meta.Error == ""
		if isSuccess && body != nil {
			b := *body
			if b.ID == nil {
				b.ID = meta.ID
			}
			if err := tx.NodeResultBody.WithContext(ctx).Create(&b); err != nil {
				return err
			}
		}
		storeTx := &Store{q: tx}
		if err := storeTx.TrimNodeResults(ctx, meta.WorkspaceID, meta.NodeID, model.MaxNodeResultsPerNode); err != nil {
			return err
		}
		return storeTx.gcNodeResultBodies(ctx, meta.WorkspaceID, meta.NodeID)
	})
}

// UpdateLatestNodeResult はノードの最新成功結果行を部分更新する。
//
// メタと本文は同一トランザクションで書く。本文行が無いときは作成する。
func (s *Store) UpdateLatestNodeResult(
	ctx context.Context,
	workspaceID, nodeID string,
	patch model.NodeResultContentPatch,
) error {
	return s.q.Transaction(func(tx *query.Query) error {
		nr := tx.NodeResult
		rows, err := nr.WithContext(ctx).
			Where(nr.WorkspaceID.Eq(workspaceID), nr.NodeID.Eq(nodeID)).
			Order(nr.FetchedAt.Desc()).
			Find()
		if err != nil {
			return err
		}
		var target *model.NodeResult
		for _, row := range rows {
			if row == nil {
				continue
			}
			if row.Error != nil && *row.Error != "" {
				continue
			}
			target = row
			break
		}
		if target == nil || target.ID == nil {
			return fmt.Errorf("no successful result for node %s", nodeID)
		}
		edited := int32(0)
		if patch.ManuallyEdited {
			edited = 1
		}
		metaUpdates := []field.AssignExpr{
			nr.ManuallyEdited.Value(edited),
		}
		if patch.ContentHash != nil {
			metaUpdates = append(metaUpdates, nr.ContentHash.Value(*patch.ContentHash))
		}
		if patch.LinksHash != nil {
			metaUpdates = append(metaUpdates, nr.LinksHash.Value(*patch.LinksHash))
		}
		_, err = nr.WithContext(ctx).
			Where(nr.ID.Eq(*target.ID)).
			UpdateSimple(metaUpdates...)
		if err != nil {
			return err
		}

		nb := tx.NodeResultBody
		bodyUpdates := []field.AssignExpr{}
		if patch.Markdown != nil {
			bodyUpdates = append(bodyUpdates, nb.Markdown.Value(*patch.Markdown))
		}
		if patch.HTML != nil {
			bodyUpdates = append(bodyUpdates, nb.HTML.Value(*patch.HTML))
		}
		if patch.RawHTML != nil {
			bodyUpdates = append(bodyUpdates, nb.RawHTML.Value(*patch.RawHTML))
		}
		if patch.JSONBody != nil {
			bodyUpdates = append(bodyUpdates, nb.JSONBody.Value(*patch.JSONBody))
		}
		if len(bodyUpdates) == 0 {
			return nil
		}
		// 本文行が無い場合は upsert する。
		existing, findErr := nb.WithContext(ctx).Where(nb.ID.Eq(*target.ID)).First()
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			body := model.NodeResultBody{ID: target.ID}
			if patch.Markdown != nil {
				body.Markdown = patch.Markdown
			}
			if patch.HTML != nil {
				body.HTML = patch.HTML
			}
			if patch.RawHTML != nil {
				body.RawHTML = patch.RawHTML
			}
			if patch.JSONBody != nil {
				body.JSONBody = patch.JSONBody
			}
			return nb.WithContext(ctx).Create(&body)
		}
		if findErr != nil {
			return findErr
		}
		_, err = nb.WithContext(ctx).Where(nb.ID.Eq(*existing.ID)).UpdateSimple(bodyUpdates...)
		return err
	})
}

// DeleteNodeResult は result id 指定でメタ行を削除する。
//
// 本文は node_result_bodies の FK CASCADE で消える。
func (s *Store) DeleteNodeResult(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	nr := s.q.NodeResult
	row := &model.NodeResult{ID: model.StrPtr(id)}
	_, err := nr.WithContext(ctx).Delete(row)
	return err
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

// gcNodeResultBodies は latest 成功と baseline 以外の本文行を削除する。
func (s *Store) gcNodeResultBodies(ctx context.Context, workspaceID, nodeID string) error {
	nr := s.q.NodeResult
	metas, err := nr.WithContext(ctx).
		Where(nr.WorkspaceID.Eq(workspaceID), nr.NodeID.Eq(nodeID)).
		Order(nr.FetchedAt.Desc()).
		Find()
	if err != nil {
		return err
	}
	keep := map[string]struct{}{}
	for _, row := range metas {
		if row == nil || row.ID == nil {
			continue
		}
		if row.Error == nil || *row.Error == "" {
			keep[*row.ID] = struct{}{}
			break
		}
	}
	ws := s.q.Workspace
	workspace, err := ws.WithContext(ctx).Where(ws.ID.Eq(workspaceID)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if workspace != nil && workspace.BaselineRunID != nil && *workspace.BaselineRunID != "" {
		baselineID := *workspace.BaselineRunID
		for _, row := range metas {
			if row == nil || row.ID == nil {
				continue
			}
			if row.RunID == baselineID {
				keep[*row.ID] = struct{}{}
				break
			}
		}
	}

	nb := s.q.NodeResultBody
	metaIDList := make([]string, 0, len(metas))
	for _, row := range metas {
		if row != nil && row.ID != nil {
			metaIDList = append(metaIDList, *row.ID)
		}
	}
	if len(metaIDList) == 0 {
		return nil
	}
	bodies, err := nb.WithContext(ctx).Where(nb.ID.In(metaIDList...)).Find()
	if err != nil {
		return err
	}
	var toDelete []*model.NodeResultBody
	for _, b := range bodies {
		if b == nil || b.ID == nil {
			continue
		}
		if _, ok := keep[*b.ID]; ok {
			continue
		}
		toDelete = append(toDelete, b)
	}
	if len(toDelete) == 0 {
		return nil
	}
	_, err = nb.WithContext(ctx).Delete(toDelete...)
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

// GetGraphNodeStatuses は指定ノードの status と last_error を返す。
// nodeIDs が空のときは空スライスを返す。
func (s *Store) GetGraphNodeStatuses(ctx context.Context, workspaceID string, nodeIDs []string) ([]model.GraphNodeStatusDTO, error) {
	if len(nodeIDs) == 0 {
		return []model.GraphNodeStatusDTO{}, nil
	}
	gn := s.q.GraphNode
	rows, err := gn.WithContext(ctx).
		Select(gn.ID, gn.Status, gn.LastError).
		Where(gn.WorkspaceID.Eq(workspaceID), gn.ID.In(nodeIDs...)).
		Find()
	if err != nil {
		return nil, err
	}
	out := make([]model.GraphNodeStatusDTO, 0, len(rows))
	for _, r := range rows {
		dto := model.GraphNodeStatusDTO{NodeID: r.ID}
		if r.Status != nil {
			dto.Status = *r.Status
		}
		if r.LastError != nil {
			dto.LastError = *r.LastError
		}
		out = append(out, dto)
	}
	return out, nil
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

// derefNodeResultBodies はポインタスライスを値スライスにする。
func derefNodeResultBodies(ptrs []*model.NodeResultBody) []model.NodeResultBody {
	out := make([]model.NodeResultBody, len(ptrs))
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

func graphNodeIDs(nodes []model.GraphNode) []string {
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
	}
	return ids
}

func graphEdgeIDs(edges []model.GraphEdge) []string {
	ids := make([]string, len(edges))
	for i, e := range edges {
		ids[i] = e.ID
	}
	return ids
}
