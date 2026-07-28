import type { NodeStatus } from '@/types/graph';

export type GraphNodeStatusRow = {
	nodeId: string;
	status: string;
	lastError?: string;
};

export type NodeStatusPatch = {
	nodeId: string;
	status: NodeStatus;
	lastError?: string;
};

const NODE_STATUSES: ReadonlySet<string> = new Set([
	'idle',
	'running',
	'success',
	'error',
	'skipped',
]);

function asNodeStatus(status: string): NodeStatus | null {
	if (!NODE_STATUSES.has(status)) return null;
	return status as NodeStatus;
}

/**
 * API の status 行を UI パッチへ変換する。
 * 未知 status は捨てる。error 以外は lastError をクリアする。
 */
export function buildNodeStatusPatches(
	rows: GraphNodeStatusRow[],
): NodeStatusPatch[] {
	const out: NodeStatusPatch[] = [];
	for (const row of rows) {
		const status = asNodeStatus(row.status);
		if (!status) continue;
		const patch: NodeStatusPatch = { nodeId: row.nodeId, status };
		if (status === 'error') {
			patch.lastError = row.lastError || undefined;
		} else {
			patch.lastError = undefined;
		}
		out.push(patch);
	}
	return out;
}

/** UI 上 running のノード ID を集める。 */
export function collectRunningNodeIds(
	nodes: ReadonlyArray<{ id: string; status: NodeStatus }>,
): string[] {
	return nodes.filter((n) => n.status === 'running').map((n) => n.id);
}

/**
 * result なしの nodeSucceeded 相当の status パッチ。
 * Event 経由では本文を埋めず status のみ success にする。
 */
export function nodeSucceededStatusPatch(nodeId: string): NodeStatusPatch {
	return { nodeId, status: 'success', lastError: undefined };
}
