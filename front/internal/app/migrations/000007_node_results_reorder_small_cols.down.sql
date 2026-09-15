-- 000007 の巻き戻し: content_hash / manually_edited を巨大 TEXT の後ろへ戻す。
PRAGMA foreign_keys = OFF;

CREATE TABLE node_results_new (
    id              TEXT PRIMARY KEY,
    run_id          TEXT NOT NULL,
    workspace_id    TEXT NOT NULL,
    node_id         TEXT NOT NULL,
    url             TEXT NOT NULL,
    markdown        TEXT,
    html            TEXT,
    raw_html        TEXT,
    json_body       TEXT,
    links_json      TEXT,
    metadata_json   TEXT,
    error           TEXT,
    fetched_at      TEXT NOT NULL,
    content_hash    TEXT,
    manually_edited INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (workspace_id, node_id)
        REFERENCES graph_nodes(workspace_id, id) ON DELETE CASCADE,
    UNIQUE (run_id, node_id)
);

INSERT INTO node_results_new (
    id, run_id, workspace_id, node_id, url,
    markdown, html, raw_html, json_body,
    links_json, metadata_json, error, fetched_at,
    content_hash, manually_edited
)
SELECT
    id, run_id, workspace_id, node_id, url,
    markdown, html, raw_html, json_body,
    links_json, metadata_json, error, fetched_at,
    content_hash, manually_edited
FROM node_results;

DROP TABLE node_results;

ALTER TABLE node_results_new RENAME TO node_results;

CREATE INDEX idx_node_results_run ON node_results(run_id);
CREATE INDEX idx_node_results_ws_node_fetched
    ON node_results(workspace_id, node_id, fetched_at DESC);

PRAGMA foreign_keys = ON;
