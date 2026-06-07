import { MiniMap, Panel } from '@xyflow/react';
import { Map as MapIcon, Minimize2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { messages } from '@/i18n/messages';
import { useAppStore } from '@/stores/appStore';
import type { UrlNodeData } from './UrlNode';

export const GRAPH_MINIMAP_WIDTH = 200;
export const GRAPH_MINIMAP_HEIGHT = 150;

export function GraphMinimap() {
	const minimapCollapsed = useAppStore((s) => s.minimapCollapsed);
	const toggleMinimap = useAppStore((s) => s.toggleMinimap);

	if (minimapCollapsed) {
		return (
			<Panel position='bottom-right' className='graph-minimap-collapsed'>
				<Button
					type='button'
					variant='outline'
					size='icon-sm'
					onClick={toggleMinimap}
					title={messages.graph.minimapOpen}
					aria-label={messages.graph.minimapOpen}
				>
					<MapIcon className='size-4' strokeWidth={2} />
				</Button>
			</Panel>
		);
	}

	return (
		<>
			<Panel
				position='bottom-right'
				className='graph-minimap-toolbar'
				style={{
					marginBottom: GRAPH_MINIMAP_HEIGHT,
					width: GRAPH_MINIMAP_WIDTH,
				}}
			>
				<div className='graph-minimap-toolbar-inner'>
					<span className='graph-minimap-label'>
						{messages.graph.minimapTitle}
					</span>
					<Button
						type='button'
						variant='ghost'
						size='icon-xs'
						onClick={toggleMinimap}
						title={messages.graph.minimapClose}
						aria-label={messages.graph.minimapClose}
					>
						<Minimize2 className='size-3.5' strokeWidth={2} />
					</Button>
				</div>
			</Panel>
			<MiniMap
				pannable
				position='bottom-right'
				className='graph-minimap graph-minimap-expanded'
				style={{ width: GRAPH_MINIMAP_WIDTH, height: GRAPH_MINIMAP_HEIGHT }}
				nodeColor={(n) => {
					const st = (n.data as UrlNodeData).status;
					if (st === 'error') return '#ef4444';
					if (st === 'success') return '#22c55e';
					if (st === 'running') return '#3b82f6';
					return '#6b7280';
				}}
				maskColor='color-mix(in oklch, var(--background) 55%, transparent)'
			/>
		</>
	);
}
