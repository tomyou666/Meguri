import type { DbGraphEdge, DbGraphNode, WorkspaceBundle } from '@/types/db';
import { parseSettingsJson, stringifySettings } from '@/types/db';
import type { GraphEdge, GraphNode } from '@/types/graph';
import type { Workspace } from '@/types/workspace';

export function workspaceToDb(ws: Workspace): WorkspaceBundle {
	return {
		workspace: {
			id: ws.id,
			name: ws.name,
			seed_url: ws.seedUrl,
			settings_json: stringifySettings(ws.settings),
			exclude_urls_json: JSON.stringify(ws.exclude_urls),
			graph_layout_direction: ws.graphLayoutDirection,
			baseline_run_id: ws.baselineRunId ?? null,
			created_at: ws.createdAt ?? new Date().toISOString(),
			updated_at: new Date().toISOString(),
		},
		nodes: ws.nodes.map((n) => graphNodeToDb(ws.id, n)),
		edges: ws.edges.map((e) => graphEdgeToDb(ws.id, e)),
		uiState: {
			workspace_id: ws.id,
			collapsed_node_ids_json: JSON.stringify({
				collapsed: ws.collapsedNodeIds ?? [],
				expandedDetail: ws.expandedDetailNodeIds ?? [],
			}),
		},
	};
}

export function workspaceFromDb(
	bundle: WorkspaceBundle,
	hydratedResults?: Map<string, GraphNode['lastResult']>,
): Workspace {
	const settings = parseSettingsJson(bundle.workspace.settings_json);
	const exclude_urls = JSON.parse(
		bundle.workspace.exclude_urls_json,
	) as string[];
	let collapsedNodeIds: string[] = [];
	let expandedDetailNodeIds: string[] = [];
	if (bundle.uiState) {
		const parsed = JSON.parse(bundle.uiState.collapsed_node_ids_json) as
			| string[]
			| { collapsed?: string[]; expandedDetail?: string[] };
		if (Array.isArray(parsed)) {
			collapsedNodeIds = parsed;
		} else {
			collapsedNodeIds = parsed.collapsed ?? [];
			expandedDetailNodeIds = parsed.expandedDetail ?? [];
		}
	}

	return {
		id: bundle.workspace.id,
		name: bundle.workspace.name,
		seedUrl: bundle.workspace.seed_url,
		settings,
		exclude_urls,
		nodes: bundle.nodes.map((n) =>
			graphNodeFromDb(n, hydratedResults?.get(n.id)),
		),
		edges: bundle.edges.map(graphEdgeFromDb),
		graphLayoutDirection: bundle.workspace.graph_layout_direction,
		baselineRunId: bundle.workspace.baseline_run_id ?? undefined,
		collapsedNodeIds,
		expandedDetailNodeIds,
		createdAt: bundle.workspace.created_at,
	};
}

function graphNodeToDb(workspaceId: string, n: GraphNode): DbGraphNode {
	return {
		workspace_id: workspaceId,
		id: n.id,
		url_normalized: n.urlNormalized,
		label: n.label,
		position_x: n.position.x,
		position_y: n.position.y,
		user_positioned: n.userPositioned ? 1 : 0,
		node_settings_json: stringifySettings(n.nodeSettings),
		crawl_exclude: n.crawlExclude ? 1 : 0,
		status: n.status,
		last_error: n.lastError ?? null,
	};
}

function graphNodeFromDb(
	row: DbGraphNode,
	lastResult?: GraphNode['lastResult'],
): GraphNode {
	return {
		id: row.id,
		urlNormalized: row.url_normalized,
		label: row.label,
		position: { x: row.position_x, y: row.position_y },
		userPositioned: row.user_positioned === 1,
		nodeSettings: parseSettingsJson(row.node_settings_json),
		crawlExclude: row.crawl_exclude === 1,
		status: row.status,
		lastError: row.last_error ?? undefined,
		lastResult,
	};
}

function graphEdgeToDb(workspaceId: string, e: GraphEdge): DbGraphEdge {
	return {
		workspace_id: workspaceId,
		id: e.id,
		source_node_id: e.source,
		target_node_id: e.target,
	};
}

function graphEdgeFromDb(row: DbGraphEdge): GraphEdge {
	return {
		id: row.id,
		source: row.source_node_id,
		target: row.target_node_id,
	};
}
