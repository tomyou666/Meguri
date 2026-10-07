import { describe, expect, it } from 'vitest';
import {
	buildInitialFlatTree,
	computeSemiCheckedIds,
	type ExportFlatNode,
	initialCheckedIds,
	parseExportSeparator,
	preorderNodeIds,
	resolveExportSeparator,
	toggleExportNodeCheck,
} from '@/lib/exportTree';
import type { GraphEdge, GraphNode } from '@/types/graph';

function node(
	id: string,
	url: string,
	status: GraphNode['status'] = 'success',
	label?: string,
	crawlExclude = false,
): GraphNode {
	return {
		id,
		urlNormalized: url,
		label: label ?? url,
		position: { x: 0, y: 0 },
		nodeSettings: {},
		crawlExclude,
		status,
	};
}

function flatNode(
	partial: Pick<ExportFlatNode, 'id' | 'parent_id'> &
		Partial<Omit<ExportFlatNode, 'id' | 'parent_id'>>,
): ExportFlatNode {
	return {
		urlNormalized: `https://${partial.id}`,
		label: partial.id,
		status: 'success',
		crawlExclude: false,
		...partial,
	};
}

describe('buildInitialFlatTree', () => {
	it('シードから BFS で親子を決め、success のみ含める', () => {
		const nodes = [
			node('a', 'https://example.com/'),
			node('b', 'https://example.com/a'),
			node('c', 'https://example.com/b', 'idle'),
		];
		const edges: GraphEdge[] = [
			{ id: 'e1', source: 'a', target: 'b' },
			{ id: 'e2', source: 'b', target: 'c' },
		];
		const flat = buildInitialFlatTree(
			nodes,
			edges,
			'https://example.com/',
			'all',
			[],
		);
		expect(flat.map((n) => n.id)).toEqual(['a', 'b']);
		expect(flat.find((n) => n.id === 'b')?.parent_id).toBe('a');
	});

	it('選択モードでは選択ノードのみ、親未選択はルート化する', () => {
		const nodes = [
			node('a', 'https://example.com/'),
			node('b', 'https://example.com/a'),
			node('c', 'https://example.com/c'),
		];
		const edges: GraphEdge[] = [
			{ id: 'e1', source: 'a', target: 'b' },
			{ id: 'e2', source: 'a', target: 'c' },
		];
		const flat = buildInitialFlatTree(
			nodes,
			edges,
			'https://example.com/',
			'selected',
			['b', 'c'],
		);
		expect(flat).toHaveLength(2);
		expect(flat.every((n) => n.parent_id === null)).toBe(true);
	});
});

describe('initialCheckedIds', () => {
	it('success かつ非 crawlExclude のみ初期 ON', () => {
		const flat = [
			flatNode({ id: 'ok', parent_id: null }),
			flatNode({ id: 'ex', parent_id: null, crawlExclude: true }),
			flatNode({ id: 'err', parent_id: null, status: 'error' }),
			flatNode({ id: 'skip', parent_id: null, status: 'skipped' }),
			flatNode({ id: 'idle', parent_id: null, status: 'idle' }),
		];
		expect(initialCheckedIds(flat)).toEqual(['ok']);
	});

	it('親が crawlExclude でも success の子は初期 ON', () => {
		const flat = [
			flatNode({ id: 'parent', parent_id: null, crawlExclude: true }),
			flatNode({ id: 'child', parent_id: 'parent' }),
		];
		expect(initialCheckedIds(flat)).toEqual(['child']);
	});
});

describe('preorderNodeIds', () => {
	it('チェック ON のノードのみ深さ優先で返す', () => {
		const flat = [
			{
				id: 'a',
				parent_id: null,
				urlNormalized: 'https://a',
				label: 'a',
				status: 'success',
				crawlExclude: false,
			},
			{
				id: 'b',
				parent_id: 'a',
				urlNormalized: 'https://b',
				label: 'b',
				status: 'success',
				crawlExclude: false,
			},
			{
				id: 'c',
				parent_id: 'a',
				urlNormalized: 'https://c',
				label: 'c',
				status: 'success',
				crawlExclude: false,
			},
		];
		expect(preorderNodeIds(flat, ['a', 'c'])).toEqual(['a', 'c']);
	});

	it('親未チェックでも子のみチェック ON なら子を含める', () => {
		const flat = [
			{
				id: 'a',
				parent_id: null,
				urlNormalized: 'https://a',
				label: 'a',
				status: 'success',
				crawlExclude: false,
			},
			{
				id: 'b',
				parent_id: 'a',
				urlNormalized: 'https://b',
				label: 'b',
				status: 'success',
				crawlExclude: false,
			},
		];
		expect(preorderNodeIds(flat, ['b'])).toEqual(['b']);
	});
});

describe('parseExportSeparator', () => {
	it('\\r\\n \\n \\t を制御文字に変換する', () => {
		expect(parseExportSeparator('\\r\\n')).toBe('\r\n');
		expect(parseExportSeparator('a\\nb')).toBe('a\nb');
		expect(parseExportSeparator('\\t')).toBe('\t');
		expect(parseExportSeparator('\\\\')).toBe('\\');
	});
});

describe('resolveExportSeparator', () => {
	it('HTML 形式では区切り文字をエスケープする', () => {
		expect(resolveExportSeparator('<script>x</script>', 'html')).toBe(
			'&lt;script&gt;x&lt;/script&gt;',
		);
	});

	it('markdown ではエスケープシーケンスだけ展開する', () => {
		expect(resolveExportSeparator('\\n---\\n', 'markdown')).toBe('\n---\n');
	});
});

describe('toggleExportNodeCheck', () => {
	const flat = [
		{
			id: 'a',
			parent_id: null,
			urlNormalized: 'https://a',
			label: 'a',
			status: 'success',
			crawlExclude: false,
		},
		{
			id: 'b',
			parent_id: 'a',
			urlNormalized: 'https://b',
			label: 'b',
			status: 'success',
			crawlExclude: false,
		},
	];

	it('子を OFF にしても親は ON のまま', () => {
		const next = toggleExportNodeCheck(flat, ['a', 'b'], 'b', false);
		expect(next).toEqual(['a']);
	});

	it('親を ON にすると配下も ON', () => {
		const next = toggleExportNodeCheck(flat, [], 'a', true);
		expect(next.sort()).toEqual(['a', 'b']);
	});

	it('cascade=false ではクリックしたノードのみ切り替える', () => {
		const next = toggleExportNodeCheck(flat, ['a', 'b'], 'b', false, false);
		expect(next).toEqual(['a']);
	});
});

describe('computeSemiCheckedIds', () => {
	const flat = [
		{
			id: 'a',
			parent_id: null,
			urlNormalized: 'https://a',
			label: 'a',
			status: 'success',
			crawlExclude: false,
		},
		{
			id: 'b',
			parent_id: 'a',
			urlNormalized: 'https://b',
			label: 'b',
			status: 'success',
			crawlExclude: false,
		},
		{
			id: 'c',
			parent_id: 'a',
			urlNormalized: 'https://c',
			label: 'c',
			status: 'success',
			crawlExclude: false,
		},
	];

	it('子の一部のみ ON の親を返す', () => {
		expect(computeSemiCheckedIds(flat, ['b'])).toEqual(['a']);
	});

	it('子がすべて ON の親は含めない', () => {
		expect(computeSemiCheckedIds(flat, ['b', 'c'])).toEqual([]);
	});
});
