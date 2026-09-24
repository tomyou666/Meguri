-- node_results から本文列を node_result_bodies へ分離する。
-- 本文はノードあたり latest 成功 + baseline 行のみ残す（同一なら 1 件）。
-- links_hash は Go バックフィルで埋める（SQLite に SHA-256 がないため）。
PRAGMA foreign_keys = OFF;

-- 一時的に FK なしで本文を退避する（node_results rebuild 後に FK 付きへ作り直す）。
CREATE TABLE node_result_bodies_tmp (
    id              TEXT PRIMARY KEY,
    links_json      TEXT,
    metadata_json   TEXT,
    markdown        TEXT,
    html            TEXT,
    raw_html        TEXT,
    json_body       TEXT
);

INSERT INTO node_result_bodies_tmp (
    id, links_json, metadata_json, markdown, html, raw_html, json_body
)
SELECT
    nr.id, nr.links_json, nr.metadata_json, nr.markdown, nr.html, nr.raw_html, nr.json_body
FROM node_results nr
WHERE nr.id IN (
    SELECT id FROM (
        SELECT
            id,
            ROW_NUMBER() OVER (
                PARTITION BY workspace_id, node_id
                ORDER BY fetched_at DESC
            ) AS rn
        FROM node_results
        WHERE error IS NULL OR error = ''
    ) ranked
    WHERE rn = 1
)
OR nr.id IN (
    SELECT nr2.id
    FROM node_results nr2
    INNER JOIN workspaces w
        ON w.id = nr2.workspace_id
       AND w.baseline_run_id IS NOT NULL
       AND w.baseline_run_id != ''
       AND nr2.run_id = w.baseline_run_id
);

CREATE TABLE node_results_new (
    id              TEXT PRIMARY KEY,
    run_id          TEXT NOT NULL,
    workspace_id    TEXT NOT NULL,
    node_id         TEXT NOT NULL,
    url             TEXT NOT NULL,
    content_hash    TEXT,
    links_hash      TEXT,
    manually_edited INTEGER NOT NULL DEFAULT 0,
    error           TEXT,
    fetched_at      TEXT NOT NULL,
    FOREIGN KEY (workspace_id, node_id)
        REFERENCES graph_nodes(workspace_id, id) ON DELETE CASCADE,
    UNIQUE (run_id, node_id)
);

INSERT INTO node_results_new (
    id, run_id, workspace_id, node_id, url,
    content_hash, links_hash, manually_edited, error, fetched_at
)
SELECT
    id, run_id, workspace_id, node_id, url,
    content_hash, NULL, manually_edited, error, fetched_at
FROM node_results;

DROP TABLE node_results;
ALTER TABLE node_results_new RENAME TO node_results;

CREATE TABLE node_result_bodies (
    id              TEXT PRIMARY KEY,
    links_json      TEXT,
    metadata_json   TEXT,
    markdown        TEXT,
    html            TEXT,
    raw_html        TEXT,
    json_body       TEXT,
    FOREIGN KEY (id) REFERENCES node_results(id) ON DELETE CASCADE
);

INSERT INTO node_result_bodies
SELECT * FROM node_result_bodies_tmp;

DROP TABLE node_result_bodies_tmp;

CREATE INDEX idx_node_results_run ON node_results(run_id);
CREATE INDEX idx_node_results_ws_node_fetched
    ON node_results(workspace_id, node_id, fetched_at DESC);

PRAGMA foreign_keys = ON;
