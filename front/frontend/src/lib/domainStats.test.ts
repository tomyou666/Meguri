import { describe, expect, it } from 'vitest';
import type { GraphNode } from '@/types/graph';
import { countNodesByStatus, groupNodesByHost } from './domainStats';

// host グループ化とステータス集計を検証する。
describe('domainStats', () => {
	const nodes: GraphNode[] = [
		{
			id: '1',
			urlNormalized: 'https://a.example/page',
			label: 'a',
			position: { x: 0, y: 0 },
			nodeSettings: {},
			crawlExclude: false,
			status: 'success',
		},
		{
			id: '2',
			urlNormalized: 'https://a.example/other',
			label: 'b',
			position: { x: 0, y: 0 },
			nodeSettings: {},
			crawlExclude: false,
			status: 'error',
		},
		{
			id: '3',
			urlNormalized: 'https://b.example/',
			label: 'c',
			position: { x: 0, y: 0 },
			nodeSettings: {},
			crawlExclude: false,
			status: 'idle',
		},
	];

	it('host ごとにノードをグループ化する', () => {
		const grouped = groupNodesByHost(nodes);
		expect(grouped.get('a.example')?.length).toBe(2);
		expect(grouped.get('b.example')?.length).toBe(1);
	});

	it('ステータス別件数を集計する', () => {
		const counts = countNodesByStatus(nodes.filter((n) => n.id !== '3'));
		expect(counts.success).toBe(1);
		expect(counts.error).toBe(1);
		expect(counts.idle).toBe(0);
	});
});
