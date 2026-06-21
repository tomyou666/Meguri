package scrb

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"scraperbot-front/internal/model"
)

const formatVersion = 1

type manifest struct {
	FormatVersion int    `json:"formatVersion"`
	ExportedAt    string `json:"exportedAt"`
	App           string `json:"app"`
	WorkspaceName string `json:"workspaceName"`
}

// Export は WorkspaceBundle を .scrb ZIP バイト列にエンコードする。
func Export(bundle model.WorkspaceBundle) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	m := manifest{
		FormatVersion: formatVersion,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		App:           "scraper-bot",
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
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Import は .scrb ZIP から WorkspaceBundle をデコードする。
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
	if m.FormatVersion != formatVersion {
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
	return bundle, nil
}

func writeJSON(w *zip.Writer, name string, v any) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
