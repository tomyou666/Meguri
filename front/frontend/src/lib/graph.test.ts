import { describe, expect, it } from 'vitest';
import type { GraphEdge, GraphNode } from '@/types/graph';
import {
	getDescendantNodeIds,
	getForwardReachableExisting,
	isExcludedSubtree,
} from './graph';

const nodes: GraphNode[] = [
	{
		id: 'a',
		urlNormalized: 'https://x/a',
		label: 'a',
		position: { x: 0, y: 0 },
		nodeSettings: {},
		crawlExclude: false,
		status: 'idle',
	},
	{
		id: 'b',
		urlNormalized: 'https://x/b',
		label: 'b',
		position: { x: 0, y: 0 },
		nodeSettings: {},
		crawlExclude: false,
		status: 'idle',
	},
	{
		id: 'c',
		urlNormalized: 'https://x/c',
		label: 'c',
		position: { x: 0, y: 0 },
		nodeSettings: {},
		crawlExclude: false,
		status: 'idle',
	},
];
const edges: GraphEdge[] = [
	{ id: 'e1', source: 'a', target: 'b' },
	{ id: 'e2', source: 'b', target: 'c' },
];

// グラフ走査と除外サブツリー判定を検証する。
describe('graph', () => {
	it('起点から BFS 順で到達可能な既存ノード ID を返す（起点除く）', () => {
		expect(getForwardReachableExisting('a', nodes, edges)).toEqual(['b', 'c']);
	});

	it('起点からエッジを辿った全子孫ノード ID を返す', () => {
		const d = getDescendantNodeIds('a', edges);
		expect([...d]).toEqual(['b', 'c']);
	});

	it('除外祖先があるサブツリー内のノードを検出する', () => {
		const withExclude: GraphNode[] = [
			{ ...nodes[0], crawlExclude: true },
			nodes[1],
			nodes[2],
		];
		expect(isExcludedSubtree('c', withExclude, edges)).toBe(true);
		expect(isExcludedSubtree('a', withExclude, edges)).toBe(true);
	});

	it('循環エッジでも除外判定が終了する', () => {
		const cycleEdges: GraphEdge[] = [
			{ id: 'e1', source: 'a', target: 'b' },
			{ id: 'e2', source: 'b', target: 'c' },
			{ id: 'e3', source: 'c', target: 'a' },
		];
		expect(isExcludedSubtree('b', nodes, cycleEdges)).toBe(false);
	});
});
