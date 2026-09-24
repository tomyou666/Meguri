package scrb

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"meguri-app/internal/model"
)

// formatVersionV1 は結合 results.json の旧形式。
const formatVersionV1 = 1

// formatVersionV2 は results + result_bodies の現行形式。
const formatVersionV2 = 2

type manifest struct {
	// FormatVersion は .scrb 形式番号。
	FormatVersion int `json:"formatVersion"`
	// ExportedAt はエクスポート日時（ISO 8601）。
	ExportedAt string `json:"exportedAt"`
	// App は出力元アプリ名。
	App string `json:"app"`
	// WorkspaceName はワークスペース名。
	WorkspaceName string `json:"workspaceName"`
}

// legacyResultV1 は formatVersion 1 の結合 results.json 行。
type legacyResultV1 struct {
	// ID は結果行 ID。
	ID *string `json:"id"`
	// RunID は crawl run ID。
	RunID string `json:"run_id"`
	// WorkspaceID は所属ワークスペース ID。
	WorkspaceID string `json:"workspace_id"`
	// NodeID はグラフノード ID。
	NodeID string `json:"node_id"`
	// URL は取得時点の URL。
	URL string `json:"url"`
	// Markdown は抽出 Markdown。
	Markdown *string `json:"markdown"`
	// HTML は整形 HTML。
	HTML *string `json:"html"`
	// RawHTML は生 HTML。
	RawHTML *string `json:"raw_html"`
	// JSONBody は JSON 本文。
	JSONBody *string `json:"json_body"`
	// LinksJSON は抽出リンクの JSON 配列。
	LinksJSON *string `json:"links_json"`
	// MetadataJSON はメタデータ JSON。
	MetadataJSON *string `json:"metadata_json"`
	// Error は失敗時の文言。
	Error *string `json:"error"`
	// FetchedAt は取得日時（ISO 8601）。
	FetchedAt string `json:"fetched_at"`
	// ContentHash は canonical markdown の SHA-256 十六進。
	ContentHash *string `json:"content_hash"`
	// ManuallyEdited は手動編集済みなら 1。
	ManuallyEdited int32 `json:"manually_edited"`
}

// Export は WorkspaceBundle を .scrb ZIP バイト列にエンコードする（formatVersion 2）。
func Export(bundle model.WorkspaceBundle) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	m := manifest{
		FormatVersion: formatVersionV2,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		App:           "meguri",
		WorkspaceName: bundle.Workspace.Name,
	}
	if err := writeJSON(w, "manifest.json", m); err != nil {
		return nil, err
	}
	if err := writeJSON(w, "workspace.json", bundle.Workspace); err != nil {
		return nil, err
	}
	if err := writeJSON(w, "nodes.json", bundle.Nodes); err != nil {
		return nil, err
	}
	if err := writeJSON(w, "edges.json", bundle.Edges); err != nil {
		return nil, err
	}
	ui := bundle.UIState
	if ui == nil {
		ui = &model.GraphUIState{CollapsedNodeIdsJSON: `{"collapsed":[],"expandedDetail":[]}`}
	}
	if err := writeJSON(w, "ui_state.json", ui); err != nil {
		return nil, err
	}
	if len(bundle.Results) > 0 {
		if err := writeJSON(w, "results.json", bundle.Results); err != nil {
			return nil, err
		}
	}
	if len(bundle.ResultBodies) > 0 {
		if err := writeJSON(w, "result_bodies.json", bundle.ResultBodies); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Import は .scrb ZIP から WorkspaceBundle をデコードする。
//
// formatVersion 1（結合 results.json）と 2（results + result_bodies）を受理する。
func Import(data []byte) (model.WorkspaceBundle, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return model.WorkspaceBundle{}, fmt.Errorf("invalid zip: %w", err)
	}
	files := map[string][]byte{}
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return model.WorkspaceBundle{}, err
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return model.WorkspaceBundle{}, err
		}
		files[f.Name] = b
	}
	var m manifest
	if err := json.Unmarshal(files["manifest.json"], &m); err != nil {
		return model.WorkspaceBundle{}, fmt.Errorf("manifest: %w", err)
	}
	if m.FormatVersion != formatVersionV1 && m.FormatVersion != formatVersionV2 {
		return model.WorkspaceBundle{}, fmt.Errorf("unsupported formatVersion: %d", m.FormatVersion)
	}
	var bundle model.WorkspaceBundle
	if err := json.Unmarshal(files["workspace.json"], &bundle.Workspace); err != nil {
		return model.WorkspaceBundle{}, err
	}
	if err := json.Unmarshal(files["nodes.json"], &bundle.Nodes); err != nil {
		return model.WorkspaceBundle{}, err
	}
	if err := json.Unmarshal(files["edges.json"], &bundle.Edges); err != nil {
		return model.WorkspaceBundle{}, err
	}
	if b, ok := files["ui_state.json"]; ok {
		var ui model.GraphUIState
		if err := json.Unmarshal(b, &ui); err != nil {
			return model.WorkspaceBundle{}, err
		}
		bundle.UIState = &ui
	}
	if b, ok := files["results.json"]; ok {
		if m.FormatVersion == formatVersionV1 {
			metas, bodies, err := splitLegacyResults(b)
			if err != nil {
				return model.WorkspaceBundle{}, err
			}
			bundle.Results = metas
			bundle.ResultBodies = bodies
		} else {
			var results []model.NodeResult
			if err := json.Unmarshal(b, &results); err != nil {
				return model.WorkspaceBundle{}, fmt.Errorf("results: %w", err)
			}
			bundle.Results = results
			if bb, ok := files["result_bodies.json"]; ok {
				var bodies []model.NodeResultBody
				if err := json.Unmarshal(bb, &bodies); err != nil {
					return model.WorkspaceBundle{}, fmt.Errorf("result_bodies: %w", err)
				}
				bundle.ResultBodies = bodies
			}
		}
	}
	return bundle, nil
}

// splitLegacyResults は v1 結合 results.json をメタと本文に分解する。
func splitLegacyResults(data []byte) ([]model.NodeResult, []model.NodeResultBody, error) {
	var legacy []legacyResultV1
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, nil, fmt.Errorf("results: %w", err)
	}
	metas := make([]model.NodeResult, 0, len(legacy))
	bodies := make([]model.NodeResultBody, 0)
	for _, row := range legacy {
		meta := model.NodeResult{
			ID:             row.ID,
			RunID:          row.RunID,
			WorkspaceID:    row.WorkspaceID,
			NodeID:         row.NodeID,
			URL:            row.URL,
			ContentHash:    row.ContentHash,
			ManuallyEdited: row.ManuallyEdited,
			Error:          row.Error,
			FetchedAt:      row.FetchedAt,
		}
		metas = append(metas, meta)
		isSuccess := row.Error == nil || *row.Error == ""
		if !isSuccess || row.ID == nil {
			continue
		}
		hasBody := row.Markdown != nil || row.HTML != nil || row.RawHTML != nil ||
			row.JSONBody != nil || row.LinksJSON != nil || row.MetadataJSON != nil
		if !hasBody {
			continue
		}
		bodies = append(bodies, model.NodeResultBody{
			ID:           row.ID,
			LinksJSON:    row.LinksJSON,
			MetadataJSON: row.MetadataJSON,
			Markdown:     row.Markdown,
			HTML:         row.HTML,
			RawHTML:      row.RawHTML,
			JSONBody:     row.JSONBody,
		})
	}
	return metas, bodies, nil
}

// writeJSON は ZIP 内にインデント付き JSON ファイルを書く。
func writeJSON(w *zip.Writer, name string, v any) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
