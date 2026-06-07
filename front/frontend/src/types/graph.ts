import type { PartialConfig } from './config';
import type { CrawlResultPreview } from './crawl';

export type NodeStatus = 'idle' | 'running' | 'success' | 'error' | 'skipped';

export type NodeOrigin = 'crawl' | 'manual';

export interface GraphNode {
	id: string;
	urlNormalized: string;
	label: string;
	position: { x: number; y: number };
	/** ユーザーがドラッグで配置した場合 true（再生時の自動配置対象外） */
	userPositioned?: boolean;
	/** crawl: リンク発見・seed 由来 / manual: 手動追加 */
	origin?: NodeOrigin;
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
