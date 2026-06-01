import { Handle, type NodeProps } from '@xyflow/react';
import {
	AlertCircle,
	CheckCircle2,
	Circle,
	Loader2,
	SkipForward,
} from 'lucide-react';
import { memo } from 'react';
import { messages } from '@/i18n/messages';
import {
	DAGRE_NODE_HEIGHT,
	DAGRE_NODE_WIDTH,
	type DagreLayoutDirection,
	handlePositionsForDirection,
} from '@/lib/dagreLayout';
import { cn } from '@/lib/utils';
import type { NodeStatus } from '@/types/graph';

export type UrlNodeData = {
	label: string;
	status: NodeStatus;
	selected?: boolean;
	layoutDirection?: DagreLayoutDirection;
};

const statusConfig: Record<
	NodeStatus,
	{ icon: React.ReactNode; border: string; label: string }
> = {
	idle: {
		icon: <Circle className='size-3 text-muted-foreground' />,
		border: 'border-border',
		label: messages.status.idle,
	},
	running: {
		icon: <Loader2 className='size-3 animate-spin text-blue-400' />,
		border: 'border-blue-500',
		label: messages.status.running,
	},
	success: {
		icon: <CheckCircle2 className='size-3 text-emerald-400' />,
		border: 'border-emerald-500',
		label: messages.status.success,
	},
	error: {
		icon: <AlertCircle className='size-3 text-destructive' />,
		border: 'border-destructive',
		label: messages.status.error,
	},
	skipped: {
		icon: <SkipForward className='size-3 text-amber-400' />,
		border: 'border-amber-500',
		label: messages.status.skipped,
	},
};

function UrlNodeComponent({ data }: NodeProps) {
	const d = data as UrlNodeData;
	const cfg = statusConfig[d.status] ?? statusConfig.idle;
	const handles = handlePositionsForDirection(d.layoutDirection ?? 'LR');

	return (
		<div
			className={cn(
				'box-border shrink-0 overflow-hidden rounded-lg border-2 bg-card px-2 py-1.5 shadow-sm',
				cfg.border,
				d.selected && 'ring-2 ring-ring',
			)}
			style={{ width: DAGRE_NODE_WIDTH, height: DAGRE_NODE_HEIGHT }}
		>
			<Handle
				type='target'
				position={handles.target}
				className='!bg-muted-foreground'
			/>
			<div className='flex h-full min-h-0 items-start gap-1.5'>
				{cfg.icon}
				<div className='min-w-0 flex-1 overflow-hidden'>
					<p
						className='truncate text-[10px] font-medium leading-tight'
						title={d.label}
					>
						{d.label}
					</p>
					<p className='truncate text-[9px] text-muted-foreground'>
						{cfg.label}
					</p>
				</div>
			</div>
			<Handle
				type='source'
				position={handles.source}
				className='!bg-muted-foreground'
			/>
		</div>
	);
}

export const UrlNode = memo(UrlNodeComponent);
