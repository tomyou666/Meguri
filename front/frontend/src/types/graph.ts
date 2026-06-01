import type { PartialConfig } from './config';
import type { CrawlResultPreview } from './crawl';

export type NodeStatus = 'idle' | 'running' | 'success' | 'error' | 'skipped';

export interface GraphNode {
	id: string;
	urlNormalized: string;
	label: string;
	position: { x: number; y: number };
	/** ユーザーがドラッグで配置した場合 true（再生時の自動配置対象外） */
	userPositioned?: boolean;
	nodeSettings: PartialConfig;
	crawlExclude: boolean;
	status: NodeStatus;
	lastResult?: CrawlResultPreview;
	lastError?: string;
}

export interface GraphEdge {
	id: string;
	source: string;
	target: string;
}
