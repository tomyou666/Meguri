import type { GraphEdge, GraphNode, NodeStatus } from '@/types/graph';

export type NodeFlatNode = {
	id: string;
	parent_id: string | null;
	urlNormalized: string;
	label: string;
	status: NodeStatus;
};

/** テキスト / status フィルタに必要なフラットノード最小形。 */
export type FilterableFlatNode = {
	id: string;
	parent_id: string | null;
	urlNormalized: string;
	label: string;
	/** status フィルタ時のみ必要（NodeStatus または同等の文字列） */
	status?: string;
};

export type NodeTreeFilterResult = {
	/** 表示するノード（ヒット + 祖先） */
	visibleIds: Set<string>;
	/** テキスト・status 条件を満たす実ヒット */
	hitIds: Set<string>;
	hitCount: number;
};

/** 全ノードを edges から BFS で親子付けしたフラットツリーを構築する。 */
export function buildNodeFlatTree(
	nodes: GraphNode[],
	edges: GraphEdge[],
	seedUrl?: string,
): NodeFlatNode[] {
	if (nodes.length === 0) return [];

	const visibleIds = new Set(nodes.map((n) => n.id));
	const adj = new Map<string, string[]>();
	for (const e of edges) {
		if (!visibleIds.has(e.source) || !visibleIds.has(e.target)) continue;
		const list = adj.get(e.source) ?? [];
		list.push(e.target);
		adj.set(e.source, list);
	}

	const seedNode =
		(seedUrl ? nodes.find((n) => n.urlNormalized === seedUrl) : undefined) ??
		nodes.find(
			(n) => !edges.some((e) => e.target === n.id && visibleIds.has(e.source)),
		);

	const parentById = new Map<string, string | null>();
	const queue: string[] = [];

	if (seedNode) {
		parentById.set(seedNode.id, null);
		queue.push(seedNode.id);
	}

	while (queue.length > 0) {
		const current = queue.shift();
		if (!current) continue;
		for (const child of adj.get(current) ?? []) {
			if (parentById.has(child)) continue;
			parentById.set(child, current);
			queue.push(child);
		}
	}

	for (const n of nodes) {
		if (!parentById.has(n.id)) {
			parentById.set(n.id, null);
		}
	}

	return nodes.map((n) => ({
		id: n.id,
		parent_id: parentById.get(n.id) ?? null,
		urlNormalized: n.urlNormalized,
		label: n.label,
		status: n.status,
	}));
}

/** parent_id ごとの子一覧（出現順）を返す。 */
export function groupChildrenByParent(
	flat: NodeFlatNode[],
): Map<string | null, NodeFlatNode[]> {
	const byParent = new Map<string | null, NodeFlatNode[]>();
	for (const node of flat) {
		const list = byParent.get(node.parent_id) ?? [];
		list.push(node);
		byParent.set(node.parent_id, list);
	}
	return byParent;
}

/** 子を持つノード ID を返す（展開/折りたたみ対象）。 */
export function expandableNodeIds(flat: NodeFlatNode[]): string[] {
	const parents = new Set<string>();
	for (const n of flat) {
		if (n.parent_id) parents.add(n.parent_id);
	}
	return [...parents];
}

function matchesQuery(node: FilterableFlatNode, query: string): boolean {
	if (!query) return true;
	const q = query.toLowerCase();
	return (
		node.label.toLowerCase().includes(q) ||
		node.urlNormalized.toLowerCase().includes(q)
	);
}

function matchesStatus(
	node: FilterableFlatNode,
	statuses: ReadonlySet<NodeStatus>,
): boolean {
	if (statuses.size === 0) return true;
	if (node.status === undefined) return false;
	return (statuses as ReadonlySet<string>).has(node.status);
}

/** テキスト検索と status フィルタ（AND）で visible / hit を計算する。 */
export function filterNodeTree(
	flat: FilterableFlatNode[],
	query: string,
	statuses: ReadonlyArray<NodeStatus>,
): NodeTreeFilterResult {
	const trimmed = query.trim();
	const statusSet = new Set(statuses);
	const hasFilter = trimmed.length > 0 || statusSet.size > 0;

	if (!hasFilter) {
		return {
			visibleIds: new Set(flat.map((n) => n.id)),
			hitIds: new Set(),
			hitCount: 0,
		};
	}

	const byId = new Map(flat.map((n) => [n.id, n]));
	const hitIds = new Set<string>();
	for (const n of flat) {
		if (matchesQuery(n, trimmed) && matchesStatus(n, statusSet)) {
			hitIds.add(n.id);
		}
	}

	const visibleIds = new Set<string>(hitIds);
	for (const id of hitIds) {
		let parentId = byId.get(id)?.parent_id ?? null;
		while (parentId) {
			if (visibleIds.has(parentId)) break;
			visibleIds.add(parentId);
			parentId = byId.get(parentId)?.parent_id ?? null;
		}
	}

	return { visibleIds, hitIds, hitCount: hitIds.size };
}

/** ヒット祖先をすべて展開した expanded 集合を返す。 */
export function expandIdsForHits(
	flat: Pick<FilterableFlatNode, 'id' | 'parent_id'>[],
	hitIds: ReadonlySet<string>,
): Set<string> {
	const byId = new Map(flat.map((n) => [n.id, n]));
	const expanded = new Set<string>();
	for (const id of hitIds) {
		let parentId = byId.get(id)?.parent_id ?? null;
		while (parentId) {
			expanded.add(parentId);
			parentId = byId.get(parentId)?.parent_id ?? null;
		}
	}
	return expanded;
}
