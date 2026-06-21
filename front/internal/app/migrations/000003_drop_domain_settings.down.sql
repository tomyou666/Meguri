CREATE TABLE domain_settings (
    workspace_id    TEXT NOT NULL,
    host            TEXT NOT NULL,
    settings_json   TEXT NOT NULL DEFAULT '{}',
    PRIMARY KEY (workspace_id, host),
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);
