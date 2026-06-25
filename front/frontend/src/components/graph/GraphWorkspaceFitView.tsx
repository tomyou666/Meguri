import { useNodes, useReactFlow } from '@xyflow/react';
import { useEffect, useRef } from 'react';
import { useAppStore } from '@/stores/appStore';
import { GRAPH_MIN_ZOOM } from './GraphCanvasControls';

/** ReactFlow 内部でのみマウントすること（Provider 必須） */
export function GraphWorkspaceFitView() {
	const activeWorkspaceId = useAppStore((s) => s.activeWorkspaceId);
	const ws = useAppStore((s) => s.getActiveWorkspace());
	const flowNodes = useNodes();
	const { fitView } = useReactFlow();
	const prevActiveWorkspaceId = useRef<string | null>(null);

	const wsNodeIds =
		ws?.nodes
			.map((n) => n.id)
			.sort()
			.join('\0') ?? '';
	const flowNodeIds = flowNodes
		.map((n) => n.id)
		.sort()
		.join('\0');

	useEffect(() => {
		if (!activeWorkspaceId || wsNodeIds !== flowNodeIds) return;

		if (prevActiveWorkspaceId.current === null) {
			prevActiveWorkspaceId.current = activeWorkspaceId;
			return;
		}

		if (prevActiveWorkspaceId.current === activeWorkspaceId) return;

		prevActiveWorkspaceId.current = activeWorkspaceId;
		const frame = requestAnimationFrame(() => {
			fitView({ padding: 0.2, duration: 100, minZoom: GRAPH_MIN_ZOOM });
		});
		return () => cancelAnimationFrame(frame);
	}, [activeWorkspaceId, wsNodeIds, flowNodeIds, fitView]);

	return null;
}
