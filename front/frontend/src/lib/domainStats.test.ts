import { describe, expect, it } from 'vitest';
import type { GraphNode } from '@/types/graph';
import {
	countNodesByStatus,
	domainStatusKey,
	groupNodesByHost,
	isRobotsCacheHit,
	robotsTargetsFromNodes,
	robotsTargetsKey,
} from './domainStats';

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

	it('position 変更だけでは robotsTargetsKey / domainStatusKey が変わらない', () => {
		const moved = nodes.map((n) =>
			n.id === '1' ? { ...n, position: { x: 999, y: 888 } } : n,
		);
		const targets = robotsTargetsFromNodes(nodes);
		const movedTargets = robotsTargetsFromNodes(moved);
		expect(robotsTargetsKey(movedTargets)).toBe(robotsTargetsKey(targets));
		expect(domainStatusKey(moved)).toBe(domainStatusKey(nodes));
	});

	it('URL 追加や status 変更で domainStatusKey が変わる', () => {
		const baseKey = domainStatusKey(nodes);
		const added: GraphNode = {
			id: '4',
			urlNormalized: 'https://c.example/',
			label: 'd',
			position: { x: 0, y: 0 },
			nodeSettings: {},
			crawlExclude: false,
			status: 'idle',
		};
		expect(domainStatusKey([...nodes, added])).not.toBe(baseKey);
		const statusChanged = nodes.map((n) =>
			n.id === '1' ? { ...n, status: 'running' as const } : n,
		);
		expect(domainStatusKey(statusChanged)).not.toBe(baseKey);
	});

	it('host 追加で robotsTargetsKey が変わる', () => {
		const targets = robotsTargetsFromNodes(nodes);
		const added: GraphNode = {
			id: '4',
			urlNormalized: 'https://c.example/',
			label: 'd',
			position: { x: 0, y: 0 },
			nodeSettings: {},
			crawlExclude: false,
			status: 'idle',
		};
		const withNewHost = robotsTargetsFromNodes([...nodes, added]);
		expect(robotsTargetsKey(withNewHost)).not.toBe(robotsTargetsKey(targets));
	});

	it('loading は robots キャッシュヒットにしない', () => {
		expect(isRobotsCacheHit(undefined)).toBe(false);
		expect(isRobotsCacheHit({ status: 'loading' })).toBe(false);
		expect(isRobotsCacheHit({ status: 'found' })).toBe(true);
		expect(isRobotsCacheHit({ status: 'not_found' })).toBe(true);
		expect(isRobotsCacheHit({ status: 'error' })).toBe(true);
	});
});
