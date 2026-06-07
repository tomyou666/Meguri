import type { GraphEdge, GraphNode } from '@/types/graph';

export function hasChildNodes(nodeId: string, edges: GraphEdge[]): boolean {
	return edges.some((e) => e.source === nodeId);
}

export function getOutgoingEdges(
	nodeId: string,
	edges: GraphEdge[],
): GraphEdge[] {
	return edges.filter((e) => e.source === nodeId);
}

export function getDescendantNodeIds(
	rootId: string,
	edges: GraphEdge[],
): Set<string> {
	const descendants = new Set<string>();
	const queue = [rootId];
	while (queue.length > 0) {
		const current = queue.shift()!;
		for (const edge of getOutgoingEdges(current, edges)) {
			if (!descendants.has(edge.target) && edge.target !== rootId) {
				descendants.add(edge.target);
				queue.push(edge.target);
			}
		}
	}
	return descendants;
}

/** モード3: 選択ノードから有向に到達可能な既存ノード（BFS順） */
export function getForwardReachableExisting(
	startId: string,
	nodes: GraphNode[],
	edges: GraphEdge[],
): string[] {
	const nodeIds = new Set(nodes.map((n) => n.id));
	const visited = new Set<string>();
	const order: string[] = [];
	const queue = [startId];
	visited.add(startId);

	while (queue.length > 0) {
		const current = queue.shift()!;
		if (current !== startId) {
			order.push(current);
		}
		for (const edge of getOutgoingEdges(current, edges)) {
			if (nodeIds.has(edge.target) && !visited.has(edge.target)) {
				visited.add(edge.target);
				queue.push(edge.target);
			}
		}
	}
	return order;
}

/** 全ノードを BFS 順（seed から到達可能な順 + 孤立は末尾） */
export function getBfsNodeOrder(
	seedNodeId: string | undefined,
	nodes: GraphNode[],
	edges: GraphEdge[],
): string[] {
	if (!nodes.length) return [];
	const start =
		seedNodeId && nodes.some((n) => n.id === seedNodeId)
			? seedNodeId
			: nodes[0].id;
	const order: string[] = [];
	const visited = new Set<string>();
	const queue = [start];
	visited.add(start);
	while (queue.length > 0) {
		const id = queue.shift()!;
		order.push(id);
		for (const e of getOutgoingEdges(id, edges)) {
			if (!visited.has(e.target)) {
				visited.add(e.target);
				queue.push(e.target);
			}
		}
	}
	for (const n of nodes) {
		if (!visited.has(n.id)) order.push(n.id);
	}
	return order;
}

/** 折りたたみルートの子孫 ID（ルート自身は含まない） */
export function getHiddenDescendantIds(
	collapsedRootIds: string[],
	edges: GraphEdge[],
): Set<string> {
	const hidden = new Set<string>();
	for (const rootId of collapsedRootIds) {
		for (const id of getDescendantNodeIds(rootId, edges)) {
			hidden.add(id);
		}
	}
	return hidden;
}

export function isExcludedSubtree(
	nodeId: string,
	nodes: GraphNode[],
	edges: GraphEdge[],
	visited: Set<string> = new Set(),
): boolean {
	if (visited.has(nodeId)) return false;
	visited.add(nodeId);

	const n = nodes.find((x) => x.id === nodeId);
	if (!n) return false;
	if (n.crawlExclude) return true;
	for (const e of edges) {
		if (e.target === nodeId) {
			if (isExcludedSubtree(e.source, nodes, edges, visited)) return true;
		}
	}
	return false;
}

export function collectDescendantUrls(
	rootId: string,
	nodes: GraphNode[],
	edges: GraphEdge[],
): string[] {
	const desc = getDescendantNodeIds(rootId, edges);
	const root = nodes.find((n) => n.id === rootId);
	const urls: string[] = root ? [root.urlNormalized] : [];
	for (const id of desc) {
		const n = nodes.find((node) => node.id === id);
		if (n) urls.push(n.urlNormalized);
	}
	return urls;
}
