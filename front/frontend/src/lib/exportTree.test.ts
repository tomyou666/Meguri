import { describe, expect, it } from 'vitest';
import {
	buildExportPreview,
	buildExportPreviewSections,
	buildInitialFlatTree,
	buildSplitExportFiles,
	computeSemiCheckedIds,
	mergeExportContent,
	parseExportSeparator,
	preorderNodeIds,
	sanitizeExportFileName,
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
				splitSave: false,
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
				splitSave: false,
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
				splitSave: false,
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
				splitSave: false,
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

	it('子の一部のみ ON の親を返す', () => {
		expect(computeSemiCheckedIds(flat, ['b'])).toEqual(['a']);
	});

	it('子がすべて ON の親は含めない', () => {
		expect(computeSemiCheckedIds(flat, ['b', 'c'])).toEqual([]);
	});
});

describe('sanitizeExportFileName', () => {
	it('Windows 禁止文字を置換する', () => {
		expect(sanitizeExportFileName('a/b:c')).toBe('a_b_c');
	});
});

describe('buildSplitExportFiles', () => {
	it('ノードごとにユニークなファイル名と本文を返す', () => {
		const flat = [
			{
				id: 'a',
				parent_id: null,
				urlNormalized: 'https://example.com/page-a',
				label: 'A',
				status: 'success',
			},
			{
				id: 'b',
				parent_id: null,
				urlNormalized: 'https://example.com/page-b',
				label: 'B',
				status: 'success',
			},
		];
		const files = buildSplitExportFiles(
			['a', 'b'],
			flat,
			[
				{ url: 'https://example.com/page-a', markdown: 'body a' },
				{ url: 'https://example.com/page-b', markdown: 'body b' },
			],
			{
				format: 'markdown',
				separator: '\n',
				includeHeading: true,
				headingField: 'url',
				splitSave: true,
			},
		);
		expect(files).toHaveLength(2);
		expect(files[0]?.name).toBe('page-a.md');
		expect(files[1]?.name).toBe('page-b.md');
		expect(files[0]?.content).toContain('body a');
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
				splitSave: false,
			},
		);
		expect(preview.includedCount).toBe(1);
		expect(preview.skippedCount).toBe(1);
		expect(preview.content).toBe('ok');
	});
});

describe('buildExportPreviewSections', () => {
	it('結果なしノードをスキップし各セクションに baseUrl を付与する', () => {
		const flat = [
			{
				id: 'a',
				parent_id: null,
				urlNormalized: 'https://a.com/page1',
				label: 'a',
				status: 'success',
			},
			{
				id: 'b',
				parent_id: null,
				urlNormalized: 'https://b.com/page2',
				label: 'b',
				status: 'success',
			},
		];
		const sections = buildExportPreviewSections(
			['a', 'b'],
			flat,
			[
				{ url: 'https://a.com/page1', markdown: 'body a' },
				{ url: 'https://b.com/page2', markdown: 'body b' },
			],
			{
				format: 'markdown',
				separator: '\n',
				includeHeading: true,
				headingField: 'url',
				splitSave: false,
			},
		);
		expect(sections).toHaveLength(2);
		expect(sections[0]).toEqual({
			id: 'a',
			baseUrl: 'https://a.com/page1',
			body: '## https://a.com/page1\n\nbody a',
		});
		expect(sections[1]).toEqual({
			id: 'b',
			baseUrl: 'https://b.com/page2',
			body: '## https://b.com/page2\n\nbody b',
		});
	});

	it('結果なしノードはセクションに含めない', () => {
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
		const sections = buildExportPreviewSections(
			['a', 'b'],
			flat,
			[{ url: 'https://a', markdown: 'ok' }],
			{
				format: 'markdown',
				separator: '\n',
				includeHeading: false,
				headingField: 'url',
				splitSave: false,
			},
		);
		expect(sections).toHaveLength(1);
		expect(sections[0]?.baseUrl).toBe('https://a');
		expect(sections[0]?.body).toBe('ok');
	});
});
