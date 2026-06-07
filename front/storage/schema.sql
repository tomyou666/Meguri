-- scraper-bot UI 永続層スキーマ（SQLite）
-- フロント DTO・ScraperPort・将来 Wails SQLite 実装の単一の正。
-- JSON 列は types/config.ts の AppConfig / PartialConfig 形状に一致させる。
-- 日時列は ISO 8601 UTC 文字列（例: datetime('now')）を格納する。

PRAGMA foreign_keys = ON;

-- ---------------------------------------------------------------------------
-- app_config — アプリ全体のデフォルト設定（1 行のみの singleton）
--   id:            常に 1。INSERT は初回 bootstrap のみ
--   defaults_json: AppConfig 全体（request, content, pdf, crawl, plugins, output）の JSON
--   updated_at:    最終更新日時（ISO 8601）
-- ---------------------------------------------------------------------------
CREATE TABLE app_config (
    id              INTEGER PRIMARY KEY CHECK (id = 1),
    defaults_json   TEXT NOT NULL,
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ---------------------------------------------------------------------------
-- workspaces — ワークスペース（グラフ・設定の単位）
--   id:                      UUID 等の不透明 ID
--   name:                    左 SB 表示名
--   seed_url:                モード 1 の起点 URL（正規化前の入力値を保持可）
--   settings_json:           PartialConfig: request / content / pdf / crawl のみ（plugins・output は含めない）
--   exclude_urls_json:       正規化 URL の JSON 配列。「クロールしない」ノード配下と同期
--   graph_layout_direction:  React Flow / dagre の自動配置向き（LR | TB）
--   baseline_run_id:         Phase4 baseline: crawl_runs.id（循環 FK 回避のため DB 制約なし。存在検証はアプリ層）
--   created_at:              作成日時（ISO 8601）
--   updated_at:              最終更新日時（ISO 8601）
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
-- graph_nodes — グラフノード（ワークスペース内。正規化 URL は WS 内で一意）
--   workspace_id:        所属ワークスペース ID
--   id:                  ノード ID（UI・エッジ参照用。URL とは独立）
--   url_normalized:      normalizeUrl 適用後の URL（重複禁止）
--   label:               グラフ上の短いラベル（通常は URL 表示用）
--   position_x:          キャンバス X 座標
--   position_y:          キャンバス Y 座標
--   user_positioned:     1 = ユーザーがドラッグ配置済み（自動レイアウト対象外）
--   node_settings_json:  PartialConfig: ノード単位の設定上書き
--   crawl_exclude:       1 = このノードと配下をクロールしない
--   origin:              crawl | manual（手動追加ノードは manual）
--   status:              idle | running | success | error | skipped
--   last_error:          直近のノード単位エラー文言（status=error 時など）
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
    origin              TEXT NOT NULL DEFAULT 'crawl' CHECK (origin IN ('crawl', 'manual')),
    status              TEXT NOT NULL DEFAULT 'idle'
                        CHECK (status IN ('idle', 'running', 'success', 'error', 'skipped')),
    last_error          TEXT,
    PRIMARY KEY (workspace_id, id),
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    UNIQUE (workspace_id, url_normalized)
);

CREATE INDEX idx_graph_nodes_workspace_status ON graph_nodes(workspace_id, status);

-- ---------------------------------------------------------------------------
-- graph_edges — グラフエッジ（有向。手動追加・リンク発見の両方）
--   workspace_id:    所属ワークスペース ID
--   id:              エッジ ID
--   source_node_id:  始点ノード ID
--   target_node_id:  終点ノード ID
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
    UNIQUE (workspace_id, source_node_id, target_node_id),
    CHECK (source_node_id != target_node_id)
);

-- ---------------------------------------------------------------------------
-- domain_settings — ドメイン別設定（host 単位の PartialConfig 上書き）
--   workspace_id:  所属ワークスペース ID
--   host:          URL の host 部分（小文字化推奨）
--   settings_json: PartialConfig の JSON
-- ---------------------------------------------------------------------------
CREATE TABLE domain_settings (
    workspace_id    TEXT NOT NULL,
    host            TEXT NOT NULL,
    settings_json   TEXT NOT NULL DEFAULT '{}',
    PRIMARY KEY (workspace_id, host),
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

-- ---------------------------------------------------------------------------
-- crawl_runs — クロール実行履歴（UI runHistory・baseline 参照）
-- WS あたり直近 20 件をアプリ層で保持する想定（DB にはそれ以上残してもよい）
--   id:            実行 ID
--   workspace_id:  所属ワークスペース ID
--   mode:          RunMode: 1=WS 全体, 2=選択ノード（アプリ既定のみ）, 3=既存ノード再クロール
--   status:        running | paused | completed | stopped | error
--   started_at:    開始日時（ISO 8601）
--   finished_at:   終了日時（ISO 8601）
--   summary_json:  CrawlRunSummary 相当の JSON
--   error_message: 実行全体の失敗メッセージ
-- ---------------------------------------------------------------------------
CREATE TABLE crawl_runs (
    id              TEXT PRIMARY KEY,
    workspace_id    TEXT NOT NULL,
    mode            INTEGER NOT NULL CHECK (mode IN (1, 2, 3)),
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
-- node_results — ノード単位のスクレイピング結果（永続層）
-- ノードごとに fetched_at 降順で最大 20 行をアプリ層で保持
-- DeleteResults は「そのノードの最新 1 行のみ」削除
--   id:            結果行 ID
--   run_id:        紐づく crawl_runs.id
--   workspace_id:  所属ワークスペース ID
--   node_id:       グラフノード ID
--   url:           取得時点の URL（表示用）
--   markdown:      抽出 Markdown
--   html:          整形 HTML
--   raw_html:      生 HTML
--   json_body:     JSON レスポンス本文
--   links_json:    抽出リンク URL の JSON 配列（Phase4 links 差分はこの列のみを比較）
--   metadata_json: メタデータ JSON
--   error:         取得失敗時のエラー文言（成功時は NULL）
--   fetched_at:    取得日時（ISO 8601）
--   content_hash:  Phase4 content 差分: canonical markdown の SHA-256 十六進
--                  算法: UTF-8( trim + LF 正規化した markdown ) の SHA-256（front/frontend/src/lib/contentHash.ts）
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
        REFERENCES graph_nodes(workspace_id, id) ON DELETE CASCADE,
    UNIQUE (run_id, node_id)
);

CREATE INDEX idx_node_results_run ON node_results(run_id);
CREATE INDEX idx_node_results_ws_node_fetched
    ON node_results(workspace_id, node_id, fetched_at DESC);

-- ---------------------------------------------------------------------------
-- graph_ui_state — UI 補助状態（ワークスペースごと）
--   workspace_id:            所属ワークスペース ID
--   collapsed_node_ids_json: { "collapsed": string[], "expandedDetail": string[] }
-- ---------------------------------------------------------------------------
CREATE TABLE graph_ui_state (
    workspace_id            TEXT PRIMARY KEY,
    collapsed_node_ids_json TEXT NOT NULL DEFAULT '{"collapsed":[],"expandedDetail":[]}',
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);
