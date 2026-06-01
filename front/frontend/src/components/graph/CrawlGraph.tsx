import {
	applyNodeChanges,
	Background,
	type Edge,
	type Node,
	type OnEdgesDelete,
	type OnNodesChange,
	ReactFlow,
	useEdgesState,
	useNodesState,
} from '@xyflow/react';
import { useCallback, useEffect, useMemo } from 'react';
import { handlePositionsForDirection } from '@/lib/dagreLayout';
import '@xyflow/react/dist/style.css';
import { useAppStore } from '@/stores/appStore';
import { GraphCanvasControls } from './GraphCanvasControls';
import { UrlNode, type UrlNodeData } from './UrlNode';

const nodeTypes = { urlNode: UrlNode };

export function CrawlGraph() {
	const proOptions = { hideAttribution: true };
	const ws = useAppStore((s) => s.getActiveWorkspace());
	const selectedNodeId = useAppStore((s) => s.selectedNodeId);
	const selectedDomain = useAppStore((s) => s.selectedDomain);
	const selectNode = useAppStore((s) => s.selectNode);
	const updateNodePosition = useAppStore((s) => s.updateNodePosition);
	const removeEdges = useAppStore((s) => s.removeEdges);
	const openAddNodeDialog = useAppStore((s) => s.openAddNodeDialog);

	const flowNodes: Node<UrlNodeData>[] = useMemo(() => {
		if (!ws) return [];
		const direction = ws.graphLayoutDirection ?? 'LR';
		const handles = handlePositionsForDirection(direction);
		return ws.nodes.map((n) => ({
			id: n.id,
			type: 'urlNode',
			position: n.position,
			sourcePosition: handles.source,
			targetPosition: handles.target,
			data: {
				label: n.label,
				status: n.status,
				selected: n.id === selectedNodeId,
				layoutDirection: direction,
			},
		}));
	}, [ws, selectedNodeId]);

	const flowEdges: Edge[] = useMemo(() => {
		if (!ws) return [];
		return ws.edges.map((e) => ({
			id: e.id,
			source: e.source,
			target: e.target,
			animated: ws.nodes.find((n) => n.id === e.target)?.status === 'running',
		}));
	}, [ws]);

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

	const onEdgesDelete: OnEdgesDelete = useCallback(
		(deleted) => {
			removeEdges(deleted.map((e) => e.id));
		},
		[removeEdges],
	);

	const onPaneContextMenu = useCallback(
		(e: MouseEvent | React.MouseEvent) => {
			e.preventDefault();
			const clientX = 'clientX' in e ? e.clientX : 0;
			const clientY = 'clientY' in e ? e.clientY : 0;
			openAddNodeDialog({ x: clientX, y: clientY });
		},
		[openAddNodeDialog],
	);

	if (!ws) {
		return (
			<div className='flex flex-1 items-center justify-center text-sm text-muted-foreground'>
				ワークスペースを作成してください
			</div>
		);
	}

	return (
		<div className='flex-1 bg-background'>
			<ReactFlow
				nodes={nodes}
				edges={edges}
				onNodesChange={handleNodesChange}
				onEdgesChange={onEdgesChange}
				onEdgesDelete={onEdgesDelete}
				nodeTypes={nodeTypes}
				onNodeClick={(_, node) => selectNode(node.id)}
				onPaneContextMenu={onPaneContextMenu}
				fitView
				className='bg-background'
				proOptions={proOptions}
			>
				<Background gap={16} />
				<GraphCanvasControls />
			</ReactFlow>
			{selectedDomain && (
				<div className='pointer-events-none absolute bottom-2 left-2 rounded bg-card/90 px-2 py-1 text-xs text-muted-foreground'>
					ドメイン: {selectedDomain}
				</div>
			)}
		</div>
	);
}
