package wails_service

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"meguri-app/internal/domain"
	"meguri-app/internal/domain/exportpreview"
	"meguri-app/internal/model"
)

// ClearExportPreviewCache はプレビュー行キャッシュを捨てる。
func (s *storeService) ClearExportPreviewCache() {
	if s.exportPreview == nil {
		s.exportPreview = &exportPreviewCache{}
	}
	s.exportPreview.clear()
}

// GetExportPreviewMeta は全行のバイト長と先頭行フラグを返す（本文なし）。
func (s *storeService) GetExportPreviewMeta(req model.ExportPreviewMetaRequest) (model.ExportPreviewMetaResponse, error) {
	if s.exportPreview == nil {
		s.exportPreview = &exportPreviewCache{}
	}
	format, err := parseExportFormat(req.Format)
	if err != nil {
		return model.ExportPreviewMetaResponse{}, err
	}

	bodies, err := s.results.GetExportNodeBodies(s.ctx(), req.WorkspaceID, req.NodeIDs, string(format))
	if err != nil {
		return model.ExportPreviewMetaResponse{}, err
	}
	byNode := make(map[string]domain.ExportNodeBody, len(bodies))
	for _, b := range bodies {
		byNode[b.NodeID] = b
	}

	skipped := 0
	rows := make([]exportPreviewRow, 0)
	metaRows := make([]model.ExportPreviewRowMetaDTO, 0)
	for _, nodeID := range req.NodeIDs {
		body, ok := byNode[nodeID]
		if !ok || body.Body == "" {
			skipped++
			continue
		}
		var frags []exportpreview.Fragment
		if format == exportpreview.FormatHTML {
			frags = exportpreview.SplitHTML(body.Body)
		} else {
			frags = exportpreview.SplitMarkdown(body.Body)
		}
		if len(frags) == 0 {
			skipped++
			continue
		}
		for i, frag := range frags {
			id := exportpreview.RowID(nodeID, frag.Start)
			row := exportPreviewRow{
				id:         id,
				nodeID:     nodeID,
				url:        body.URL,
				byteLength: frag.Length,
				isFirst:    i == 0,
				start:      frag.Start,
				length:     frag.Length,
			}
			rows = append(rows, row)
			metaRows = append(metaRows, model.ExportPreviewRowMetaDTO{
				ID:         id,
				NodeID:     nodeID,
				ByteLength: frag.Length,
				IsFirst:    i == 0,
			})
		}
	}

	s.exportPreview.store(&exportPreviewSession{
		workspaceID: req.WorkspaceID,
		format:      format,
		generation:  req.Generation,
		nodeIDs:     append([]string(nil), req.NodeIDs...),
		rows:        rows,
	})

	return model.ExportPreviewMetaResponse{
		Generation:   req.Generation,
		Rows:         metaRows,
		SkippedCount: skipped,
	}, nil
}

// GetExportPreviewBodies は index 範囲の本文断片を返す。
func (s *storeService) GetExportPreviewBodies(req model.ExportPreviewBodiesRequest) (model.ExportPreviewBodiesResponse, error) {
	if s.exportPreview == nil {
		s.exportPreview = &exportPreviewCache{}
	}
	format, err := parseExportFormat(req.Format)
	if err != nil {
		return model.ExportPreviewBodiesResponse{}, err
	}
	session := s.exportPreview.get(req.WorkspaceID, format, req.Generation)
	if session == nil {
		return model.ExportPreviewBodiesResponse{}, fmt.Errorf("export preview session expired")
	}
	start := req.StartIndex
	end := req.EndIndex
	if start < 0 {
		start = 0
	}
	if end > len(session.rows) {
		end = len(session.rows)
	}
	if start > end {
		start = end
	}
	if start >= end {
		return model.ExportPreviewBodiesResponse{
			Generation: req.Generation,
			StartIndex: start,
			EndIndex:   end,
			Rows:       []model.ExportPreviewRowBodyDTO{},
		}, nil
	}

	needed := map[string]struct{}{}
	for i := start; i < end; i++ {
		needed[session.rows[i].nodeID] = struct{}{}
	}
	nodeIDs := make([]string, 0, len(needed))
	for id := range needed {
		nodeIDs = append(nodeIDs, id)
	}
	bodies, err := s.results.GetExportNodeBodies(s.ctx(), req.WorkspaceID, nodeIDs, string(format))
	if err != nil {
		return model.ExportPreviewBodiesResponse{}, err
	}
	bodyByNode := make(map[string]domain.ExportNodeBody, len(bodies))
	for _, b := range bodies {
		bodyByNode[b.NodeID] = b
	}

	out := make([]model.ExportPreviewRowBodyDTO, 0, end-start)
	for i := start; i < end; i++ {
		row := session.rows[i]
		src, ok := bodyByNode[row.nodeID]
		if !ok {
			continue
		}
		full := src.Body
		if row.start < 0 || row.start+row.length > len(full) {
			continue
		}
		out = append(out, model.ExportPreviewRowBodyDTO{
			ID:     row.id,
			NodeID: row.nodeID,
			URL:    row.url,
			Body:   full[row.start : row.start+row.length],
		})
	}
	return model.ExportPreviewBodiesResponse{
		Generation: req.Generation,
		StartIndex: start,
		EndIndex:   end,
		Rows:       out,
	}, nil
}

// parseExportFormat は文字列を Format にする。
func parseExportFormat(raw string) (exportpreview.Format, error) {
	switch strings.ToLower(raw) {
	case "html":
		return exportpreview.FormatHTML, nil
	case "markdown", "md", "":
		return exportpreview.FormatMarkdown, nil
	default:
		return "", fmt.Errorf("unsupported export format: %s", raw)
	}
}

// exportWritable は保存可能な 1 ノード分。
type exportWritable struct {
	meta    exportpreview.NodeMeta
	content string
}

// exportSaveState は進行中の保存を追跡する。
type exportSaveState struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	tmp    string
}

// CancelExportSave は進行中の保存を中止する。
func (s *storeService) CancelExportSave() error {
	if s.exportSave == nil {
		return nil
	}
	s.exportSave.mu.Lock()
	defer s.exportSave.mu.Unlock()
	if s.exportSave.cancel != nil {
		s.exportSave.cancel()
	}
	if s.exportSave.tmp != "" {
		_ = os.Remove(s.exportSave.tmp)
		s.exportSave.tmp = ""
	}
	return nil
}

// SaveExport はダイアログ後に Go が追記保存する。
func (s *storeService) SaveExport(req model.SaveExportRequest) error {
	if s.app == nil {
		return fmt.Errorf("app not initialized")
	}
	if s.exportSave == nil {
		s.exportSave = &exportSaveState{}
	}
	format, err := parseExportFormat(req.Format)
	if err != nil {
		return err
	}
	settings := exportpreview.MergeSettings{
		Format:         format,
		Separator:      req.Separator,
		IncludeHeading: req.IncludeHeading,
		HeadingField:   exportpreview.HeadingField(req.HeadingField),
		SplitSave:      req.SplitSave,
	}
	if settings.HeadingField != exportpreview.HeadingLabel {
		settings.HeadingField = exportpreview.HeadingURL
	}

	ext := exportpreview.ExtForFormat(format)
	path, err := s.promptExportPath(req.SplitSave, ext)
	if err != nil || path == "" {
		return err
	}

	ctx, cancel := context.WithCancel(s.ctx())
	s.exportSave.mu.Lock()
	if s.exportSave.cancel != nil {
		s.exportSave.cancel()
	}
	s.exportSave.cancel = cancel
	s.exportSave.tmp = ""
	s.exportSave.mu.Unlock()
	defer func() {
		s.exportSave.mu.Lock()
		s.exportSave.cancel = nil
		s.exportSave.tmp = ""
		s.exportSave.mu.Unlock()
		cancel()
	}()

	ws, err := s.workspaces.Load(ctx, req.WorkspaceID)
	if err != nil || ws == nil {
		return fmt.Errorf("workspace not found")
	}
	metaByID := make(map[string]exportpreview.NodeMeta, len(ws.Nodes))
	for _, n := range ws.Nodes {
		metaByID[n.ID] = exportpreview.NodeMeta{
			ID:            n.ID,
			URLNormalized: n.URLNormalized,
			Label:         n.Label,
		}
	}

	bodies, err := s.results.GetExportNodeBodies(ctx, req.WorkspaceID, req.NodeIDs, string(format))
	if err != nil {
		return err
	}
	bodyByNode := make(map[string]domain.ExportNodeBody, len(bodies))
	for _, b := range bodies {
		bodyByNode[b.NodeID] = b
	}

	writables := make([]exportWritable, 0, len(req.NodeIDs))
	for _, id := range req.NodeIDs {
		meta, ok := metaByID[id]
		if !ok {
			continue
		}
		body, ok := bodyByNode[id]
		if !ok || body.Body == "" {
			continue
		}
		content := exportpreview.WrapNodeContent(body.Body, meta, settings)
		if content == "" {
			continue
		}
		writables = append(writables, exportWritable{meta: meta, content: content})
	}
	if len(writables) == 0 {
		return fmt.Errorf("no exportable content")
	}

	tmp, err := os.CreateTemp("", "meguri-export-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	s.exportSave.mu.Lock()
	s.exportSave.tmp = tmpPath
	s.exportSave.mu.Unlock()

	cleanupTmp := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		s.exportSave.mu.Lock()
		if s.exportSave.tmp == tmpPath {
			s.exportSave.tmp = ""
		}
		s.exportSave.mu.Unlock()
	}

	total := len(writables)
	s.emitExportProgress(0, total)

	if req.SplitSave {
		if err := s.writeExportZip(ctx, tmp, writables, settings, total); err != nil {
			cleanupTmp()
			return err
		}
	} else {
		if err := s.writeExportMerged(ctx, tmp, writables, settings, total); err != nil {
			cleanupTmp()
			return err
		}
	}
	if err := tmp.Close(); err != nil {
		cleanupTmp()
		return err
	}
	if err := moveFile(tmpPath, path); err != nil {
		cleanupTmp()
		return err
	}
	s.exportSave.mu.Lock()
	s.exportSave.tmp = ""
	s.exportSave.mu.Unlock()
	s.emitExportProgress(total, total)
	return nil
}

// promptExportPath は保存ダイアログを出す。
func (s *storeService) promptExportPath(split bool, ext string) (string, error) {
	if split {
		path, err := s.app.Dialog.SaveFile().
			SetMessage("Save export ZIP").
			SetFilename("export.zip").
			AddFilter("ZIP archive", "*.zip").
			AddFilter("All Files", "*.*").
			PromptForSingleSelection()
		if err != nil || path == "" {
			return "", err
		}
		if filepath.Ext(path) == "" {
			path += ".zip"
		}
		return path, nil
	}
	filterName := "Markdown"
	filterPattern := "*.md"
	defaultName := "export.md"
	if ext == "html" {
		filterName = "HTML"
		filterPattern = "*.html"
		defaultName = "export.html"
	}
	path, err := s.app.Dialog.SaveFile().
		SetMessage("Save export").
		SetFilename(defaultName).
		AddFilter(filterName, filterPattern).
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()
	if err != nil || path == "" {
		return "", err
	}
	if filepath.Ext(path) == "" {
		path += "." + ext
	}
	return path, nil
}

// emitExportProgress は保存進捗イベントを送る。
func (s *storeService) emitExportProgress(done, total int) {
	if s.app == nil {
		return
	}
	s.app.Event.Emit(topicExportSaveProgress, model.ExportSaveProgressEvent{
		Done:  done,
		Total: total,
	})
}

// writeExportMerged は単一ファイルへ追記する。
func (s *storeService) writeExportMerged(
	ctx context.Context,
	f *os.File,
	items []exportWritable,
	settings exportpreview.MergeSettings,
	total int,
) error {
	sep := exportpreview.ResolveSeparator(settings.Separator, settings.Format)
	for i, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}
		if i > 0 {
			if _, err := f.WriteString(sep); err != nil {
				return err
			}
		}
		if _, err := f.WriteString(item.content); err != nil {
			return err
		}
		s.emitExportProgress(i+1, total)
	}
	return nil
}

// writeExportZip は ZIP へノードごとに書き込む。
func (s *storeService) writeExportZip(
	ctx context.Context,
	f *os.File,
	items []exportWritable,
	settings exportpreview.MergeSettings,
	total int,
) error {
	ext := exportpreview.ExtForFormat(settings.Format)
	bases := make([]string, len(items))
	for i, item := range items {
		bases[i] = exportpreview.BaseNameForMeta(item.meta, settings.HeadingField)
	}
	names := exportpreview.AssignUniqueFileNames(bases, ext)
	zw := zip.NewWriter(f)
	for i, item := range items {
		if err := ctx.Err(); err != nil {
			_ = zw.Close()
			return err
		}
		w, err := zw.Create(names[i])
		if err != nil {
			_ = zw.Close()
			return err
		}
		if _, err := w.Write([]byte(item.content)); err != nil {
			_ = zw.Close()
			return err
		}
		s.emitExportProgress(i+1, total)
	}
	return zw.Close()
}

// moveFile は一時ファイルを確定パスへ移す。
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	return os.Remove(src)
}
