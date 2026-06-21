-- RunMode 4（選択ノードのみ取得）を crawl_runs.mode に許可する。
PRAGMA foreign_keys = OFF;

CREATE TABLE crawl_runs_new (
    id              TEXT PRIMARY KEY,
    workspace_id    TEXT NOT NULL,
    mode            INTEGER NOT NULL CHECK (mode IN (1, 2, 3, 4)),
    status          TEXT NOT NULL
                    CHECK (status IN ('running', 'paused', 'completed', 'stopped', 'error')),
    started_at      TEXT NOT NULL,
    finished_at     TEXT,
    summary_json    TEXT,
    error_message   TEXT,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

INSERT INTO crawl_runs_new
SELECT id, workspace_id, mode, status, started_at, finished_at, summary_json, error_message
FROM crawl_runs;

DROP TABLE crawl_runs;

ALTER TABLE crawl_runs_new RENAME TO crawl_runs;

CREATE INDEX idx_crawl_runs_workspace_started ON crawl_runs(workspace_id, started_at DESC);

PRAGMA foreign_keys = ON;
