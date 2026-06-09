import { describe, expect, it } from 'vitest';
import type { Workspace } from '@/types/workspace';
import { workspaceFromDb, workspaceToDb } from './dbMappers';

// DB バンドルと UI ワークスペース型の相互変換を検証する。
describe('dbMappers', () => {
	it('ワークスペースを DB 保存後に復元して主要フィールドを保持する', () => {
		const ws: Workspace = {
			id: 'ws1',
			name: 'Test',
			seedUrl: 'https://example.com/',
			settings: { crawl: { max_depth: 3 } },
			exclude_urls: [],
			nodes: [
				{
					id: 'n1',
					urlNormalized: 'https://example.com/',
					label: 'root',
					position: { x: 0, y: 0 },
					nodeSettings: {},
					crawlExclude: false,
					status: 'idle',
				},
			],
			edges: [],
			graphLayoutDirection: 'LR',
			domainSettings: {},
			collapsedNodeIds: ['n1'],
		};
		const bundle = workspaceToDb(ws);
		const back = workspaceFromDb(bundle);
		expect(back.id).toBe(ws.id);
		expect(back.settings.crawl?.max_depth).toBe(3);
		expect(back.collapsedNodeIds).toEqual(['n1']);
	});
});
