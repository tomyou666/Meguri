import type { Workspace } from './workspace';

export type RunMode = 1 | 2 | 3 | 4;

export type CrawlRunStatus = 'idle' | 'running' | 'paused';

export interface CrawlResultPreview {
	url: string;
	markdown?: string;
	html?: string;
	raw_html?: string;
	links?: string[];
	metadata?: Record<string, string>;
}

export type LinkSkipReason = 'duplicate_existing' | 'duplicate_in_run';

export interface CrawlLogEntry {
	at: string;
	parentUrl: string;
	targetUrl: string;
	reason: LinkSkipReason;
}

export interface CrawlRunSummary {
	id: string;
	mode: RunMode;
	startedAt: string;
	finishedAt?: string;
	enqueued: number;
	succeeded: number;
	failed: number;
	skipped: number;
	skippedDuplicateLinks?: number;
	stoppedReason?: 'completed' | 'stopped' | 'error';
	errorMessage?: string;
}

export type CrawlError = {
	type: 'crawl';
	message: string;
	runId?: string;
	at: string;
} | null;

export interface CrawlEventHandlers {
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
}

export interface CrawlStubOptions {
	mode: RunMode;
	startNodeId?: string;
	/** 一括スクレイプ: 訪問するノード ID の明示リスト */
	nodeIds?: string[];
	workspaceId: string;
	getWorkspace: () => Workspace;
	signal: AbortSignal;
	isPaused: () => boolean;
	waitWhilePaused: () => Promise<void>;
	debugScenario?: 'global_fail' | 'node_fail' | 'stop_mid';
	failNodeUrl?: string;
}
