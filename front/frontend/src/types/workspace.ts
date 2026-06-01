import type { DagreLayoutDirection } from '@/lib/dagreLayout';
import type { PartialConfig } from './config';
import type { GraphEdge, GraphNode } from './graph';

export interface Workspace {
	id: string;
	name: string;
	seedUrl: string;
	settings: PartialConfig;
	exclude_urls: string[];
	nodes: GraphNode[];
	edges: GraphEdge[];
	/** dagre 自動配置の向き（TB=縦, LR=横） */
	graphLayoutDirection: DagreLayoutDirection;
	domainSettings: Record<string, PartialConfig>;
}
