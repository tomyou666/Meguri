import type { PartialConfig } from './config';
import type { CrawlRunSummary } from './crawl';
import type { NodeStatus } from './graph';

export type GraphLayoutDirection = 'LR' | 'TB';

export interface DbAppConfig {
	id: 1;
	defaults_json: string;
	updated_at: string;
}

export interface DbWorkspace {
	id: string;
	name: string;
	seed_url: string;
	settings_json: string;
	exclude_urls_json: string;
	graph_layout_direction: GraphLayoutDirection;
	baseline_run_id: string | null;
	created_at: string;
	updated_at: string;
}

export interface DbGraphNode {
	workspace_id: string;
	id: string;
	url_normalized: string;
	label: string;
	position_x: number;
	position_y: number;
	user_positioned: 0 | 1;
	node_settings_json: string;
	crawl_exclude: 0 | 1;
	status: NodeStatus;
	last_error: string | null;
}

export interface DbGraphEdge {
	workspace_id: string;
	id: string;
	source_node_id: string;
	target_node_id: string;
}

export interface DbCrawlRun {
	id: string;
	workspace_id: string;
	mode: 1 | 2 | 3 | 4;
	status: 'running' | 'paused' | 'completed' | 'stopped' | 'error';
	started_at: string;
	finished_at: string | null;
	summary_json: string | null;
	error_message: string | null;
}

export interface DbNodeResult {
	id: string;
	run_id: string;
	workspace_id: string;
	node_id: string;
	url: string;
	markdown: string | null;
	html: string | null;
	raw_html: string | null;
	json_body: string | null;
	links_json: string | null;
	metadata_json: string | null;
	error: string | null;
	fetched_at: string;
	content_hash: string | null;
	manually_edited: 0 | 1;
}

export interface DbGraphUiState {
	workspace_id: string;
	collapsed_node_ids_json: string;
}

export interface WorkspaceBundle {
	workspace: DbWorkspace;
	nodes: DbGraphNode[];
	edges: DbGraphEdge[];
	uiState: DbGraphUiState | null;
}

export function parseSettingsJson(json: string): PartialConfig {
	return JSON.parse(json) as PartialConfig;
}

export function stringifySettings(config: PartialConfig): string {
	return JSON.stringify(config);
}

export function parseSummaryJson(json: string | null): CrawlRunSummary | null {
	if (!json) return null;
	return JSON.parse(json) as CrawlRunSummary;
}
