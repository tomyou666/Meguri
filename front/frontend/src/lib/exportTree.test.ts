import { describe, expect, it } from 'vitest';
import {
	buildExportPreview,
	buildInitialFlatTree,
	mergeExportContent,
	parseExportSeparator,
	preorderNodeIds,
	toggleExportNodeCheck,
} from '@/lib/exportTree';
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

describe('preorderNodeIds', () => {
	it('チェック ON のノードのみ深さ優先で返す', () => {
		const flat = [
			{
				id: 'a',
				parent_id: null,
				urlNormalized: 'https://a',
				label: 'a',
				status: 'success',
			},
			{
				id: 'b',
				parent_id: 'a',
				urlNormalized: 'https://b',
				label: 'b',
				status: 'success',
			},
			{
				id: 'c',
				parent_id: 'a',
				urlNormalized: 'https://c',
				label: 'c',
				status: 'success',
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
			},
			{
				id: 'b',
				parent_id: 'a',
				urlNormalized: 'https://b',
				label: 'b',
				status: 'success',
			},
		];
		expect(preorderNodeIds(flat, ['b'])).toEqual(['b']);
	});
});

describe('mergeExportContent', () => {
	it('見出し・区切り・markdown を反映する', () => {
		const content = mergeExportContent({
			results: [
				{ url: 'https://a', markdown: 'body a' },
				{ url: 'https://b', markdown: 'body b' },
			],
			nodeMeta: [
				{ id: '1', urlNormalized: 'https://a', label: 'A' },
				{ id: '2', urlNormalized: 'https://b', label: 'B' },
			],
			settings: {
				format: 'markdown',
				separator: '\n---\n',
				includeHeading: true,
				headingField: 'url',
			},
		});
		expect(content).toContain('## https://a');
		expect(content).toContain('body a');
		expect(content).toContain('---');
	});

	it('見出し OFF では本文のみ連結する', () => {
		const content = mergeExportContent({
			results: [{ url: 'https://a', markdown: 'only body' }],
			nodeMeta: [{ id: '1', urlNormalized: 'https://a', label: 'A' }],
			settings: {
				format: 'markdown',
				separator: '\n',
				includeHeading: false,
				headingField: 'label',
			},
		});
		expect(content).toBe('only body');
	});

	it('区切り文字のエスケープシーケンスを展開する', () => {
		const content = mergeExportContent({
			results: [
				{ url: 'https://a', markdown: 'a' },
				{ url: 'https://b', markdown: 'b' },
			],
			nodeMeta: [
				{ id: '1', urlNormalized: 'https://a', label: 'A' },
				{ id: '2', urlNormalized: 'https://b', label: 'B' },
			],
			settings: {
				format: 'markdown',
				separator: '\\n---\\n',
				includeHeading: false,
				headingField: 'url',
			},
		});
		expect(content).toBe('a\n---\nb');
	});

	it('HTML 形式では区切り文字をエスケープする', () => {
		const content = mergeExportContent({
			results: [
				{ url: 'https://a', html: '<p>a</p>' },
				{ url: 'https://b', html: '<p>b</p>' },
			],
			nodeMeta: [
				{ id: '1', urlNormalized: 'https://a', label: 'A' },
				{ id: '2', urlNormalized: 'https://b', label: 'B' },
			],
			settings: {
				format: 'html',
				separator: '<script>x</script>',
				includeHeading: false,
				headingField: 'url',
			},
		});
		expect(content).toBe('<p>a</p>&lt;script&gt;x&lt;/script&gt;<p>b</p>');
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

describe('toggleExportNodeCheck', () => {
	const flat = [
		{
			id: 'a',
			parent_id: null,
			urlNormalized: 'https://a',
			label: 'a',
			status: 'success',
		},
		{
			id: 'b',
			parent_id: 'a',
			urlNormalized: 'https://b',
			label: 'b',
			status: 'success',
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
});

describe('buildExportPreview', () => {
	it('結果なしノードをスキップして件数を返す', () => {
		const flat = [
			{
				id: 'a',
				parent_id: null,
				urlNormalized: 'https://a',
				label: 'a',
				status: 'success',
			},
			{
				id: 'b',
				parent_id: null,
				urlNormalized: 'https://b',
				label: 'b',
				status: 'success',
			},
		];
		const preview = buildExportPreview(
			['a', 'b'],
			flat,
			[{ url: 'https://a', markdown: 'ok' }],
			{
				format: 'markdown',
				separator: '\n',
				includeHeading: false,
				headingField: 'url',
			},
		);
		expect(preview.includedCount).toBe(1);
		expect(preview.skippedCount).toBe(1);
		expect(preview.content).toBe('ok');
	});
});
