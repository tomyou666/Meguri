import {
	addEdge,
	applyNodeChanges,
	Background,
	type Connection,
	type Edge,
	type Node,
	type OnEdgesDelete,
	type OnNodesChange,
	ReactFlow,
	SelectionMode,
	useEdgesState,
	useNodesState,
} from '@xyflow/react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { messages } from '@/i18n/messages';
import { handlePositionsForDirection } from '@/lib/dagreLayout';
import {
	getHiddenDescendantIds,
	hasChildNodes,
	isExcludedSubtree,
} from '@/lib/graph';
import '@xyflow/react/dist/style.css';
import { useAppStore } from '@/stores/appStore';
import type { WorkspaceDiff } from '@/types/adapter';
import { GRAPH_MIN_ZOOM, GraphCanvasControls } from './GraphCanvasControls';
import { GraphSelectionSync } from './GraphSelectionSync';
import { UrlNode, type UrlNodeData } from './UrlNode';

const nodeTypes = { urlNode: UrlNode };
const graphFitViewOptions = { padding: 0.2, minZoom: GRAPH_MIN_ZOOM };

export function CrawlGraph() {
	const proOptions = { hideAttribution: true };
	const ws = useAppStore((s) => s.getActiveWorkspace());
	const selectedNodeIds = useAppStore((s) => s.selectedNodeIds);
	const selectedDomain = useAppStore((s) => s.selectedDomain);
	const workspaceDiffCache = useAppStore((s) => s.workspaceDiffCache);
	const selectNode = useAppStore((s) => s.selectNode);
	const clearNodeSelection = useAppStore((s) => s.clearNodeSelection);
	const graphToolMode = useAppStore((s) => s.graphToolMode);
	const isSelectTool = graphToolMode === 'select';
	const updateNodePosition = useAppStore((s) => s.updateNodePosition);
	const removeEdges = useAppStore((s) => s.removeEdges);
	const addEdgeToStore = useAppStore((s) => s.addEdge);
	const openAddNodeDialog = useAppStore((s) => s.openAddNodeDialog);
	const collapseNodes = useAppStore((s) => s.collapseNodes);
	const expandNodes = useAppStore((s) => s.expandNodes);
	const deleteSelectedNodes = useAppStore((s) => s.deleteSelectedNodes);
	const bulkScrapeSelected = useAppStore((s) => s.bulkScrapeSelected);
	const previewSelectedResults = useAppStore((s) => s.previewSelectedResults);
	const setNodeCrawlExclude = useAppStore((s) => s.setNodeCrawlExclude);

	const [contextMenu, setContextMenu] = useState<{
		x: number;
		y: number;
		kind: 'node' | 'edge' | 'pane';
		id?: string;
	} | null>(null);

	const diff: WorkspaceDiff | undefined = ws
		? workspaceDiffCache[ws.id]
		: undefined;

	const hiddenIds = useMemo(() => {
		if (!ws) return new Set<string>();
		return getHiddenDescendantIds(ws.collapsedNodeIds ?? [], ws.edges);
	}, [ws]);

	const flowNodes: Node<UrlNodeData>[] = useMemo(() => {
		if (!ws) return [];
		const direction = ws.graphLayoutDirection ?? 'LR';
		const handles = handlePositionsForDirection(direction);
		return ws.nodes
			.filter((n) => !hiddenIds.has(n.id))
			.map((n) => {
				const nodeDiff = diff?.nodes.find((d) => d.nodeId === n.id);
				const grayed = isExcludedSubtree(n.id, ws.nodes, ws.edges);
				const collapsedRoots = ws.collapsedNodeIds ?? [];
				return {
					id: n.id,
					type: 'urlNode',
					position: n.position,
					selected: selectedNodeIds.includes(n.id),
					sourcePosition: handles.source,
					targetPosition: handles.target,
					hidden: false,
					selectable: true,
					data: {
						label: n.label,
						status: n.status,
						selected: selectedNodeIds.includes(n.id),
						detailExpanded: (ws.expandedDetailNodeIds ?? []).includes(n.id),
						subtreeCollapsed: collapsedRoots.includes(n.id),
						hasChildren: hasChildNodes(n.id, ws.edges),
						layoutDirection: direction,
						grayed,
						diffKinds: nodeDiff?.kinds,
						url: n.urlNormalized,
					},
				};
			});
	}, [ws, selectedNodeIds, hiddenIds, diff]);

	const flowEdges: Edge[] = useMemo(() => {
		if (!ws) return [];
		return ws.edges
			.filter((e) => !hiddenIds.has(e.source) && !hiddenIds.has(e.target))
			.map((e) => ({
				id: e.id,
				source: e.source,
				target: e.target,
				animated: ws.nodes.find((n) => n.id === e.target)?.status === 'running',
			}));
	}, [ws, hiddenIds]);

	const [nodes, setNodes] = useNodesState(flowNodes);
	const [edges, setEdges, onEdgesChange] = useEdgesState(flowEdges);

	useEffect(() => {
		setNodes(flowNodes);
	}, [flowNodes, setNodes]);

	useEffect(() => {
		setEdges(flowEdges);
	}, [flowEdges, setEdges]);

	const handleNodesChange: OnNodesChange<Node<UrlNodeData>> = useCallback(
		(changes) => {
			setNodes((nds) => applyNodeChanges(changes, nds));
			for (const ch of changes) {
				if (ch.type === 'position' && ch.position && !ch.dragging) {
					updateNodePosition(ch.id, ch.position);
				}
			}
		},
		[setNodes, updateNodePosition],
	);

	const onConnect = useCallback(
		(conn: Connection) => {
			if (!conn.source || !conn.target) return;
			if (addEdgeToStore(conn.source, conn.target)) {
				setEdges((eds) =>
					addEdge(
						{
							...conn,
							id: `e-${conn.source}-${conn.target}`,
						} as Edge,
						eds,
					),
				);
			}
		},
		[addEdgeToStore, setEdges],
	);

	const onEdgesDelete: OnEdgesDelete = useCallback(
		(deleted) => {
			removeEdges(deleted.map((e) => e.id));
		},
		[removeEdges],
	);

	const onNodeClick = useCallback(
		(e: React.MouseEvent, node: Node) => {
			e.stopPropagation();
			useAppStore.setState({ _suppressSelectionSync: true });
			selectNode(node.id, {
				additive: !e.shiftKey && (e.ctrlKey || e.metaKey),
				range: e.shiftKey,
			});
			queueMicrotask(() => {
				useAppStore.setState({ _suppressSelectionSync: false });
			});
		},
		[selectNode],
	);

	const onPaneContextMenu = useCallback((e: MouseEvent | React.MouseEvent) => {
		e.preventDefault();
		const clientX = 'clientX' in e ? e.clientX : 0;
		const clientY = 'clientY' in e ? e.clientY : 0;
		setContextMenu({ x: clientX, y: clientY, kind: 'pane' });
	}, []);

	const onNodeContextMenu = useCallback(
		(e: React.MouseEvent, node: Node) => {
			e.preventDefault();
			selectNode(node.id);
			setContextMenu({
				x: e.clientX,
				y: e.clientY,
				kind: 'node',
				id: node.id,
			});
		},
		[selectNode],
	);

	const onEdgeContextMenu = useCallback((e: React.MouseEvent, edge: Edge) => {
		e.preventDefault();
		setContextMenu({
			x: e.clientX,
			y: e.clientY,
			kind: 'edge',
			id: edge.id,
		});
	}, []);

	if (!ws) {
		return (
			<div className='flex flex-1 items-center justify-center text-sm text-muted-foreground'>
				ワークスペースを作成してください
			</div>
		);
	}

	return (
		<div className='relative flex-1 bg-background'>
			<ReactFlow
				nodes={nodes}
				edges={edges}
				onNodesChange={handleNodesChange}
				onEdgesChange={onEdgesChange}
				onEdgesDelete={onEdgesDelete}
				onConnect={onConnect}
				nodeTypes={nodeTypes}
				onNodeClick={onNodeClick}
				onNodeContextMenu={onNodeContextMenu}
				onEdgeContextMenu={onEdgeContextMenu}
				onPaneContextMenu={onPaneContextMenu}
				onPaneClick={() => {
					setContextMenu(null);
					clearNodeSelection();
				}}
				panOnDrag={!isSelectTool}
				selectionOnDrag={isSelectTool}
				selectionMode={SelectionMode.Partial}
				selectionKeyCode={null}
				multiSelectionKeyCode={['Control', 'Meta']}
				panOnScroll={false}
				zoomOnScroll
				panActivationKeyCode={['Shift', 'Alt', 'Meta']}
				zoomActivationKeyCode={null}
				minZoom={GRAPH_MIN_ZOOM}
				fitView
				fitViewOptions={graphFitViewOptions}
				className={
					isSelectTool
						? 'bg-background rf-tool-select'
						: 'bg-background rf-tool-pan'
				}
				proOptions={proOptions}
			>
				<Background gap={16} />
				<GraphSelectionSync />
				<GraphCanvasControls />
			</ReactFlow>
			{selectedDomain && (
				<div className='pointer-events-none absolute bottom-2 left-2 rounded bg-card/90 px-2 py-1 text-xs text-muted-foreground'>
					ドメイン: {selectedDomain}
				</div>
			)}
			{contextMenu && (
				<div
					className='fixed z-50 min-w-40 rounded-md border border-border bg-popover p-1 shadow-md'
					style={{ left: contextMenu.x, top: contextMenu.y }}
				>
					{contextMenu.kind === 'pane' && (
						<button
							type='button'
							className='block w-full px-2 py-1 text-left text-xs hover:bg-muted'
							onClick={() => {
								openAddNodeDialog({
									x: contextMenu.x,
									y: contextMenu.y,
								});
								setContextMenu(null);
							}}
						>
							{messages.sidebar.newNode}
						</button>
					)}
					{contextMenu.kind === 'node' && contextMenu.id && (
						<>
							<button
								type='button'
								className='block w-full px-2 py-1 text-left text-xs hover:bg-muted'
								onClick={() => {
									collapseNodes([contextMenu.id!]);
									setContextMenu(null);
								}}
							>
								{messages.graph.contextCollapse}
							</button>
							<button
								type='button'
								className='block w-full px-2 py-1 text-left text-xs hover:bg-muted'
								onClick={() => {
									expandNodes([contextMenu.id!]);
									setContextMenu(null);
								}}
							>
								{messages.graph.contextExpand}
							</button>
							<button
								type='button'
								className='block w-full px-2 py-1 text-left text-xs hover:bg-muted'
								onClick={() => {
									setNodeCrawlExclude(contextMenu.id!, true);
									setContextMenu(null);
								}}
							>
								{messages.graph.contextExcludeCrawl}
							</button>
							<button
								type='button'
								className='block w-full px-2 py-1 text-left text-xs hover:bg-muted'
								onClick={() => {
									void bulkScrapeSelected();
									setContextMenu(null);
								}}
							>
								{messages.graph.contextScrape}
							</button>
							<button
								type='button'
								className='block w-full px-2 py-1 text-left text-xs hover:bg-muted'
								onClick={() => {
									void previewSelectedResults();
									setContextMenu(null);
								}}
							>
								{messages.graph.contextPreviewResult}
							</button>
							<button
								type='button'
								className='block w-full px-2 py-1 text-left text-xs text-destructive hover:bg-muted'
								onClick={() => {
									deleteSelectedNodes();
									setContextMenu(null);
								}}
							>
								{messages.graph.contextDelete}
							</button>
						</>
					)}
					{contextMenu.kind === 'edge' && contextMenu.id && (
						<button
							type='button'
							className='block w-full px-2 py-1 text-left text-xs text-destructive hover:bg-muted'
							onClick={() => {
								removeEdges([contextMenu.id!]);
								setContextMenu(null);
							}}
						>
							エッジ削除
						</button>
					)}
				</div>
			)}
		</div>
	);
}
