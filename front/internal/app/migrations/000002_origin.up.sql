ALTER TABLE graph_nodes ADD COLUMN origin TEXT NOT NULL DEFAULT 'crawl' CHECK (origin IN ('crawl', 'manual'));
