import { describe, expect, it } from 'vitest';
import {
	buildNodeFlatTree,
	expandableNodeIds,
	expandIdsForHits,
	filterNodeTree,
	groupChildrenByParent,
} from '@/lib/nodeTree';
import type { GraphEdge, GraphNode } from '@/types/graph';

function node(
	id: string,
	url: string,
	status: GraphNode['status'] = 'success',
	label?: string,
): GraphNode {
	return {
		id,
		urlNormalized: url,
		label: label ?? url,
		position: { x: 0, y: 0 },
		nodeSettings: {},
		crawlExclude: false,
		status,
	};
}

describe('buildNodeFlatTree', () => {
	it('全 status のノードを含め、シードから BFS で親子を決める', () => {
		const nodes = [
			node('a', 'https://example.com/'),
			node('b', 'https://example.com/a'),
			node('c', 'https://example.com/b', 'idle'),
		];
		const edges: GraphEdge[] = [
			{ id: 'e1', source: 'a', target: 'b' },
			{ id: 'e2', source: 'b', target: 'c' },
		];
		const flat = buildNodeFlatTree(nodes, edges, 'https://example.com/');
		expect(flat.map((n) => n.id).sort()).toEqual(['a', 'b', 'c']);
		expect(flat.find((n) => n.id === 'b')?.parent_id).toBe('a');
		expect(flat.find((n) => n.id === 'c')?.parent_id).toBe('b');
	});

	it('到達できないノードはルートになる', () => {
		const nodes = [
			node('a', 'https://example.com/'),
			node('orphan', 'https://other.example/'),
		];
		const edges: GraphEdge[] = [];
		const flat = buildNodeFlatTree(nodes, edges, 'https://example.com/');
		expect(flat.find((n) => n.id === 'orphan')?.parent_id).toBeNull();
	});
});

describe('filterNodeTree', () => {
	const flat = buildNodeFlatTree(
		[
			node('a', 'https://example.com/', 'success', 'Home'),
			node('b', 'https://example.com/docs', 'error', 'Docs'),
			node('c', 'https://example.com/about', 'idle', 'About'),
		],
		[
			{ id: 'e1', source: 'a', target: 'b' },
			{ id: 'e2', source: 'a', target: 'c' },
		],
		'https://example.com/',
	);

	it('フィルタなしでは全件 visible、ヒットなし', () => {
		const result = filterNodeTree(flat, '', []);
		expect(result.visibleIds.size).toBe(3);
		expect(result.hitCount).toBe(0);
		expect(result.hitIds.size).toBe(0);
	});

	it('テキスト一致ノードと祖先を残し、件数は実ヒットのみ', () => {
		const result = filterNodeTree(flat, 'docs', []);
		expect([...result.hitIds]).toEqual(['b']);
		expect(result.hitCount).toBe(1);
		expect(result.visibleIds.has('a')).toBe(true);
		expect(result.visibleIds.has('b')).toBe(true);
		expect(result.visibleIds.has('c')).toBe(false);
	});

	it('テキストと status は AND', () => {
		const result = filterNodeTree(flat, 'example', ['error']);
		expect([...result.hitIds]).toEqual(['b']);
		expect(result.hitCount).toBe(1);
	});
});

describe('expandIdsForHits / expandableNodeIds / groupChildrenByParent', () => {
	it('ヒット祖先を展開対象にし、子持ち ID と親子マップを返す', () => {
		const flat = [
			{
				id: 'a',
				parent_id: null,
				urlNormalized: 'https://a',
				label: 'a',
				status: 'success' as const,
			},
			{
				id: 'b',
				parent_id: 'a',
				urlNormalized: 'https://b',
				label: 'b',
				status: 'success' as const,
			},
			{
				id: 'c',
				parent_id: 'b',
				urlNormalized: 'https://c',
				label: 'c',
				status: 'error' as const,
			},
		];
		expect([...expandIdsForHits(flat, new Set(['c']))].sort()).toEqual([
			'a',
			'b',
		]);
		expect(expandableNodeIds(flat).sort()).toEqual(['a', 'b']);
		const byParent = groupChildrenByParent(flat);
		expect(byParent.get(null)?.map((n) => n.id)).toEqual(['a']);
		expect(byParent.get('a')?.map((n) => n.id)).toEqual(['b']);
	});
});
