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

describe('graph', () => {
	it('getForwardReachableExisting returns BFS order without start', () => {
		expect(getForwardReachableExisting('a', nodes, edges)).toEqual(['b', 'c']);
	});

	it('getDescendantNodeIds', () => {
		const d = getDescendantNodeIds('a', edges);
		expect([...d]).toEqual(['b', 'c']);
	});

	it('isExcludedSubtree detects excluded ancestor', () => {
		const withExclude: GraphNode[] = [
			{ ...nodes[0], crawlExclude: true },
			nodes[1],
			nodes[2],
		];
		expect(isExcludedSubtree('c', withExclude, edges)).toBe(true);
		expect(isExcludedSubtree('a', withExclude, edges)).toBe(true);
	});

	it('isExcludedSubtree terminates on cyclic edges', () => {
		const cycleEdges: GraphEdge[] = [
			{ id: 'e1', source: 'a', target: 'b' },
			{ id: 'e2', source: 'b', target: 'c' },
			{ id: 'e3', source: 'c', target: 'a' },
		];
		expect(isExcludedSubtree('b', nodes, cycleEdges)).toBe(false);
	});
});
