import { workspaceFromDb, workspaceToDb } from '@/lib/dbMappers';
import { DEFAULT_APP_CONFIG } from '@/lib/defaults';
import { runCrawlStub } from '@/services/crawlStub';
import type {
	MergeResultsResponse,
	SaveSettingsResponse,
	ScraperPort,
	StartCrawlParams,
	WorkspaceDiff,
} from '@/types/adapter';
import type { PartialConfig } from '@/types/config';
import type { CrawlResultPreview } from '@/types/crawl';
import type { DbNodeResult } from '@/types/db';
import type { Workspace } from '@/types/workspace';

function uid(): string {
	return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

function contentHash(text: string): string {
	let h = 0;
	for (let i = 0; i < text.length; i++) {
		h = (Math.imul(31, h) + text.charCodeAt(i)) | 0;
	}
	return `h${(h >>> 0).toString(16)}`;
}

function resultToDb(
	runId: string,
	workspaceId: string,
	nodeId: string,
	result: CrawlResultPreview,
): DbNodeResult {
	const markdown = result.markdown ?? null;
	return {
		id: uid(),
		run_id: runId,
		workspace_id: workspaceId,
		node_id: nodeId,
		url: result.url,
		markdown,
		html: null,
		raw_html: null,
		json_body: null,
		links_json: result.links ? JSON.stringify(result.links) : null,
		metadata_json: result.metadata ? JSON.stringify(result.metadata) : null,
		error: null,
		fetched_at: new Date().toISOString(),
		content_hash: markdown ? contentHash(markdown) : null,
	};
}

function dbToPreview(row: DbNodeResult): CrawlResultPreview {
	return {
		url: row.url,
		markdown: row.markdown ?? undefined,
		links: row.links_json
			? (JSON.parse(row.links_json) as string[])
			: undefined,
		metadata: row.metadata_json
			? (JSON.parse(row.metadata_json) as Record<string, string>)
			: undefined,
	};
}

export class MockScraperAdapter implements ScraperPort {
	private appDefaults: PartialConfig = { ...DEFAULT_APP_CONFIG };
	private workspaces = new Map<string, Workspace>();
	private results = new Map<string, DbNodeResult[]>();
	private baselineResults = new Map<string, Map<string, DbNodeResult>>();
	private runs = new Map<string, string>();

	async getAppDefaults(): Promise<PartialConfig> {
		return structuredClone(this.appDefaults);
	}

	async setAppDefaults(config: PartialConfig): Promise<void> {
		this.appDefaults = structuredClone(config);
	}

	async saveAppDefaults(config: PartialConfig): Promise<SaveSettingsResponse> {
		await this.setAppDefaults(config);
		return { ok: true, scope: 'app' };
	}

	async saveWorkspaceSettings(
		workspaceId: string,
		settings: PartialConfig,
	): Promise<SaveSettingsResponse> {
		const ws = this.workspaces.get(workspaceId);
		if (!ws) throw new Error('Workspace not found');
		ws.settings = { ...ws.settings, ...structuredClone(settings) };
		await this.saveWorkspace(ws);
		return { ok: true, scope: 'workspace' };
	}

	async saveDomainSettings(
		workspaceId: string,
		domain: string,
		settings: PartialConfig,
	): Promise<SaveSettingsResponse> {
		const ws = this.workspaces.get(workspaceId);
		if (!ws) throw new Error('Workspace not found');
		ws.domainSettings = {
			...ws.domainSettings,
			[domain]: { ...ws.domainSettings[domain], ...structuredClone(settings) },
		};
		await this.saveWorkspace(ws);
		return { ok: true, scope: 'domain' };
	}

	async saveNodeSettings(
		workspaceId: string,
		nodeId: string,
		settings: PartialConfig,
	): Promise<SaveSettingsResponse> {
		const ws = this.workspaces.get(workspaceId);
		if (!ws) throw new Error('Workspace not found');
		const node = ws.nodes.find((n) => n.id === nodeId);
		if (!node) throw new Error('Node not found');
		node.nodeSettings = { ...node.nodeSettings, ...structuredClone(settings) };
		await this.saveWorkspace(ws);
		return { ok: true, scope: 'node' };
	}

	async loadWorkspace(id: string): Promise<Workspace | null> {
		const ws = this.workspaces.get(id);
		if (!ws) return null;
		return this.hydrateWorkspace(ws);
	}

	private hydrateWorkspace(ws: Workspace): Workspace {
		const hydrated = new Map<string, CrawlResultPreview>();
		const rows = this.results.get(ws.id) ?? [];
		const latestByNode = new Map<string, DbNodeResult>();
		for (const row of rows) {
			const prev = latestByNode.get(row.node_id);
			if (!prev || row.fetched_at > prev.fetched_at) {
				latestByNode.set(row.node_id, row);
			}
		}
		for (const [nodeId, row] of latestByNode) {
			hydrated.set(nodeId, dbToPreview(row));
		}
		const bundle = workspaceToDb(ws);
		return workspaceFromDb(bundle, hydrated);
	}

	async saveWorkspace(ws: Workspace): Promise<void> {
		this.workspaces.set(ws.id, structuredClone(ws));
	}

	async duplicateWorkspace(id: string): Promise<Workspace> {
		const src = this.workspaces.get(id);
		if (!src) throw new Error('Workspace not found');
		const newId = uid();
		const idMap = new Map<string, string>();
		for (const n of src.nodes) {
			idMap.set(n.id, uid());
		}
		const copy: Workspace = {
			...structuredClone(src),
			id: newId,
			name: `${src.name} (copy)`,
			baselineRunId: undefined,
			createdAt: new Date().toISOString(),
			nodes: src.nodes.map((n) => ({
				...n,
				id: idMap.get(n.id)!,
				status: 'idle',
				lastResult: undefined,
				lastError: undefined,
			})),
			edges: src.edges.map((e) => ({
				id: `e-${idMap.get(e.source)}-${idMap.get(e.target)}`,
				source: idMap.get(e.source)!,
				target: idMap.get(e.target)!,
			})),
		};
		this.workspaces.set(newId, copy);
		this.results.set(newId, []);
		return copy;
	}

	async getNodeResult(
		workspaceId: string,
		nodeId: string,
	): Promise<CrawlResultPreview | null> {
		const rows = this.results.get(workspaceId) ?? [];
		const nodeRows = rows
			.filter((r) => r.node_id === nodeId)
			.sort((a, b) => b.fetched_at.localeCompare(a.fetched_at));
		return nodeRows[0] ? dbToPreview(nodeRows[0]) : null;
	}

	async getNodeResults(
		workspaceId: string,
		nodeIds: string[],
	): Promise<CrawlResultPreview[]> {
		const out: CrawlResultPreview[] = [];
		for (const nodeId of nodeIds) {
			const r = await this.getNodeResult(workspaceId, nodeId);
			if (r) out.push(r);
		}
		return out;
	}

	async mergeResults(
		workspaceId: string,
		nodeIds: string[] | null,
		formats: string[] = ['markdown'],
	): Promise<MergeResultsResponse> {
		const ws = this.workspaces.get(workspaceId);
		if (!ws) throw new Error('Workspace not found');
		const ids =
			nodeIds ??
			ws.nodes.filter((n) => n.status === 'success').map((n) => n.id);
		const previews = await this.getNodeResults(workspaceId, ids);
		const parts: string[] = [];
		for (const p of previews) {
			if (formats.includes('markdown') && p.markdown) {
				parts.push(`## ${p.url}\n\n${p.markdown}`);
			}
		}
		return {
			merged: parts.join('\n\n---\n\n'),
			format: 'markdown',
			nodeCount: previews.length,
		};
	}

	async saveResults(workspaceId: string, nodeIds: string[]): Promise<void> {
		const rows = this.results.get(workspaceId) ?? [];
		const snapshot = new Map<string, DbNodeResult>();
		for (const nodeId of nodeIds) {
			const latest = rows
				.filter((r) => r.node_id === nodeId)
				.sort((a, b) => b.fetched_at.localeCompare(a.fetched_at))[0];
			if (latest) snapshot.set(nodeId, { ...latest });
		}
		this.baselineResults.set(workspaceId, snapshot);
	}

	async deleteResults(workspaceId: string, nodeIds: string[]): Promise<void> {
		const rows = this.results.get(workspaceId) ?? [];
		const set = new Set(nodeIds);
		this.results.set(
			workspaceId,
			rows.filter((r) => !set.has(r.node_id)),
		);
	}

	async saveResultsSnapshot(
		workspaceId: string,
		runId?: string,
	): Promise<string> {
		const rid = runId ?? this.runs.get(workspaceId) ?? uid();
		const ws = this.workspaces.get(workspaceId);
		if (ws) {
			ws.baselineRunId = rid;
			const snapshot = new Map<string, DbNodeResult>();
			const rows = this.results.get(workspaceId) ?? [];
			for (const nodeId of new Set(rows.map((r) => r.node_id))) {
				const latest = rows
					.filter((r) => r.node_id === nodeId)
					.sort((a, b) => b.fetched_at.localeCompare(a.fetched_at))[0];
				if (latest) snapshot.set(nodeId, { ...latest });
			}
			this.baselineResults.set(workspaceId, snapshot);
		}
		return rid;
	}

	async getWorkspaceDiff(workspaceId: string): Promise<WorkspaceDiff> {
		const ws = this.workspaces.get(workspaceId);
		const baseline = this.baselineResults.get(workspaceId);
		const nodes: WorkspaceDiff['nodes'] = [];
		let content = 0;
		let links = 0;
		let fetch = 0;

		if (!ws) {
			return {
				workspaceId,
				hasDiff: false,
				baselineRunId: null,
				nodes: [],
				summary: { content: 0, links: 0, fetch: 0 },
			};
		}

		const currentRows = this.results.get(workspaceId) ?? [];
		const latestCurrent = new Map<string, DbNodeResult>();
		for (const row of currentRows) {
			const prev = latestCurrent.get(row.node_id);
			if (!prev || row.fetched_at > prev.fetched_at) {
				latestCurrent.set(row.node_id, row);
			}
		}

		for (const node of ws.nodes) {
			const kinds: WorkspaceDiff['nodes'][0]['kinds'] = [];
			const base = baseline?.get(node.id);
			const cur = latestCurrent.get(node.id);

			if (base?.content_hash !== cur?.content_hash) {
				kinds.push('content');
				content++;
			}

			const baseLinks = base?.links_json ? JSON.parse(base.links_json) : [];
			const curLinks = cur?.links_json ? JSON.parse(cur.links_json) : [];
			const outEdges = ws.edges
				.filter((e) => e.source === node.id)
				.map((e) => e.target)
				.sort()
				.join(',');
			const baseOut = ws.edges
				.filter((e) => e.source === node.id)
				.map((e) => e.target)
				.sort()
				.join(',');
			if (
				JSON.stringify(baseLinks) !== JSON.stringify(curLinks) ||
				(!base && cur && outEdges)
			) {
				if (
					baseOut !== outEdges ||
					JSON.stringify(baseLinks) !== JSON.stringify(curLinks)
				) {
					kinds.push('links');
					links++;
				}
			}

			const baseStatus = base ? 'success' : 'idle';
			const curStatus = node.status;
			if (baseStatus !== curStatus || (base && !cur) || (!base && cur)) {
				kinds.push('fetch');
				fetch++;
			}

			if (kinds.length > 0) {
				nodes.push({ nodeId: node.id, url: node.urlNormalized, kinds });
			}
		}

		return {
			workspaceId,
			hasDiff: nodes.length > 0,
			baselineRunId: ws.baselineRunId ?? null,
			nodes,
			summary: { content, links, fetch },
		};
	}

	async startCrawl(params: StartCrawlParams): Promise<void> {
		const ws = params.getWorkspace();
		const runId = uid();
		this.runs.set(params.workspaceId, runId);
		const rows = this.results.get(params.workspaceId) ?? [];

		await runCrawlStub(
			ws,
			params.appDefaults,
			ws.seedUrl,
			{
				onNodeStarted: params.onNodeStarted,
				onNodeSucceeded: (nodeId, result) => {
					rows.push(resultToDb(runId, params.workspaceId, nodeId, result));
					this.results.set(params.workspaceId, rows);
					params.onNodeSucceeded(nodeId, result);
				},
				onNodeFailed: params.onNodeFailed,
				onNodeSkipped: params.onNodeSkipped,
				onEdgeDiscovered: params.onEdgeDiscovered,
				onCrawlCompleted: params.onCrawlCompleted,
				onCrawlError: params.onCrawlError,
			},
			{
				mode: params.mode,
				startNodeId: params.startNodeId,
				nodeIds: params.nodeIds,
				workspaceId: params.workspaceId,
				getWorkspace: params.getWorkspace,
				signal: params.signal,
				isPaused: params.isPaused,
				waitWhilePaused: params.waitWhilePaused,
			},
		);
	}

	/** Sync in-memory store from UI (bootstrap / external) */
	syncFromUi(workspaces: Workspace[], appDefaults: PartialConfig): void {
		this.appDefaults = structuredClone(appDefaults);
		this.workspaces.clear();
		for (const ws of workspaces) {
			this.workspaces.set(ws.id, structuredClone(ws));
			if (!this.results.has(ws.id)) {
				this.results.set(ws.id, []);
			}
		}
	}

	getWorkspaces(): Workspace[] {
		return [...this.workspaces.values()];
	}
}

export const mockScraperAdapter = new MockScraperAdapter();
