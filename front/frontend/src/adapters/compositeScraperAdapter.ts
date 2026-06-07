import { contentHashFromMarkdown } from '@/lib/contentHash';
import { DEFAULT_APP_CONFIG } from '@/lib/defaults';
import {
	partialConfigToRaw,
	workspaceFromDTO,
	workspaceToDTO,
} from '@/lib/wailsMappers';
import { runCrawlStub } from '@/services/crawlStub';
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
} from '../../bindings/scraperbot-front/internal/model/models';
import * as StoreService from '../../bindings/scraperbot-front/internal/usecase/wails_service/storeservice';

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
			formats ?? ['markdown'],
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

	async startCrawl(params: StartCrawlParams): Promise<void> {
		const ws = params.getWorkspace();
		const runId = uid();
		const startedAt = new Date().toISOString();

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

		await runCrawlStub(
			ws,
			params.appDefaults,
			ws.seedUrl,
			{
				onNodeStarted: (nodeId, url) => {
					void StoreService.PatchGraphNodeStatus(
						new PatchGraphNodeStatusRequest({
							workspaceId: params.workspaceId,
							nodeId,
							status: 'running',
						}),
					);
					params.onNodeStarted(nodeId, url);
				},
				onNodeSucceeded: (nodeId, result) => {
					void (async () => {
						const markdown = result.markdown ?? '';
						const contentHash = markdown
							? await contentHashFromMarkdown(markdown)
							: '';
						await StoreService.AppendNodeResult(
							new AppendNodeResultRequest({
								workspaceId: params.workspaceId,
								runId,
								nodeId,
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
								nodeId,
								status: 'success',
							}),
						);
						params.onNodeSucceeded(nodeId, result);
					})();
				},
				onNodeFailed: (nodeId, url, error) => {
					void (async () => {
						await StoreService.AppendNodeResult(
							new AppendNodeResultRequest({
								workspaceId: params.workspaceId,
								runId,
								nodeId,
								url,
								error,
								fetchedAt: new Date().toISOString(),
							}),
						);
						await StoreService.PatchGraphNodeStatus(
							new PatchGraphNodeStatusRequest({
								workspaceId: params.workspaceId,
								nodeId,
								status: 'error',
								lastError: error,
							}),
						);
						params.onNodeFailed(nodeId, url, error);
					})();
				},
				onNodeSkipped: (nodeId, url, reason) => {
					void StoreService.PatchGraphNodeStatus(
						new PatchGraphNodeStatusRequest({
							workspaceId: params.workspaceId,
							nodeId,
							status: 'skipped',
						}),
					);
					params.onNodeSkipped(nodeId, url, reason);
				},
				onEdgeDiscovered: params.onEdgeDiscovered,
				onCrawlCompleted: (summary) => {
					void finishRun('completed', summary).then(() =>
						params.onCrawlCompleted(summary),
					);
				},
				onCrawlError: (message) => {
					void finishRun('error', undefined, message).then(() =>
						params.onCrawlError(message),
					);
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
}

export const compositeScraperAdapter = new CompositeScraperAdapter();
