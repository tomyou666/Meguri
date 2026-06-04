import { describe, expect, it } from 'vitest';
import type { DbNodeResult } from '@/types/db';
import {
	appendNodeResult,
	deleteLatestResults,
	MAX_NODE_RESULT_HISTORY,
} from './nodeResultStore';

function row(nodeId: string, fetchedAt: string, id?: string): DbNodeResult {
	return {
		id: id ?? `${nodeId}-${fetchedAt}`,
		run_id: 'run1',
		workspace_id: 'ws1',
		node_id: nodeId,
		url: `https://example.com/${nodeId}`,
		markdown: null,
		html: null,
		raw_html: null,
		json_body: null,
		links_json: null,
		metadata_json: null,
		error: null,
		fetched_at: fetchedAt,
		content_hash: null,
	};
}

describe('nodeResultStore', () => {
	it('keeps at most MAX_NODE_RESULT_HISTORY per node', () => {
		let rows: DbNodeResult[] = [];
		for (let i = 0; i < MAX_NODE_RESULT_HISTORY + 5; i++) {
			rows = appendNodeResult(
				rows,
				row('n1', `2020-01-${String(i + 1).padStart(2, '0')}`),
			);
		}
		expect(rows.filter((r) => r.node_id === 'n1').length).toBe(
			MAX_NODE_RESULT_HISTORY,
		);
	});

	it('deleteLatestResults removes only the newest row per node', () => {
		const rows = [
			row('n1', '2020-01-02', 'new'),
			row('n1', '2020-01-01', 'old'),
			row('n2', '2020-01-03', 'n2new'),
		];
		const next = deleteLatestResults(rows, ['n1']);
		expect(next.map((r) => r.id).sort()).toEqual(['n2new', 'old']);
	});
});
