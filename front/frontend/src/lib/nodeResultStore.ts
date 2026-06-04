import type { DbNodeResult } from '@/types/db';

/** ノードごとの node_results 履歴上限（Grill 確定の実行履歴 20 件と同数）。 */
export const MAX_NODE_RESULT_HISTORY = 20;

/** WS ごとの crawl_runs 履歴上限（UI runHistory と同数）。 */
export const MAX_CRAWL_RUN_HISTORY = 20;

export function latestByNode(rows: DbNodeResult[]): Map<string, DbNodeResult> {
	const map = new Map<string, DbNodeResult>();
	for (const row of rows) {
		const prev = map.get(row.node_id);
		if (!prev || row.fetched_at > prev.fetched_at) {
			map.set(row.node_id, row);
		}
	}
	return map;
}

/** 表示・マージ用: ノードごとの最新成功行（error IS NULL）。 */
export function latestSuccessByNode(
	rows: DbNodeResult[],
): Map<string, DbNodeResult> {
	const map = new Map<string, DbNodeResult>();
	for (const row of rows) {
		if (row.error) continue;
		const prev = map.get(row.node_id);
		if (!prev || row.fetched_at > prev.fetched_at) {
			map.set(row.node_id, row);
		}
	}
	return map;
}

export function rowsForRun(
	rows: DbNodeResult[],
	runId: string,
): Map<string, DbNodeResult> {
	const map = new Map<string, DbNodeResult>();
	for (const row of rows) {
		if (row.run_id !== runId) continue;
		map.set(row.node_id, row);
	}
	return map;
}

export function trimNodeResults(rows: DbNodeResult[]): DbNodeResult[] {
	const byNode = new Map<string, DbNodeResult[]>();
	for (const row of rows) {
		const list = byNode.get(row.node_id) ?? [];
		list.push(row);
		byNode.set(row.node_id, list);
	}
	const out: DbNodeResult[] = [];
	for (const list of byNode.values()) {
		const sorted = [...list].sort((a, b) =>
			b.fetched_at.localeCompare(a.fetched_at),
		);
		out.push(...sorted.slice(0, MAX_NODE_RESULT_HISTORY));
	}
	return out;
}

export function appendNodeResult(
	rows: DbNodeResult[],
	row: DbNodeResult,
): DbNodeResult[] {
	return trimNodeResults([...rows, row]);
}

/** 指定ノードの fetched_at 最新 1 行のみ削除。 */
export function deleteLatestResults(
	rows: DbNodeResult[],
	nodeIds: string[],
): DbNodeResult[] {
	const removeIds = new Set<string>();
	for (const nodeId of nodeIds) {
		const latest = rows
			.filter((r) => r.node_id === nodeId)
			.sort((a, b) => b.fetched_at.localeCompare(a.fetched_at))[0];
		if (latest) removeIds.add(latest.id);
	}
	return rows.filter((r) => !removeIds.has(r.id));
}
