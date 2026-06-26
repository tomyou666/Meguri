import { describe, expect, it } from 'vitest';
import {
	diffNodeCount,
	filterNodesByKind,
	summaryBadgeLabels,
} from '@/components/diff/diffSummaryUtils';
import type { WorkspaceDiff } from '@/types/adapter';

const sampleDiff: WorkspaceDiff = {
	workspaceId: 'ws-1',
	hasDiff: true,
	baselineRunId: 'run-1',
	nodes: [
		{ nodeId: 'n1', url: 'https://a', kinds: ['content', 'links'] },
		{ nodeId: 'n2', url: 'https://b', kinds: ['fetch'] },
	],
	summary: { content: 1, links: 1, fetch: 1 },
};

// DiffSummarySheet が使う件数・バッジ表示ロジックを検証する。
describe('DiffSummarySheet helpers', () => {
	it('diffNodeCount: ノード数を返す（複数 kind でも 1 件）', () => {
		expect(diffNodeCount(sampleDiff)).toBe(2);
		expect(diffNodeCount(undefined)).toBe(0);
	});

	it('summaryBadgeLabels: summary 件数をバッジ文言にする', () => {
		expect(summaryBadgeLabels(sampleDiff.summary)).toEqual({
			content: 'content 1',
			links: 'links 1',
			fetch: 'fetch 1',
		});
	});

	it('filterNodesByKind: kind フィルタでノードを絞る', () => {
		expect(filterNodesByKind(sampleDiff.nodes, 'all')).toHaveLength(2);
		expect(filterNodesByKind(sampleDiff.nodes, 'content')).toHaveLength(1);
		expect(filterNodesByKind(sampleDiff.nodes, 'fetch')).toHaveLength(1);
	});
});
