import type { PartialConfig } from '@/types/config';
import type { CrawlResultPreview } from '@/types/crawl';
import type { GraphEdge, GraphNode } from '@/types/graph';
import type { Workspace } from '@/types/workspace';
import type {
	CrawlResultDTO,
	WorkspaceDTO,
} from '../../bindings/meguri-app/internal/model/models.js';

export function workspaceFromDTO(dto: WorkspaceDTO): Workspace {
	const settings = parseJSON<PartialConfig>(dto.settings, {});
	return {
		id: dto.id,
		name: dto.name,
		seedUrl: dto.seedUrl,
		settings,
		exclude_urls: dto.exclude_urls ?? [],
		nodes: (dto.nodes ?? []).map(nodeFromDTO),
		edges: (dto.edges ?? []).map(edgeFromDTO),
		graphLayoutDirection:
			(dto.graphLayoutDirection as Workspace['graphLayoutDirection']) ?? 'LR',
		baselineRunId: dto.baselineRunId || undefined,
		collapsedNodeIds: dto.collapsedNodeIds ?? [],
		expandedDetailNodeIds: dto.expandedDetailNodeIds ?? [],
		createdAt: dto.createdAt,
	};
}

export function workspaceToDTO(ws: Workspace): WorkspaceDTO {
	return {
		id: ws.id,
		name: ws.name,
		seedUrl: ws.seedUrl,
		settings: ws.settings,
		exclude_urls: ws.exclude_urls,
		nodes: ws.nodes.map(nodeToDTO),
		edges: ws.edges.map(edgeToDTO),
		graphLayoutDirection: ws.graphLayoutDirection,
		baselineRunId: ws.baselineRunId ?? '',
		collapsedNodeIds: ws.collapsedNodeIds ?? [],
		expandedDetailNodeIds: ws.expandedDetailNodeIds ?? [],
		createdAt: ws.createdAt ?? new Date().toISOString(),
	};
}

function nodeFromDTO(n: WorkspaceDTO['nodes'][0]): GraphNode {
	return {
		id: n.id,
		urlNormalized: n.urlNormalized,
		label: n.label,
		position: { x: n.position?.x ?? 0, y: n.position?.y ?? 0 },
		userPositioned: n.userPositioned,
		nodeSettings: parseJSON<PartialConfig>(n.nodeSettings, {}),
		crawlExclude: n.crawlExclude,
		origin: (n.origin === 'manual' ? 'manual' : 'crawl') as GraphNode['origin'],
		status: n.status as GraphNode['status'],
		lastError: n.lastError,
		lastResult: n.lastResult ? crawlResultFromDTO(n.lastResult) : undefined,
	};
}

function nodeToDTO(n: GraphNode): WorkspaceDTO['nodes'][0] {
	return {
		id: n.id,
		urlNormalized: n.urlNormalized,
		label: n.label,
		position: { x: n.position.x, y: n.position.y },
		userPositioned: n.userPositioned,
		nodeSettings: n.nodeSettings,
		crawlExclude: n.crawlExclude,
		origin: n.origin ?? 'crawl',
		status: n.status,
		lastError: n.lastError,
		lastResult: n.lastResult ? crawlResultToDTO(n.lastResult) : undefined,
	};
}

function crawlResultToDTO(r: CrawlResultPreview): CrawlResultDTO {
	return {
		url: r.url,
		markdown: r.markdown,
		html: r.html,
		rawHtml: r.raw_html,
		jsonBody: r.json,
		links: r.links,
		metadata: r.metadata,
		manuallyEdited: r.manuallyEdited,
	};
}

function edgeFromDTO(e: WorkspaceDTO['edges'][0]): GraphEdge {
	return { id: e.id, source: e.source, target: e.target };
}

function edgeToDTO(e: GraphEdge): WorkspaceDTO['edges'][0] {
	return { id: e.id, source: e.source, target: e.target };
}

function crawlResultFromDTO(dto: CrawlResultDTO): CrawlResultPreview {
	return {
		url: dto.url,
		markdown: dto.markdown,
		html: dto.html,
		raw_html: dto.rawHtml,
		json: dto.jsonBody,
		links: dto.links,
		metadata: dto.metadata,
		manuallyEdited: dto.manuallyEdited,
	};
}

export { crawlResultFromDTO, crawlResultToDTO };

function parseJSON<T>(raw: unknown, fallback: T): T {
	if (raw == null || raw === '') return fallback;
	if (typeof raw === 'object') return raw as T;
	try {
		return JSON.parse(String(raw)) as T;
	} catch {
		return fallback;
	}
}

export function partialConfigToRaw(config: PartialConfig): PartialConfig {
	return config;
}
