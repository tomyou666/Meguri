import { ControlButton, Controls, MiniMap, useReactFlow } from '@xyflow/react';
import {
	ArrowDownUp,
	ArrowLeftRight,
	Hand,
	ListCollapse,
	ListTree,
	Maximize2,
	Minus,
	Plus,
	SquareDashedMousePointer,
} from 'lucide-react';
import { useCallback } from 'react';
import { messages } from '@/i18n/messages';
import type { DagreLayoutDirection } from '@/lib/dagreLayout';
import { cn } from '@/lib/utils';
import { useAppStore } from '@/stores/appStore';
import type { UrlNodeData } from './UrlNode';

export function GraphCanvasControls() {
	const ws = useAppStore((s) => s.getActiveWorkspace());
	const layoutWorkspaceGraph = useAppStore((s) => s.layoutWorkspaceGraph);
	const setGraphLayoutDirection = useAppStore((s) => s.setGraphLayoutDirection);
	const expandAllNodes = useAppStore((s) => s.expandAllNodes);
	const collapseAllNodes = useAppStore((s) => s.collapseAllNodes);
	const graphToolMode = useAppStore((s) => s.graphToolMode);
	const setGraphToolMode = useAppStore((s) => s.setGraphToolMode);
	const { fitView, zoomIn, zoomOut } = useReactFlow();

	const direction = ws?.graphLayoutDirection ?? 'LR';

	const fitAfterLayout = useCallback(() => {
		requestAnimationFrame(() => {
			fitView({ padding: 0.2, duration: 100 });
		});
	}, [fitView]);

	const onFitView = useCallback(() => {
		fitView({ padding: 0.2, duration: 100 });
	}, [fitView]);

	const onCycleLayout = useCallback(() => {
		const next: DagreLayoutDirection = direction === 'LR' ? 'TB' : 'LR';
		setGraphLayoutDirection(next);
		layoutWorkspaceGraph();
		fitAfterLayout();
	}, [
		direction,
		setGraphLayoutDirection,
		layoutWorkspaceGraph,
		fitAfterLayout,
	]);

	const layoutTitle =
		direction === 'LR'
			? `${messages.graph.layoutHorizontal} — クリックで縦方向へ切替・自動配置`
			: `${messages.graph.layoutVertical} — クリックで横方向へ切替・自動配置`;

	return (
		<>
			<MiniMap
				position='bottom-right'
				className='graph-minimap'
				nodeColor={(n) => {
					const st = (n.data as UrlNodeData).status;
					if (st === 'error') return '#ef4444';
					if (st === 'success') return '#22c55e';
					if (st === 'running') return '#3b82f6';
					return '#6b7280';
				}}
				maskColor='color-mix(in oklch, var(--background) 55%, transparent)'
			/>
			<Controls
				position='bottom-left'
				orientation='horizontal'
				showZoom={false}
				showFitView={false}
				showInteractive={false}
				className='graph-controls'
			>
				<ControlButton
					onClick={() => setGraphToolMode('pan')}
					title={messages.graph.toolPan}
					aria-label={messages.graph.toolPan}
					aria-pressed={graphToolMode === 'pan'}
					className={cn(graphToolMode === 'pan' && 'graph-controls-active')}
				>
					<Hand className='size-4' strokeWidth={2} />
				</ControlButton>
				<ControlButton
					onClick={() => setGraphToolMode('select')}
					title={messages.graph.toolSelect}
					aria-label={messages.graph.toolSelect}
					aria-pressed={graphToolMode === 'select'}
					className={cn(graphToolMode === 'select' && 'graph-controls-active')}
				>
					<SquareDashedMousePointer className='size-4' strokeWidth={2} />
				</ControlButton>
				<ControlButton
					onClick={() => zoomIn({ duration: 150 })}
					title={messages.graph.zoomIn}
					aria-label={messages.graph.zoomIn}
				>
					<Plus className='size-4' strokeWidth={2} />
				</ControlButton>
				<ControlButton
					onClick={() => zoomOut({ duration: 150 })}
					title={messages.graph.zoomOut}
					aria-label={messages.graph.zoomOut}
				>
					<Minus className='size-4' strokeWidth={2} />
				</ControlButton>
				<ControlButton
					onClick={onFitView}
					title={messages.graph.fitView}
					aria-label={messages.graph.fitView}
				>
					<Maximize2 className='size-4' strokeWidth={2} />
				</ControlButton>
				<ControlButton
					onClick={onCycleLayout}
					title={layoutTitle}
					aria-label={layoutTitle}
				>
					{direction === 'LR' ? (
						<ArrowLeftRight className='size-4' strokeWidth={2} />
					) : (
						<ArrowDownUp className='size-4' strokeWidth={2} />
					)}
				</ControlButton>
				<ControlButton
					onClick={expandAllNodes}
					title={messages.graph.expandAll}
					aria-label={messages.graph.expandAll}
				>
					<ListTree className='size-4' strokeWidth={2} />
				</ControlButton>
				<ControlButton
					onClick={collapseAllNodes}
					title={messages.graph.collapseAll}
					aria-label={messages.graph.collapseAll}
				>
					<ListCollapse className='size-4' strokeWidth={2} />
				</ControlButton>
			</Controls>
		</>
	);
}
