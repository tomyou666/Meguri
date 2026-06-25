import { Handle, type NodeProps } from '@xyflow/react';
import {
	AlertCircle,
	CheckCircle2,
	ChevronDown,
	ChevronRight,
	Circle,
	Info,
	Loader2,
	SkipForward,
} from 'lucide-react';
import { memo, type ReactNode } from 'react';
import { NodeDetailSettings } from '@/components/graph/NodeDetailSettings';
import { ActionTooltip } from '@/components/ui/action-tooltip';
import { Badge } from '@/components/ui/badge';
import { ExternalLink } from '@/components/ui/external-link';
import { messages } from '@/i18n/messages';
import {
	DAGRE_NODE_HEIGHT,
	DAGRE_NODE_WIDTH,
	type DagreLayoutDirection,
	handlePositionsForDirection,
	NODE_DETAIL_EXPANDED_WIDTH,
} from '@/lib/dagreLayout';
import { cn } from '@/lib/utils';
import { useAppStore } from '@/stores/appStore';
import type { DiffKind } from '@/types/adapter';
import type { NodeStatus } from '@/types/graph';

export type UrlNodeData = {
	label: string;
	status: NodeStatus;
	selected?: boolean;
	detailExpanded?: boolean;
	subtreeCollapsed?: boolean;
	hasChildren?: boolean;
	layoutDirection?: DagreLayoutDirection;
	grayed?: boolean;
	diffKinds?: DiffKind[];
	url?: string;
};

const statusConfig: Record<
	NodeStatus,
	{ icon: ReactNode; border: string; label: string }
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

function NodeIconButton({
	title,
	onClick,
	children,
}: {
	title: string;
	onClick: () => void;
	children: ReactNode;
}) {
	return (
		<ActionTooltip label={title}>
			<button
				type='button'
				className='nodrag nopan flex size-5 shrink-0 items-center justify-center rounded hover:bg-muted'
				aria-label={title}
				onClick={(e) => {
					e.stopPropagation();
					onClick();
				}}
			>
				{children}
			</button>
		</ActionTooltip>
	);
}

function UrlNodeComponent({ id, data }: NodeProps) {
	const d = data as UrlNodeData;
	const cfg = statusConfig[d.status] ?? statusConfig.idle;
	const handles = handlePositionsForDirection(d.layoutDirection ?? 'LR');
	const toggleDetail = useAppStore((s) => s.toggleNodeDetailExpand);
	const toggleSubtree = useAppStore((s) => s.toggleNodeSubtreeCollapse);
	const node = useAppStore((s) => {
		const ws = s.getActiveWorkspace();
		return ws?.nodes.find((n) => n.id === id) ?? null;
	});

	const detailExpanded = d.detailExpanded ?? false;
	const minHeight = detailExpanded ? DAGRE_NODE_HEIGHT + 72 : DAGRE_NODE_HEIGHT;
	const width = detailExpanded ? NODE_DETAIL_EXPANDED_WIDTH : DAGRE_NODE_WIDTH;

	return (
		<div
			className={cn(
				'box-border shrink-0 rounded-lg border-2 bg-card px-2 py-1.5 shadow-sm',
				detailExpanded ? 'overflow-visible' : 'overflow-hidden',
				cfg.border,
				d.selected && 'ring-2 ring-ring',
				d.grayed && !detailExpanded && 'opacity-45 grayscale',
			)}
			style={{
				width,
				minHeight,
			}}
			onWheelCapture={detailExpanded ? (e) => e.stopPropagation() : undefined}
		>
			<Handle
				type='target'
				position={handles.target}
				className='bg-muted-foreground!'
			/>
			<div className='flex min-h-0 flex-col gap-1'>
				<div className='flex items-start gap-0.5'>
					<NodeIconButton
						title={
							detailExpanded
								? messages.graph.collapseDetail
								: messages.graph.expandDetail
						}
						onClick={() => toggleDetail(id)}
					>
						<Info
							className={cn('size-3.5', detailExpanded && 'text-primary')}
						/>
					</NodeIconButton>
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
					{d.hasChildren && (
						<NodeIconButton
							title={
								d.subtreeCollapsed
									? messages.graph.expandSubtree
									: messages.graph.collapseSubtree
							}
							onClick={() => toggleSubtree(id)}
						>
							{d.subtreeCollapsed ? (
								<ChevronRight className='size-3.5' />
							) : (
								<ChevronDown className='size-3.5' />
							)}
						</NodeIconButton>
					)}
				</div>
				{d.diffKinds && d.diffKinds.length > 0 && (
					<div className='flex flex-wrap gap-0.5'>
						{d.diffKinds.map((k) => (
							<Badge key={k} variant='outline' className='px-1 py-0 text-[8px]'>
								{k}
							</Badge>
						))}
					</div>
				)}
				{detailExpanded && d.url && (
					<ExternalLink
						href={d.url}
						stopPropagation
						className='break-all text-[9px] text-muted-foreground'
					/>
				)}
				{detailExpanded && node && <NodeDetailSettings node={node} />}
			</div>
			<Handle
				type='source'
				position={handles.source}
				className='bg-muted-foreground!'
			/>
		</div>
	);
}

export const UrlNode = memo(UrlNodeComponent);
