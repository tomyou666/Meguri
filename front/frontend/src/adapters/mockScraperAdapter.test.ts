import { describe, expect, it } from 'vitest';
import { canonicalizeLinksJson } from '@/lib/contentHash';
import type { DbNodeResult } from '@/types/db';
import { MockScraperAdapter } from './mockScraperAdapter';

describe('MockScraperAdapter', () => {
	it('creates crawl_runs on startCrawl path via snapshot', async () => {
		const adapter = new MockScraperAdapter();
		const wsId = 'ws-test';
		adapter['workspaces'].set(wsId, {
			id: wsId,
			name: 'T',
			seedUrl: 'https://example.com/',
			settings: {},
			exclude_urls: [],
			nodes: [],
			edges: [],
			graphLayoutDirection: 'LR',
			domainSettings: {},
		});
		adapter['results'].set(wsId, []);
		const runId = await adapter.saveResultsSnapshot(wsId);
		expect(adapter.getCrawlRuns(wsId).some((r) => r.id === runId)).toBe(true);
	});

	it('getWorkspaceDiff uses links_json only', async () => {
		const adapter = new MockScraperAdapter();
		const wsId = 'ws-diff';
		const baselineRunId = 'baseline-run';
		adapter['workspaces'].set(wsId, {
			id: wsId,
			name: 'D',
			seedUrl: 'https://example.com/',
			settings: {},
			exclude_urls: [],
			nodes: [
				{
					id: 'n1',
					urlNormalized: 'https://example.com/',
					label: 'root',
					position: { x: 0, y: 0 },
					nodeSettings: {},
					crawlExclude: false,
					status: 'success',
				},
			],
			edges: [],
			graphLayoutDirection: 'LR',
			domainSettings: {},
			baselineRunId,
		});
		const baseRow: DbNodeResult = {
			id: 'b1',
			run_id: baselineRunId,
			workspace_id: wsId,
			node_id: 'n1',
			url: 'https://example.com/',
			markdown: 'same',
			html: null,
			raw_html: null,
			json_body: null,
			links_json: JSON.stringify(['https://example.com/a']),
			metadata_json: null,
			error: null,
			fetched_at: '2020-01-01T00:00:00.000Z',
			content_hash: 'abc',
		};
		const curRow: DbNodeResult = {
			...baseRow,
			id: 'c1',
			run_id: 'run2',
			links_json: JSON.stringify(['https://example.com/b']),
			fetched_at: '2020-01-02T00:00:00.000Z',
			content_hash: 'abc',
		};
		adapter['results'].set(wsId, [baseRow, curRow]);

		const diff = await adapter.getWorkspaceDiff(wsId);
		expect(diff.nodes[0]?.kinds).toContain('links');
		expect(diff.nodes[0]?.kinds).not.toContain('content');
	});

	it('deleteResults keeps older history', async () => {
		const adapter = new MockScraperAdapter();
		const wsId = 'ws-del';
		adapter['workspaces'].set(wsId, {
			id: wsId,
			name: 'X',
			seedUrl: 'https://example.com/',
			settings: {},
			exclude_urls: [],
			nodes: [],
			edges: [],
			graphLayoutDirection: 'LR',
			domainSettings: {},
		});
		adapter['results'].set(wsId, [
			{
				id: 'new',
				run_id: 'r1',
				workspace_id: wsId,
				node_id: 'n1',
				url: 'https://example.com/',
				markdown: 'x',
				html: null,
				raw_html: null,
				json_body: null,
				links_json: null,
				metadata_json: null,
				error: null,
				fetched_at: '2020-01-02',
				content_hash: null,
			},
			{
				id: 'old',
				run_id: 'r1',
				workspace_id: wsId,
				node_id: 'n1',
				url: 'https://example.com/',
				markdown: 'y',
				html: null,
				raw_html: null,
				json_body: null,
				links_json: null,
				metadata_json: null,
				error: null,
				fetched_at: '2020-01-01',
				content_hash: null,
			},
		]);
		await adapter.deleteResults(wsId, ['n1']);
		const rows = adapter.getNodeResultsRows(wsId);
		expect(rows).toHaveLength(1);
		expect(rows[0]?.id).toBe('old');
	});
});

describe('canonicalizeLinksJson', () => {
	it('sorts URLs for stable comparison', () => {
		expect(canonicalizeLinksJson(['b', 'a'])).toBe('["a","b"]');
	});
});
