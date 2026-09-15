-- node_results の content_hash / manually_edited を巨大 TEXT より前へ移す。
-- SQLite は packed record のため、後段カラムの SELECT は手前の markdown/html の
-- overflow page を辿る。メタデータ列を前に置くと overflow を読まずに済む。
PRAGMA foreign_keys = OFF;

CREATE TABLE node_results_new (
    id              TEXT PRIMARY KEY,
    run_id          TEXT NOT NULL,
    workspace_id    TEXT NOT NULL,
    node_id         TEXT NOT NULL,
    url             TEXT NOT NULL,
    content_hash    TEXT,
    manually_edited INTEGER NOT NULL DEFAULT 0,
    markdown        TEXT,
    html            TEXT,
    raw_html        TEXT,
    json_body       TEXT,
    links_json      TEXT,
    metadata_json   TEXT,
    error           TEXT,
    fetched_at      TEXT NOT NULL,
    FOREIGN KEY (workspace_id, node_id)
        REFERENCES graph_nodes(workspace_id, id) ON DELETE CASCADE,
    UNIQUE (run_id, node_id)
);

INSERT INTO node_results_new (
    id, run_id, workspace_id, node_id, url,
    content_hash, manually_edited,
    markdown, html, raw_html, json_body,
    links_json, metadata_json, error, fetched_at
)
SELECT
    id, run_id, workspace_id, node_id, url,
    content_hash, manually_edited,
    markdown, html, raw_html, json_body,
    links_json, metadata_json, error, fetched_at
FROM node_results;

DROP TABLE node_results;

ALTER TABLE node_results_new RENAME TO node_results;

CREATE INDEX idx_node_results_run ON node_results(run_id);
CREATE INDEX idx_node_results_ws_node_fetched
    ON node_results(workspace_id, node_id, fetched_at DESC);

PRAGMA foreign_keys = ON;
