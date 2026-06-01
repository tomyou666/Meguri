import type { GraphEdge, GraphNode } from '@/types/graph';

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
