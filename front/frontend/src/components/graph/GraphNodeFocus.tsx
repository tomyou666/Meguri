import { useReactFlow } from '@xyflow/react';
import { useEffect } from 'react';
import { useAppStore } from '@/stores/appStore';
import { GRAPH_MIN_ZOOM } from './GraphCanvasControls';

/** ReactFlow 内部でのみマウントすること（Provider 必須） */
export function GraphNodeFocus() {
	const graphFocusRequest = useAppStore((s) => s.graphFocusRequest);
	const clearGraphFocusRequest = useAppStore((s) => s.clearGraphFocusRequest);
	const { fitView } = useReactFlow();

	useEffect(() => {
		if (!graphFocusRequest || graphFocusRequest.ids.length === 0) return;
		const nodes = graphFocusRequest.ids.map((id) => ({ id }));
		const frame = requestAnimationFrame(() => {
			fitView({
				nodes,
				padding: 0.35,
				duration: 200,
				minZoom: GRAPH_MIN_ZOOM,
				maxZoom: 1.5,
			});
			clearGraphFocusRequest();
		});
		return () => cancelAnimationFrame(frame);
	}, [graphFocusRequest, fitView, clearGraphFocusRequest]);

	return null;
}
