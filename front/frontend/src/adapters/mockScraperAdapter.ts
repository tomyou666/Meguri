import {
	canonicalizeLinksJson,
	contentHashFromMarkdown,
} from '@/lib/contentHash';
import { workspaceFromDb, workspaceToDb } from '@/lib/dbMappers';
import { DEFAULT_APP_CONFIG } from '@/lib/defaults';
import {
	appendNodeResult,
	deleteLatestResults,
	latestSuccessByNode,
	MAX_CRAWL_RUN_HISTORY,
	rowsForRun,
} from '@/lib/nodeResultStore';
import { runCrawlStub } from '@/services/crawlStub';
import type {
	MergeResultsResponse,
	SaveSettingsResponse,
	ScraperPort,
	StartCrawlParams,
	WorkspaceDiff,
} from '@/types/adapter';
import type { PartialConfig } from '@/types/config';
import type { CrawlResultPreview, CrawlRunSummary } from '@/types/crawl';
import type { DbCrawlRun, DbNodeResult } from '@/types/db';
import type { Workspace } from '@/types/workspace';

function uid(): string {
	return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

async function resultToDb(
	runId: string,
	workspaceId: string,
	nodeId: string,
	result: CrawlResultPreview,
): Promise<DbNodeResult> {
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
		content_hash: markdown ? await contentHashFromMarkdown(markdown) : null,
	};
}

function failureToDb(
	runId: string,
	workspaceId: string,
	nodeId: string,
	url: string,
	error: string,
): DbNodeResult {
	return {
		id: uid(),
		run_id: runId,
		workspace_id: workspaceId,
		node_id: nodeId,
		url,
		markdown: null,
		html: null,
		raw_html: null,
		json_body: null,
		links_json: null,
		metadata_json: null,
		error,
		fetched_at: new Date().toISOString(),
		content_hash: null,
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

function trimCrawlRuns(runs: DbCrawlRun[]): DbCrawlRun[] {
	return [...runs]
		.sort((a, b) => b.started_at.localeCompare(a.started_at))
		.slice(0, MAX_CRAWL_RUN_HISTORY);
}

function upsertBaselineRow(
	rows: DbNodeResult[],
	baselineRunId: string,
	source: DbNodeResult,
): DbNodeResult[] {
	const without = rows.filter(
		(r) => !(r.run_id === baselineRunId && r.node_id === source.node_id),
	);
	return [
		...without,
		{
			...source,
			id: uid(),
			run_id: baselineRunId,
			fetched_at: new Date().toISOString(),
		},
	];
}

export class MockScraperAdapter implements ScraperPort {
	private appDefaults: PartialConfig = { ...DEFAULT_APP_CONFIG };
	private workspaces = new Map<string, Workspace>();
	private results = new Map<string, DbNodeResult[]>();
	private crawlRuns = new Map<string, DbCrawlRun[]>();

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
		const rows = this.results.get(ws.id) ?? [];
		const hydrated = latestSuccessByNode(rows);
		const previewMap = new Map<string, CrawlResultPreview>();
		for (const [nodeId, row] of hydrated) {
			previewMap.set(nodeId, dbToPreview(row));
		}
		const bundle = workspaceToDb(ws);
		return workspaceFromDb(bundle, previewMap);
	}

	async saveWorkspace(ws: Workspace): Promise<void> {
		this.workspaces.set(ws.id, structuredClone(ws));
	}

	private ensureBaselineRun(workspaceId: string, ws: Workspace): string {
		if (ws.baselineRunId) return ws.baselineRunId;
		const runId = uid();
		const run: DbCrawlRun = {
			id: runId,
			workspace_id: workspaceId,
			mode: 1,
			status: 'completed',
			started_at: new Date().toISOString(),
			finished_at: new Date().toISOString(),
			summary_json: null,
			error_message: null,
		};
		const runs = this.crawlRuns.get(workspaceId) ?? [];
		this.crawlRuns.set(workspaceId, trimCrawlRuns([run, ...runs]));
		ws.baselineRunId = runId;
		return runId;
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
		this.crawlRuns.set(newId, []);
		return copy;
	}

	async getNodeResult(
		workspaceId: string,
		nodeId: string,
	): Promise<CrawlResultPreview | null> {
		const rows = this.results.get(workspaceId) ?? [];
		const row = latestSuccessByNode(rows).get(nodeId);
		return row ? dbToPreview(row) : null;
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
		const ws = this.workspaces.get(workspaceId);
		if (!ws) return;
		const baselineRunId = this.ensureBaselineRun(workspaceId, ws);
		let rows = this.results.get(workspaceId) ?? [];
		const latest = latestSuccessByNode(rows);
		for (const nodeId of nodeIds) {
			const source = latest.get(nodeId);
			if (source) {
				rows = upsertBaselineRow(rows, baselineRunId, source);
			}
		}
		this.results.set(workspaceId, rows);
	}

	async deleteResults(workspaceId: string, nodeIds: string[]): Promise<void> {
		const rows = this.results.get(workspaceId) ?? [];
		this.results.set(workspaceId, deleteLatestResults(rows, nodeIds));
	}

	async saveResultsSnapshot(
		workspaceId: string,
		runId?: string,
	): Promise<string> {
		const ws = this.workspaces.get(workspaceId);
		if (!ws) return runId ?? uid();

		const rid = runId ?? uid();
		const runs = this.crawlRuns.get(workspaceId) ?? [];
		if (!runs.some((r) => r.id === rid)) {
			const run: DbCrawlRun = {
				id: rid,
				workspace_id: workspaceId,
				mode: 1,
				status: 'completed',
				started_at: new Date().toISOString(),
				finished_at: new Date().toISOString(),
				summary_json: null,
				error_message: null,
			};
			this.crawlRuns.set(workspaceId, trimCrawlRuns([run, ...runs]));
		}

		let rows = this.results.get(workspaceId) ?? [];
		const latest = latestSuccessByNode(rows);
		for (const source of latest.values()) {
			rows = upsertBaselineRow(rows, rid, source);
		}
		this.results.set(workspaceId, rows);
		ws.baselineRunId = rid;
		return rid;
	}

	async getWorkspaceDiff(workspaceId: string): Promise<WorkspaceDiff> {
		const ws = this.workspaces.get(workspaceId);
		const nodes: WorkspaceDiff['nodes'] = [];
		let content = 0;
		let links = 0;
		let fetch = 0;

		if (!ws?.baselineRunId) {
			return {
				workspaceId,
				hasDiff: false,
				baselineRunId: null,
				nodes: [],
				summary: { content: 0, links: 0, fetch: 0 },
			};
		}

		const rows = this.results.get(workspaceId) ?? [];
		const baseline = rowsForRun(rows, ws.baselineRunId);
		const current = latestSuccessByNode(rows);

		for (const node of ws.nodes) {
			const kinds: WorkspaceDiff['nodes'][0]['kinds'] = [];
			const base = baseline.get(node.id);
			const cur = current.get(node.id);

			if (base?.content_hash !== cur?.content_hash) {
				kinds.push('content');
				content++;
			}

			const baseLinks = base?.links_json
				? (JSON.parse(base.links_json) as string[])
				: [];
			const curLinks = cur?.links_json
				? (JSON.parse(cur.links_json) as string[])
				: [];
			if (
				canonicalizeLinksJson(baseLinks) !== canonicalizeLinksJson(curLinks)
			) {
				kinds.push('links');
				links++;
			}

			const baseOk = base != null && !base.error;
			const curOk = cur != null && !cur.error;
			const baseFetch = baseOk ? 'success' : base ? 'error' : 'none';
			const curFetch = curOk
				? 'success'
				: cur
					? 'error'
					: node.status === 'skipped'
						? 'skipped'
						: 'none';
			if (baseFetch !== curFetch) {
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
			baselineRunId: ws.baselineRunId,
			nodes,
			summary: { content, links, fetch },
		};
	}

	async startCrawl(params: StartCrawlParams): Promise<void> {
		const ws = params.getWorkspace();
		const runId = uid();
		const startedAt = new Date().toISOString();
		const run: DbCrawlRun = {
			id: runId,
			workspace_id: params.workspaceId,
			mode: params.mode,
			status: 'running',
			started_at: startedAt,
			finished_at: null,
			summary_json: null,
			error_message: null,
		};
		const runs = this.crawlRuns.get(params.workspaceId) ?? [];
		this.crawlRuns.set(params.workspaceId, trimCrawlRuns([run, ...runs]));

		const finishRun = (
			status: DbCrawlRun['status'],
			summary?: Omit<CrawlRunSummary, 'id' | 'startedAt'>,
			errorMessage?: string,
		) => {
			const list = this.crawlRuns.get(params.workspaceId) ?? [];
			const idx = list.findIndex((r) => r.id === runId);
			if (idx >= 0) {
				list[idx] = {
					...list[idx],
					status,
					finished_at: new Date().toISOString(),
					summary_json: summary ? JSON.stringify(summary) : null,
					error_message: errorMessage ?? null,
				};
				this.crawlRuns.set(params.workspaceId, list);
			}
		};

		await runCrawlStub(
			ws,
			params.appDefaults,
			ws.seedUrl,
			{
				onNodeStarted: params.onNodeStarted,
				onNodeSucceeded: (nodeId, result) => {
					void resultToDb(runId, params.workspaceId, nodeId, result).then(
						(row) => {
							const current = this.results.get(params.workspaceId) ?? [];
							this.results.set(
								params.workspaceId,
								appendNodeResult(current, row),
							);
							params.onNodeSucceeded(nodeId, result);
						},
					);
				},
				onNodeFailed: (nodeId, url, error) => {
					const row = failureToDb(
						runId,
						params.workspaceId,
						nodeId,
						url,
						error,
					);
					const current = this.results.get(params.workspaceId) ?? [];
					this.results.set(params.workspaceId, appendNodeResult(current, row));
					params.onNodeFailed(nodeId, url, error);
				},
				onNodeSkipped: params.onNodeSkipped,
				onEdgeDiscovered: params.onEdgeDiscovered,
				onCrawlCompleted: (summary) => {
					finishRun('completed', summary);
					params.onCrawlCompleted(summary);
				},
				onCrawlError: (message) => {
					finishRun('error', undefined, message);
					params.onCrawlError(message);
				},
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
			if (!this.crawlRuns.has(ws.id)) {
				this.crawlRuns.set(ws.id, []);
			}
		}
	}

	getWorkspaces(): Workspace[] {
		return [...this.workspaces.values()];
	}

	/** テスト・デバッグ用 */
	getCrawlRuns(workspaceId: string): DbCrawlRun[] {
		return [...(this.crawlRuns.get(workspaceId) ?? [])];
	}

	getNodeResultsRows(workspaceId: string): DbNodeResult[] {
		return [...(this.results.get(workspaceId) ?? [])];
	}
}

export const mockScraperAdapter = new MockScraperAdapter();
