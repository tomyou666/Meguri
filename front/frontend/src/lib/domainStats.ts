import { hostFromUrl } from '@/lib/normalizeUrl';
import type { GraphNode, NodeStatus } from '@/types/graph';

export type NodeStatusCounts = Record<NodeStatus, number>;

const EMPTY_COUNTS: NodeStatusCounts = {
	idle: 0,
	running: 0,
	success: 0,
	error: 0,
	skipped: 0,
};

/** ノードを host ごとにグループ化する。 */
export function groupNodesByHost(nodes: GraphNode[]): Map<string, GraphNode[]> {
	const map = new Map<string, GraphNode[]>();
	for (const node of nodes) {
		const host = hostFromUrl(node.urlNormalized);
		if (!host) continue;
		const list = map.get(host);
		if (list) list.push(node);
		else map.set(host, [node]);
	}
	return map;
}

/** ノード一覧をステータス別件数に集計する。 */
export function countNodesByStatus(nodes: GraphNode[]): NodeStatusCounts {
	const counts = { ...EMPTY_COUNTS };
	for (const node of nodes) {
		counts[node.status] += 1;
	}
	return counts;
}

/** host グループから scheme 推定用の代表 URL を返す。 */
export function baseURLForHost(host: string, nodes: GraphNode[]): string {
	return (
		nodes.find((n) => hostFromUrl(n.urlNormalized) === host)?.urlNormalized ??
		`https://${host}/`
	);
}
