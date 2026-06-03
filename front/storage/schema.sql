-- scraper-bot UI 永続層スキーマ（SQLite）
-- フロント DTO・ScraperPort・将来 Wails SQLite 実装の単一の正。
-- JSON 列は types/config.ts の AppConfig / PartialConfig 形状に一致させる。

PRAGMA foreign_keys = ON;

-- ---------------------------------------------------------------------------
-- アプリ全体のデフォルト設定（singleton）
-- defaults_json: AppConfig（request, content, pdf, crawl, plugins, output）
-- ---------------------------------------------------------------------------
CREATE TABLE app_config (
    id              INTEGER PRIMARY KEY CHECK (id = 1),
    defaults_json   TEXT NOT NULL,
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ---------------------------------------------------------------------------
-- ワークスペース
-- settings_json: PartialConfig のうち request, content, pdf, crawl
-- exclude_urls_json: string[]（正規化 URL。ノード「クロールしない」と同期）
-- baseline_run_id: Phase4 差分の「既存」スナップショット（crawl_runs.id）
-- ---------------------------------------------------------------------------
CREATE TABLE workspaces (
    id                      TEXT PRIMARY KEY,
    name                    TEXT NOT NULL,
    seed_url                TEXT NOT NULL,
    settings_json           TEXT NOT NULL DEFAULT '{}',
    exclude_urls_json       TEXT NOT NULL DEFAULT '[]',
    graph_layout_direction  TEXT NOT NULL DEFAULT 'LR'
                            CHECK (graph_layout_direction IN ('LR', 'TB')),
    baseline_run_id         TEXT,
    created_at              TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at              TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_workspaces_updated_at ON workspaces(updated_at);

-- ---------------------------------------------------------------------------
-- グラフノード（ワークスペース内）
-- node_settings_json: PartialConfig（ノード上書き）
-- crawl_exclude: 1 = このノードを「クロールしない」
-- status: idle | running | success | error | skipped
-- ---------------------------------------------------------------------------
CREATE TABLE graph_nodes (
    workspace_id        TEXT NOT NULL,
    id                  TEXT NOT NULL,
    url_normalized      TEXT NOT NULL,
    label               TEXT NOT NULL,
    position_x          REAL NOT NULL,
    position_y          REAL NOT NULL,
    user_positioned     INTEGER NOT NULL DEFAULT 0 CHECK (user_positioned IN (0, 1)),
    node_settings_json  TEXT NOT NULL DEFAULT '{}',
    crawl_exclude       INTEGER NOT NULL DEFAULT 0 CHECK (crawl_exclude IN (0, 1)),
    status              TEXT NOT NULL DEFAULT 'idle'
                        CHECK (status IN ('idle', 'running', 'success', 'error', 'skipped')),
    last_error          TEXT,
    PRIMARY KEY (workspace_id, id),
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    UNIQUE (workspace_id, url_normalized)
);

CREATE INDEX idx_graph_nodes_workspace_status ON graph_nodes(workspace_id, status);

-- ---------------------------------------------------------------------------
-- グラフエッジ（有向）
-- 制約: 同一 (workspace_id, source_node_id, target_node_id) は1本
-- ---------------------------------------------------------------------------
CREATE TABLE graph_edges (
    workspace_id    TEXT NOT NULL,
    id              TEXT NOT NULL,
    source_node_id  TEXT NOT NULL,
    target_node_id  TEXT NOT NULL,
    PRIMARY KEY (workspace_id, id),
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    FOREIGN KEY (workspace_id, source_node_id)
        REFERENCES graph_nodes(workspace_id, id) ON DELETE CASCADE,
    FOREIGN KEY (workspace_id, target_node_id)
        REFERENCES graph_nodes(workspace_id, id) ON DELETE CASCADE,
    UNIQUE (workspace_id, source_node_id, target_node_id)
);

-- ---------------------------------------------------------------------------
-- ドメイン別設定（host 単位）
-- settings_json: PartialConfig
-- ---------------------------------------------------------------------------
CREATE TABLE domain_settings (
    workspace_id    TEXT NOT NULL,
    host            TEXT NOT NULL,
    settings_json   TEXT NOT NULL DEFAULT '{}',
    PRIMARY KEY (workspace_id, host),
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

-- ---------------------------------------------------------------------------
-- クロール実行履歴
-- mode: 1 | 2 | 3（RunMode）
-- status: running | paused | completed | stopped | error
-- summary_json: CrawlRunSummary 相当
-- ---------------------------------------------------------------------------
CREATE TABLE crawl_runs (
    id              TEXT PRIMARY KEY,
    workspace_id    TEXT NOT NULL,
    mode            INTEGER NOT NULL CHECK (mode IN (1,  2, 3)),
    status          TEXT NOT NULL
                    CHECK (status IN ('running', 'paused', 'completed', 'stopped', 'error')),
    started_at      TEXT NOT NULL,
    finished_at     TEXT,
    summary_json    TEXT,
    error_message   TEXT,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

CREATE INDEX idx_crawl_runs_workspace_started ON crawl_runs(workspace_id, started_at DESC);

-- ---------------------------------------------------------------------------
-- ノード単位のスクレイピング結果（永続層）
-- content_hash: Phase4 本文差分用
-- ---------------------------------------------------------------------------
CREATE TABLE node_results (
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
    FOREIGN KEY (run_id) REFERENCES crawl_runs(id) ON DELETE CASCADE,
    FOREIGN KEY (workspace_id, node_id)
        REFERENCES graph_nodes(workspace_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_node_results_run ON node_results(run_id);
CREATE INDEX idx_node_results_ws_node_fetched
    ON node_results(workspace_id, node_id, fetched_at DESC);

-- ---------------------------------------------------------------------------
-- UI 補助: 折りたたみ状態（ワークスペースごと）
-- ---------------------------------------------------------------------------
CREATE TABLE graph_ui_state (
    workspace_id            TEXT PRIMARY KEY,
    collapsed_node_ids_json TEXT NOT NULL DEFAULT '[]',
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);
