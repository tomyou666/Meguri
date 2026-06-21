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

/** robots 取得対象（host → baseURL）を position 非依存で構築する。 */
export function robotsTargetsFromNodes(
	nodes: GraphNode[],
): Map<string, string> {
	const hosts = [...groupNodesByHost(nodes).keys()].sort();
	const targets = new Map<string, string>();
	for (const host of hosts) {
		targets.set(host, baseURLForHost(host, nodes));
	}
	return targets;
}

/** robots 取得対象の安定キー（position 変更では変わらない）。 */
export function robotsTargetsKey(targets: Map<string, string>): string {
	return [...targets.entries()]
		.sort(([a], [b]) => a.localeCompare(b))
		.map(([host, baseURL]) => `${host}\0${baseURL}`)
		.join('\n');
}

/** ドメインステータス集計用の安定キー（position 変更では変わらない）。 */
export function domainStatusKey(nodes: GraphNode[]): string {
	return [...nodes]
		.map((n) => `${n.id}\0${n.urlNormalized}\0${n.status}`)
		.sort()
		.join('\n');
}

export type RobotsCacheEntry = {
	status: 'loading' | 'found' | 'not_found' | 'error';
};

/** loading 以外のみ robots セッションキャッシュとして扱う。 */
export function isRobotsCacheHit(info?: RobotsCacheEntry): boolean {
	return !!info && info.status !== 'loading';
}
