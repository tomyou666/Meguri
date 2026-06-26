import { Events } from '@wailsio/runtime';
import { DEFAULT_APP_CONFIG } from '@/lib/defaults';
import {
	crawlResultFromDTO,
	crawlResultToDTO,
	partialConfigToRaw,
	workspaceFromDTO,
	workspaceToDTO,
} from '@/lib/wailsMappers';
import type {
	ExportSessionSnapshot,
	MaximizedNodeResultSnapshot,
	MergeResultsResponse,
	NodeDiffDetail,
	NodeDiffViewerSnapshot,
	SaveSettingsResponse,
	ScraperPort,
	StartCrawlParams,
	UpdateNodeResultPatch,
	WorkspaceDiff,
	WorkspaceListItem,
} from '@/types/adapter';
import type { PartialConfig } from '@/types/config';
import type {
	CrawlResultPreview,
	CrawlRunSummary,
	LinkSkipReason,
} from '@/types/crawl';
import type { Workspace } from '@/types/workspace';
import {
	ExportSessionRequest,
	MaximizedNodeResultRequest,
	NodeDiffViewerRequest,
	StartCrawlRequest,
	UpdateNodeResultPatchDTO,
	UpdateNodeResultRequest,
} from '../../bindings/scraperbot-front/internal/model/models.js';
import * as ScraperService from '../../bindings/scraperbot-front/internal/usecase/wails_service/scraperservice';
import * as StoreService from '../../bindings/scraperbot-front/internal/usecase/wails_service/storeservice';

const TOPIC_NODE_STARTED = 'scraper:crawl:nodeStarted';
const TOPIC_NODE_SUCCEEDED = 'scraper:crawl:nodeSucceeded';
const TOPIC_NODE_FAILED = 'scraper:crawl:nodeFailed';
const TOPIC_NODE_SKIPPED = 'scraper:crawl:nodeSkipped';
const TOPIC_LINK_SKIPPED = 'scraper:crawl:linkSkipped';
const TOPIC_EDGE_DISCOVERED = 'scraper:crawl:edgeDiscovered';
const TOPIC_CRAWL_COMPLETED = 'scraper:crawl:completed';
const TOPIC_CRAWL_ERROR = 'scraper:crawl:error';

interface CrawlEventPayload {
	workspaceId: string;
	runId: string;
	nodeId?: string;
	url?: string;
	result?: CrawlResultPreview;
	error?: string;
	reason?: string;
	sourceId?: string;
	targetId?: string;
	targetUrl?: string;
	summary?: Omit<CrawlRunSummary, 'id' | 'startedAt'>;
	message?: string;
}

function parseDefaults(raw: unknown): PartialConfig {
	if (raw == null) return { ...DEFAULT_APP_CONFIG };
	const text = typeof raw === 'string' ? raw : JSON.stringify(raw);
	try {
		return JSON.parse(text) as PartialConfig;
	} catch {
		return { ...DEFAULT_APP_CONFIG };
	}
}

function eventData(ev: { data?: unknown }): CrawlEventPayload {
	return (ev.data ?? {}) as CrawlEventPayload;
}

export class CompositeScraperAdapter implements ScraperPort {
	async getAppDefaults(): Promise<PartialConfig> {
		const raw = await StoreService.GetAppDefaults();
		return parseDefaults(raw);
	}

	async setAppDefaults(config: PartialConfig): Promise<void> {
		await StoreService.SetAppDefaults(partialConfigToRaw(config));
	}

	async saveAppDefaults(config: PartialConfig): Promise<SaveSettingsResponse> {
		const res = await StoreService.SaveAppDefaults(partialConfigToRaw(config));
		return { ok: res.ok, scope: res.scope as SaveSettingsResponse['scope'] };
	}

	async listWorkspaces(): Promise<WorkspaceListItem[]> {
		const items = await StoreService.ListWorkspaces();
		return items.map((it) => ({
			id: it.id,
			name: it.name,
			updatedAt: it.updatedAt,
		}));
	}

	async loadWorkspace(id: string): Promise<Workspace | null> {
		const dto = await StoreService.LoadWorkspace(id);
		if (!dto) return null;
		return workspaceFromDTO(dto);
	}

	async saveWorkspace(ws: Workspace): Promise<void> {
		await StoreService.SaveWorkspace(workspaceToDTO(ws));
	}

	async saveWorkspaceSettings(
		workspaceId: string,
		settings: PartialConfig,
	): Promise<SaveSettingsResponse> {
		const res = await StoreService.SaveWorkspaceSettings(
			workspaceId,
			partialConfigToRaw(settings),
		);
		return { ok: res.ok, scope: res.scope as SaveSettingsResponse['scope'] };
	}

	async saveNodeSettings(
		workspaceId: string,
		nodeId: string,
		settings: PartialConfig,
	): Promise<SaveSettingsResponse> {
		const res = await StoreService.SaveNodeSettings(
			workspaceId,
			nodeId,
			partialConfigToRaw(settings),
		);
		return { ok: res.ok, scope: res.scope as SaveSettingsResponse['scope'] };
	}

	async deleteWorkspace(id: string): Promise<void> {
		await StoreService.DeleteWorkspace(id);
	}

	async duplicateWorkspace(id: string, name: string): Promise<Workspace> {
		const dto = await StoreService.DuplicateWorkspace(id, name);
		if (!dto) throw new Error('Workspace not found');
		return workspaceFromDTO(dto);
	}

	async getNodeResult(
		workspaceId: string,
		nodeId: string,
	): Promise<CrawlResultPreview | null> {
		const dto = await StoreService.GetNodeResult(workspaceId, nodeId);
		if (!dto) return null;
		return crawlResultFromDTO(dto);
	}

	async getNodeResults(
		workspaceId: string,
		nodeIds: string[],
	): Promise<CrawlResultPreview[]> {
		const rows = await StoreService.GetNodeResults(workspaceId, nodeIds);
		return rows.map((dto) => crawlResultFromDTO(dto));
	}

	async updateNodeResult(
		workspaceId: string,
		nodeId: string,
		patch: UpdateNodeResultPatch,
	): Promise<CrawlResultPreview | null> {
		const patchDto = new UpdateNodeResultPatchDTO();
		if (patch.markdown !== undefined) patchDto.markdown = patch.markdown;
		if (patch.html !== undefined) patchDto.html = patch.html;
		if (patch.raw_html !== undefined) patchDto.rawHtml = patch.raw_html;
		if (patch.json !== undefined) patchDto.jsonBody = patch.json;
		const dto = await StoreService.UpdateNodeResult(
			new UpdateNodeResultRequest({
				workspaceId,
				nodeId,
				patch: patchDto,
			}),
		);
		return dto ? crawlResultFromDTO(dto) : null;
	}

	async showMaximizedNodeResult(
		snapshot: MaximizedNodeResultSnapshot,
	): Promise<void> {
		await StoreService.ShowMaximizedNodeResult(
			new MaximizedNodeResultRequest({
				title: snapshot.title,
				workspaceId: snapshot.workspaceId,
				nodeId: snapshot.nodeId,
				activeFormat: snapshot.activeFormat,
				markdownView: snapshot.markdownView,
				formats: snapshot.formats,
				result: crawlResultToDTO(snapshot.result),
			}),
		);
	}

	async getMaximizedNodeResult(): Promise<MaximizedNodeResultSnapshot | null> {
		try {
			const dto = await StoreService.GetMaximizedNodeResult();
			if (!dto?.result) return null;
			return {
				title: dto.title,
				workspaceId: dto.workspaceId ?? '',
				nodeId: dto.nodeId ?? '',
				activeFormat: dto.activeFormat,
				markdownView: dto.markdownView === 'source' ? 'source' : 'preview',
				formats: dto.formats ?? [],
				result: crawlResultFromDTO(dto.result),
			};
		} catch {
			return null;
		}
	}

	async showExportWindow(snapshot: ExportSessionSnapshot): Promise<void> {
		await StoreService.ShowExportWindow(
			new ExportSessionRequest({
				title: snapshot.title,
				workspaceId: snapshot.workspaceId,
				mode: snapshot.mode,
				seedUrl: snapshot.seedUrl,
				nodes: snapshot.nodes.map((n) => ({
					id: n.id,
					urlNormalized: n.urlNormalized,
					label: n.label,
					status: n.status,
				})),
				edges: snapshot.edges.map((e) => ({
					source: e.source,
					target: e.target,
				})),
				selectedNodeIds: snapshot.selectedNodeIds ?? [],
			}),
		);
	}

	async getExportSession(): Promise<ExportSessionSnapshot | null> {
		try {
			const dto = await StoreService.GetExportSession();
			if (!dto?.workspaceId) return null;
			return {
				title: dto.title,
				workspaceId: dto.workspaceId,
				mode: dto.mode === 'selected' ? 'selected' : 'all',
				seedUrl: dto.seedUrl,
				nodes: (dto.nodes ?? []).map((n) => ({
					id: n.id,
					urlNormalized: n.urlNormalized,
					label: n.label,
					status: n.status,
				})),
				edges: (dto.edges ?? []).map((e) => ({
					source: e.source,
					target: e.target,
				})),
				selectedNodeIds: dto.selectedNodeIds ?? [],
			};
		} catch {
			return null;
		}
	}

	async saveExportFile(content: string, defaultExt: string): Promise<void> {
		await StoreService.SaveExportFile(content, defaultExt);
	}

	async saveExportZip(
		entries: { name: string; content: string }[],
		defaultExt: string,
	): Promise<void> {
		await StoreService.SaveExportZip(entries, defaultExt);
	}

	async mergeResults(
		workspaceId: string,
		nodeIds: string[] | null,
		formats?: string[],
	): Promise<MergeResultsResponse> {
		const res = await StoreService.MergeResults(
			workspaceId,
			nodeIds ?? [],
			formats ?? [],
		);
		return {
			merged: res.merged,
			format: res.format,
			nodeCount: res.nodeCount,
		};
	}

	async saveResults(workspaceId: string, nodeIds: string[]): Promise<void> {
		await StoreService.SaveResults(workspaceId, nodeIds);
	}

	async deleteResults(workspaceId: string, nodeIds: string[]): Promise<void> {
		await StoreService.DeleteResults(workspaceId, nodeIds);
	}

	async saveResultsSnapshot(
		workspaceId: string,
		runId?: string,
	): Promise<string> {
		return StoreService.SaveResultsSnapshot(workspaceId, runId ?? '');
	}

	async getWorkspaceDiff(workspaceId: string): Promise<WorkspaceDiff> {
		const dto = await StoreService.GetWorkspaceDiff(workspaceId);
		if (!dto) {
			return {
				workspaceId,
				hasDiff: false,
				baselineRunId: null,
				nodes: [],
				summary: { content: 0, links: 0, fetch: 0 },
			};
		}
		return {
			workspaceId: dto.workspaceId,
			hasDiff: dto.hasDiff,
			baselineRunId: dto.baselineRunId || null,
			nodes: (dto.nodes ?? []).map((n) => ({
				nodeId: n.nodeId,
				url: n.url,
				kinds: n.kinds as WorkspaceDiff['nodes'][0]['kinds'],
			})),
			summary: dto.summary,
		};
	}

	async getNodeDiffDetail(
		workspaceId: string,
		nodeId: string,
	): Promise<NodeDiffDetail> {
		const dto = await StoreService.GetNodeDiffDetail(workspaceId, nodeId);
		return {
			nodeId: dto.nodeId,
			url: dto.url,
			kinds: (dto.kinds ?? []) as NodeDiffDetail['kinds'],
			content: dto.content
				? { old: dto.content.old, new: dto.content.new }
				: undefined,
			links: dto.links ? { old: dto.links.old, new: dto.links.new } : undefined,
			fetch: dto.fetch ? { old: dto.fetch.old, new: dto.fetch.new } : undefined,
		};
	}

	async showNodeDiffWindow(snapshot: NodeDiffViewerSnapshot): Promise<void> {
		await StoreService.ShowNodeDiffWindow(
			new NodeDiffViewerRequest({
				workspaceId: snapshot.workspaceId,
				nodeId: snapshot.nodeId,
				initialKind: snapshot.initialKind ?? '',
				title: snapshot.title,
			}),
		);
	}

	async getNodeDiffViewerSession(): Promise<NodeDiffViewerSnapshot | null> {
		try {
			const dto = await StoreService.GetNodeDiffViewerSession();
			if (!dto?.workspaceId || !dto.nodeId) return null;
			return {
				workspaceId: dto.workspaceId,
				nodeId: dto.nodeId,
				initialKind: (dto.initialKind || undefined) as
					| NodeDiffViewerSnapshot['initialKind']
					| undefined,
				title: dto.title,
			};
		} catch {
			return null;
		}
	}

	async startCrawl(params: StartCrawlParams): Promise<string> {
		const ws = params.getWorkspace();
		const wsDto = workspaceToDTO(ws);

		let runId = '';
		const unsubscribers: Array<() => void> = [];
		const cleanup = () => {
			for (const off of unsubscribers) off();
			unsubscribers.length = 0;
		};

		const matchesRun = (payload: CrawlEventPayload) =>
			runId !== '' && payload.runId === runId;

		await new Promise<void>((resolve) => {
			let settled = false;
			const done = () => {
				if (settled) return;
				settled = true;
				cleanup();
				resolve();
			};

			const onAbort = () => {
				if (runId) void ScraperService.StopCrawl(runId);
			};
			params.signal.addEventListener('abort', onAbort);

			const subscribe = (
				topic: string,
				handler: (p: CrawlEventPayload) => void,
			) => {
				const off = Events.On(topic, (ev) => {
					const p = eventData(ev);
					if (!matchesRun(p)) return;
					handler(p);
				});
				unsubscribers.push(off);
			};

			subscribe(TOPIC_NODE_STARTED, (p) => {
				if (!p.nodeId || !p.url) return;
				params.onNodeStarted(p.nodeId, p.url);
			});

			subscribe(TOPIC_NODE_SUCCEEDED, (p) => {
				if (!p.nodeId || !p.result) return;
				params.onNodeSucceeded(p.nodeId, p.result);
			});

			subscribe(TOPIC_NODE_FAILED, (p) => {
				if (!p.nodeId || !p.url) return;
				params.onNodeFailed(p.nodeId, p.url, p.error ?? 'unknown error');
			});

			subscribe(TOPIC_NODE_SKIPPED, (p) => {
				if (!p.nodeId || !p.url) return;
				params.onNodeSkipped(p.nodeId, p.url, p.reason ?? 'skipped');
			});

			subscribe(TOPIC_LINK_SKIPPED, (p) => {
				if (!p.targetUrl) return;
				const reason = (p.reason ?? 'duplicate_in_run') as LinkSkipReason;
				params.onLinkSkipped(p.url ?? '', p.targetUrl, reason);
			});

			subscribe(TOPIC_EDGE_DISCOVERED, (p) => {
				if (!p.sourceId || !p.targetId || !p.targetUrl) return;
				params.onEdgeDiscovered(p.sourceId, p.targetId, p.targetUrl);
			});

			subscribe(TOPIC_CRAWL_COMPLETED, (p) => {
				params.signal.removeEventListener('abort', onAbort);
				if (p.summary) params.onCrawlCompleted(p.summary);
				done();
			});

			subscribe(TOPIC_CRAWL_ERROR, (p) => {
				params.signal.removeEventListener('abort', onAbort);
				params.onCrawlError(p.message ?? 'crawl error');
				done();
			});

			void ScraperService.StartCrawl(
				new StartCrawlRequest({
					runId: '',
					workspaceId: params.workspaceId,
					mode: params.mode,
					startNodeId: params.startNodeId ?? '',
					nodeIds: params.nodeIds ?? [],
					rescrapeExisting: params.rescrapeExisting ?? false,
					appDefaults: partialConfigToRaw(params.appDefaults),
					workspace: wsDto,
				}),
			)
				.then((id) => {
					runId = id;
					params.onRunStarted?.(runId);
				})
				.catch((err: unknown) => {
					params.signal.removeEventListener('abort', onAbort);
					const message = err instanceof Error ? err.message : String(err);
					params.onCrawlError(message);
					done();
				});
		});

		return runId;
	}
}

export const compositeScraperAdapter = new CompositeScraperAdapter();
