import dagre from '@dagrejs/dagre';
import { Position } from '@xyflow/react';
import type { GraphEdge, GraphNode } from '@/types/graph';

export type DagreLayoutDirection = 'TB' | 'LR';

/** UrlNode の固定サイズ（表示と dagre レイアウトで共有） */
export const DAGRE_NODE_WIDTH = 260;
/** 詳細展開時のノード表示幅（レイアウト計算は DAGRE_NODE_WIDTH のまま） */
export const NODE_DETAIL_EXPANDED_WIDTH = 300;
export const DAGRE_NODE_HEIGHT = 40;
const RANK_SEP = 24;
const NODE_SEP = 8;
const DISCOVERED_NODE_SEP = 48;

/**
 * ノード・エッジ全体に dagre を適用し、各ノードの左上座標を返す。
 */
export function handlePositionsForDirection(direction: DagreLayoutDirection): {
	source: Position;
	target: Position;
} {
	const isHorizontal = direction === 'LR';
	return {
		target: isHorizontal ? Position.Left : Position.Top,
		source: isHorizontal ? Position.Right : Position.Bottom,
	};
}

export function computeDagrePositions(
	nodes: GraphNode[],
	edges: GraphEdge[],
	direction: DagreLayoutDirection = 'LR',
	ranksep: number = RANK_SEP,
	nodesep: number = NODE_SEP,
): Map<string, { x: number; y: number }> {
	const g = new dagre.graphlib.Graph();
	g.setDefaultEdgeLabel(() => ({}));
	g.setGraph({
		rankdir: direction,
		ranksep: ranksep,
		nodesep: nodesep,
	});

	for (const node of nodes) {
		g.setNode(node.id, {
			width: DAGRE_NODE_WIDTH,
			height: DAGRE_NODE_HEIGHT,
		});
	}
	for (const edge of edges) {
		g.setEdge(edge.source, edge.target);
	}

	dagre.layout(g);

	const positions = new Map<string, { x: number; y: number }>();
	for (const node of nodes) {
		const laid = g.node(node.id);
		positions.set(node.id, {
			x: laid.x - DAGRE_NODE_WIDTH / 2,
			y: laid.y - DAGRE_NODE_HEIGHT / 2,
		});
	}
	return positions;
}

/**
 * 再生で追加されたノード用: 全体レイアウトを計算し、対象 ID の位置だけ返す。
 * 既存ノードの store 上の位置は変更しない。
 */
export function positionForDiscoveredNode(
	nodes: GraphNode[],
	edges: GraphEdge[],
	nodeId: string,
	fallback: { x: number; y: number },
	direction: DagreLayoutDirection = 'LR',
): { x: number; y: number } {
	if (nodes.length === 0) return fallback;
	const positions = computeDagrePositions(
		nodes,
		edges,
		direction,
		RANK_SEP,
		DISCOVERED_NODE_SEP,
	);
	return positions.get(nodeId) ?? fallback;
}

export function fallbackNearParent(
	parent: GraphNode,
	direction: DagreLayoutDirection,
): { x: number; y: number } {
	if (direction === 'LR') {
		return { x: parent.position.x + 280, y: parent.position.y };
	}
	return { x: parent.position.x, y: parent.position.y + 120 };
}
