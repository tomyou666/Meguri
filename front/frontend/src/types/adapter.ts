import type { PartialConfig } from './config';
import type {
	CrawlResultPreview,
	CrawlRunSummary,
	LinkSkipReason,
	RunMode,
} from './crawl';
import type { Workspace } from './workspace';

export type DiffKind = 'content' | 'links' | 'fetch';

export interface NodeDiff {
	nodeId: string;
	url: string;
	kinds: DiffKind[];
}

export interface WorkspaceDiff {
	workspaceId: string;
	hasDiff: boolean;
	baselineRunId: string | null;
	nodes: NodeDiff[];
	summary: {
		content: number;
		links: number;
		fetch: number;
	};
}

export interface MergeResultsResponse {
	merged: string;
	format: string;
	nodeCount: number;
}

export interface StartCrawlParams {
	workspaceId: string;
	mode: RunMode;
	startNodeId?: string;
	nodeIds?: string[];
	rescrapeExisting?: boolean;
	appDefaults: PartialConfig;
	signal: AbortSignal;
	isPaused: () => boolean;
	waitWhilePaused: () => Promise<void>;
	onNodeStarted: (nodeId: string, url: string) => void;
	onNodeSucceeded: (nodeId: string, result: CrawlResultPreview) => void;
	onNodeFailed: (nodeId: string, url: string, error: string) => void;
	onNodeSkipped: (nodeId: string, url: string, reason: string) => void;
	onLinkSkipped: (
		parentUrl: string,
		targetUrl: string,
		reason: LinkSkipReason,
	) => void;
	onEdgeDiscovered: (
		sourceId: string,
		targetId: string,
		targetUrl: string,
	) => void;
	onCrawlCompleted: (
		summary: Omit<CrawlRunSummary, 'id' | 'startedAt'>,
	) => void;
	onCrawlError: (message: string) => void;
	onRunStarted?: (runId: string) => void;
	getWorkspace: () => Workspace;
}

export type SettingsScope = 'app' | 'workspace' | 'domain' | 'node';

export interface WorkspaceListItem {
	id: string;
	name: string;
	updatedAt: string;
}

export interface SaveSettingsResponse {
	ok: boolean;
	scope: SettingsScope;
}

export interface ScraperPort {
	getAppDefaults(): Promise<PartialConfig>;
	setAppDefaults(config: PartialConfig): Promise<void>;
	/** 既定設定の保存（バリデーション済み JSON） */
	saveAppDefaults(config: PartialConfig): Promise<SaveSettingsResponse>;

	listWorkspaces(): Promise<WorkspaceListItem[]>;
	loadWorkspace(id: string): Promise<Workspace | null>;
	saveWorkspace(ws: Workspace): Promise<void>;
	saveWorkspaceSettings(
		workspaceId: string,
		settings: PartialConfig,
	): Promise<SaveSettingsResponse>;
	saveDomainSettings(
		workspaceId: string,
		domain: string,
		settings: PartialConfig,
	): Promise<SaveSettingsResponse>;
	saveNodeSettings(
		workspaceId: string,
		nodeId: string,
		settings: PartialConfig,
	): Promise<SaveSettingsResponse>;
	duplicateWorkspace(id: string): Promise<Workspace>;

	getNodeResult(
		workspaceId: string,
		nodeId: string,
	): Promise<CrawlResultPreview | null>;
	getNodeResults(
		workspaceId: string,
		nodeIds: string[],
	): Promise<CrawlResultPreview[]>;
	mergeResults(
		workspaceId: string,
		nodeIds: string[] | null,
		formats?: string[],
	): Promise<MergeResultsResponse>;
	saveResults(workspaceId: string, nodeIds: string[]): Promise<void>;
	deleteResults(workspaceId: string, nodeIds: string[]): Promise<void>;
	saveResultsSnapshot(workspaceId: string, runId?: string): Promise<string>;

	getWorkspaceDiff(workspaceId: string): Promise<WorkspaceDiff>;

	startCrawl(params: StartCrawlParams): Promise<string>;
}
