import { ControlButton, Controls, MiniMap, useReactFlow } from '@xyflow/react';
import {
	ArrowDownUp,
	ArrowLeftRight,
	LayoutGrid,
	Maximize2,
	Minus,
	Plus,
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
	const { fitView, zoomIn, zoomOut } = useReactFlow();

	const direction = ws?.graphLayoutDirection ?? 'LR';

	const fitAfterLayout = useCallback(() => {
		requestAnimationFrame(() => {
			fitView({ padding: 0.2, duration: 100 });
		});
	}, [fitView]);

	const onLayout = useCallback(() => {
		layoutWorkspaceGraph();
		fitAfterLayout();
	}, [layoutWorkspaceGraph, fitAfterLayout]);

	const onFitView = useCallback(() => {
		fitView({ padding: 0.2, duration: 100 });
	}, [fitView]);

	const onDirection = useCallback(
		(next: DagreLayoutDirection) => {
			if (direction === next) {
				layoutWorkspaceGraph();
			} else {
				setGraphLayoutDirection(next);
			}
			fitAfterLayout();
		},
		[direction, setGraphLayoutDirection, layoutWorkspaceGraph, fitAfterLayout],
	);

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
					onClick={() => onDirection('TB')}
					title={messages.graph.layoutVertical}
					aria-label={messages.graph.layoutVertical}
					aria-pressed={direction === 'TB'}
					className={cn(direction === 'TB' && 'graph-controls-active')}
				>
					<ArrowDownUp className='size-4' strokeWidth={2} />
				</ControlButton>
				<ControlButton
					onClick={() => onDirection('LR')}
					title={messages.graph.layoutHorizontal}
					aria-label={messages.graph.layoutHorizontal}
					aria-pressed={direction === 'LR'}
					className={cn(direction === 'LR' && 'graph-controls-active')}
				>
					<ArrowLeftRight className='size-4' strokeWidth={2} />
				</ControlButton>
				<ControlButton
					onClick={onLayout}
					title={messages.graph.layout}
					aria-label={messages.graph.layout}
				>
					<LayoutGrid className='size-4' strokeWidth={2} />
				</ControlButton>
			</Controls>
		</>
	);
}
