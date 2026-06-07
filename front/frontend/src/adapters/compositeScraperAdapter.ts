import { Events } from '@wailsio/runtime';
import { contentHashFromMarkdown } from '@/lib/contentHash';
import { DEFAULT_APP_CONFIG } from '@/lib/defaults';
import {
	partialConfigToRaw,
	workspaceFromDTO,
	workspaceToDTO,
} from '@/lib/wailsMappers';
import type {
	MergeResultsResponse,
	SaveSettingsResponse,
	ScraperPort,
	StartCrawlParams,
	WorkspaceDiff,
	WorkspaceListItem,
} from '@/types/adapter';
import type { PartialConfig } from '@/types/config';
import type { CrawlResultPreview, CrawlRunSummary } from '@/types/crawl';
import type { Workspace } from '@/types/workspace';
import {
	AppendNodeResultRequest,
	BeginCrawlRunRequest,
	FinishCrawlRunRequest,
	PatchGraphNodeStatusRequest,
	StartCrawlRequest,
	UpsertDiscoveredGraphRequest,
} from '../../bindings/scraperbot-front/internal/model/models';
import * as ScraperService from '../../bindings/scraperbot-front/internal/usecase/wails_service/scraperservice';
import * as StoreService from '../../bindings/scraperbot-front/internal/usecase/wails_service/storeservice';

const TOPIC_NODE_STARTED = 'scraper:crawl:nodeStarted';
const TOPIC_NODE_SUCCEEDED = 'scraper:crawl:nodeSucceeded';
const TOPIC_NODE_FAILED = 'scraper:crawl:nodeFailed';
const TOPIC_NODE_SKIPPED = 'scraper:crawl:nodeSkipped';
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

function uid(): string {
	return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
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

	async saveDomainSettings(
		workspaceId: string,
		domain: string,
		settings: PartialConfig,
	): Promise<SaveSettingsResponse> {
		const res = await StoreService.SaveDomainSettings(
			workspaceId,
			domain,
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

	async duplicateWorkspace(id: string): Promise<Workspace> {
		const dto = await StoreService.DuplicateWorkspace(id);
		if (!dto) throw new Error('Workspace not found');
		return workspaceFromDTO(dto);
	}

	async getNodeResult(
		workspaceId: string,
		nodeId: string,
	): Promise<CrawlResultPreview | null> {
		const dto = await StoreService.GetNodeResult(workspaceId, nodeId);
		if (!dto) return null;
		return {
			url: dto.url,
			markdown: dto.markdown,
			links: dto.links,
			metadata: dto.metadata,
		};
	}

	async getNodeResults(
		workspaceId: string,
		nodeIds: string[],
	): Promise<CrawlResultPreview[]> {
		const rows = await StoreService.GetNodeResults(workspaceId, nodeIds);
		return rows.map((dto) => ({
			url: dto.url,
			markdown: dto.markdown,
			links: dto.links,
			metadata: dto.metadata,
		}));
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

	async startCrawl(params: StartCrawlParams): Promise<string> {
		const ws = params.getWorkspace();
		const runId = uid();
		const startedAt = new Date().toISOString();
		params.onRunStarted?.(runId);

		await StoreService.BeginCrawlRun(
			new BeginCrawlRunRequest({
				workspaceId: params.workspaceId,
				runId,
				mode: params.mode,
				startedAt,
			}),
		);

		const finishRun = async (
			status: string,
			summary?: Omit<CrawlRunSummary, 'id' | 'startedAt'>,
			errorMessage?: string,
		) => {
			await StoreService.FinishCrawlRun(
				new FinishCrawlRunRequest({
					workspaceId: params.workspaceId,
					runId,
					status,
					finishedAt: new Date().toISOString(),
					summaryJson: summary ? JSON.stringify(summary) : undefined,
					errorMessage,
				}),
			);
		};

		const unsubscribers: Array<() => void> = [];
		const cleanup = () => {
			for (const off of unsubscribers) off();
			unsubscribers.length = 0;
		};

		const matchesRun = (payload: CrawlEventPayload) => payload.runId === runId;

		await new Promise<void>((resolve) => {
			let settled = false;
			const done = () => {
				if (settled) return;
				settled = true;
				cleanup();
				resolve();
			};

			const onAbort = () => {
				void ScraperService.StopCrawl(runId);
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
				void StoreService.PatchGraphNodeStatus(
					new PatchGraphNodeStatusRequest({
						workspaceId: params.workspaceId,
						nodeId: p.nodeId,
						status: 'running',
					}),
				);
				params.onNodeStarted(p.nodeId, p.url);
			});

			subscribe(TOPIC_NODE_SUCCEEDED, (p) => {
				if (!p.nodeId || !p.result) return;
				void (async () => {
					const result = p.result!;
					const markdown = result.markdown ?? '';
					const contentHash = markdown
						? await contentHashFromMarkdown(markdown)
						: '';
					await StoreService.AppendNodeResult(
						new AppendNodeResultRequest({
							workspaceId: params.workspaceId,
							runId,
							nodeId: p.nodeId!,
							url: result.url,
							markdown,
							linksJson: result.links
								? JSON.stringify(result.links)
								: undefined,
							metadataJson: result.metadata
								? JSON.stringify(result.metadata)
								: undefined,
							fetchedAt: new Date().toISOString(),
							contentHash,
						}),
					);
					await StoreService.PatchGraphNodeStatus(
						new PatchGraphNodeStatusRequest({
							workspaceId: params.workspaceId,
							nodeId: p.nodeId!,
							status: 'success',
						}),
					);
					params.onNodeSucceeded(p.nodeId!, result);
				})();
			});

			subscribe(TOPIC_NODE_FAILED, (p) => {
				if (!p.nodeId || !p.url) return;
				void (async () => {
					const error = p.error ?? 'unknown error';
					await StoreService.AppendNodeResult(
						new AppendNodeResultRequest({
							workspaceId: params.workspaceId,
							runId,
							nodeId: p.nodeId!,
							url: p.url!,
							error,
							fetchedAt: new Date().toISOString(),
						}),
					);
					await StoreService.PatchGraphNodeStatus(
						new PatchGraphNodeStatusRequest({
							workspaceId: params.workspaceId,
							nodeId: p.nodeId!,
							status: 'error',
							lastError: error,
						}),
					);
					params.onNodeFailed(p.nodeId!, p.url!, error);
				})();
			});

			subscribe(TOPIC_NODE_SKIPPED, (p) => {
				if (!p.nodeId || !p.url) return;
				void StoreService.PatchGraphNodeStatus(
					new PatchGraphNodeStatusRequest({
						workspaceId: params.workspaceId,
						nodeId: p.nodeId,
						status: 'skipped',
					}),
				);
				params.onNodeSkipped(p.nodeId, p.url, p.reason ?? 'skipped');
			});

			subscribe(TOPIC_EDGE_DISCOVERED, (p) => {
				if (!p.sourceId || !p.targetId || !p.targetUrl) return;
				void (async () => {
					await StoreService.UpsertDiscoveredGraph(
						new UpsertDiscoveredGraphRequest({
							workspaceId: params.workspaceId,
							sourceId: p.sourceId,
							targetId: p.targetId,
							targetUrl: p.targetUrl,
						}),
					);
					params.onEdgeDiscovered(p.sourceId, p.targetId, p.targetUrl);
				})();
			});

			subscribe(TOPIC_CRAWL_COMPLETED, (p) => {
				params.signal.removeEventListener('abort', onAbort);
				const summary = p.summary;
				void finishRun('completed', summary).then(() => {
					if (summary) params.onCrawlCompleted(summary);
					done();
				});
			});

			subscribe(TOPIC_CRAWL_ERROR, (p) => {
				params.signal.removeEventListener('abort', onAbort);
				const message = p.message ?? 'crawl error';
				void finishRun('error', undefined, message).then(() => {
					params.onCrawlError(message);
					done();
				});
			});

			const wsDto = workspaceToDTO(ws);
			void ScraperService.StartCrawl(
				new StartCrawlRequest({
					runId,
					workspaceId: params.workspaceId,
					mode: params.mode,
					startNodeId: params.startNodeId ?? '',
					nodeIds: params.nodeIds ?? [],
					appDefaults: partialConfigToRaw(params.appDefaults),
					workspace: wsDto,
				}),
			).catch((err: unknown) => {
				params.signal.removeEventListener('abort', onAbort);
				const message = err instanceof Error ? err.message : String(err);
				void finishRun('error', undefined, message).then(() => {
					params.onCrawlError(message);
					done();
				});
			});
		});

		return runId;
	}
}

export const compositeScraperAdapter = new CompositeScraperAdapter();
