import { Position } from '@xyflow/react';
import { describe, expect, it } from 'vitest';
import type { GraphEdge, GraphNode } from '@/types/graph';
import {
	computeDagrePositions,
	DAGRE_NODE_HEIGHT,
	DAGRE_NODE_WIDTH,
	fallbackNearParent,
	handlePositionsForDirection,
	positionForDiscoveredNode,
} from './dagreLayout';

function node(id: string, x = 0, y = 0): GraphNode {
	return {
		id,
		urlNormalized: `https://x/${id}`,
		label: id,
		position: { x, y },
		nodeSettings: {},
		crawlExclude: false,
		status: 'idle',
	};
}

describe('dagreLayout', () => {
	it('computeDagrePositions returns distinct positions for a chain', () => {
		const nodes = [node('a'), node('b'), node('c')];
		const edges: GraphEdge[] = [
			{ id: 'e1', source: 'a', target: 'b' },
			{ id: 'e2', source: 'b', target: 'c' },
		];
		const positions = computeDagrePositions(nodes, edges, 'TB');
		expect(positions.size).toBe(3);
		const a = positions.get('a')!;
		const b = positions.get('b')!;
		const c = positions.get('c')!;
		expect(b.y).toBeGreaterThan(a.y);
		expect(c.y).toBeGreaterThan(b.y);
	});

	it('positionForDiscoveredNode only affects layout math for the new id', () => {
		const nodes = [node('a', 10, 10), node('b', 500, 500)];
		const edges: GraphEdge[] = [{ id: 'e1', source: 'a', target: 'b' }];
		const withC = [...nodes, node('c')];
		const edgesWithC: GraphEdge[] = [
			...edges,
			{ id: 'e2', source: 'b', target: 'c' },
		];
		const pos = positionForDiscoveredNode(withC, edgesWithC, 'c', {
			x: 0,
			y: 0,
		});
		expect(pos.x).toBeDefined();
		expect(pos.y).toBeDefined();
		expect(nodes[0].position).toEqual({ x: 10, y: 10 });
	});

	it('uses fallback when node id is missing from graph', () => {
		const positions = positionForDiscoveredNode([], [], 'x', {
			x: 99,
			y: 88,
		});
		expect(positions).toEqual({ x: 99, y: 88 });
	});

	it('exports node dimensions for UrlNode alignment', () => {
		expect(DAGRE_NODE_WIDTH).toBeGreaterThan(0);
		expect(DAGRE_NODE_HEIGHT).toBeGreaterThan(0);
	});

	it('handlePositionsForDirection switches handles', () => {
		expect(handlePositionsForDirection('TB').target).toBe(Position.Top);
		expect(handlePositionsForDirection('LR').source).toBe(Position.Right);
	});

	it('fallbackNearParent offsets by direction', () => {
		const parent = node('a', 100, 200);
		const v = fallbackNearParent(parent, 'TB');
		const h = fallbackNearParent(parent, 'LR');
		expect(v.y).toBeGreaterThan(parent.position.y);
		expect(h.x).toBeGreaterThan(parent.position.x);
	});
});
