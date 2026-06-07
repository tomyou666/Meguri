-- SQLite 3.35+ supports DROP COLUMN; recreate table if unsupported on target platform.
ALTER TABLE graph_nodes DROP COLUMN origin;
