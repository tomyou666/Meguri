import {
	ChevronDown,
	ChevronRight,
	Copy,
	FoldVertical,
	Menu,
	UnfoldVertical,
	X,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { NodeContextMenuItems } from '@/components/graph/NodeContextMenuItems';
import { ActionTooltip } from '@/components/ui/action-tooltip';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from '@/components/ui/tooltip';
import { messages } from '@/i18n/messages';
import { hasChildNodes, isExcludedSubtree } from '@/lib/graph';
import { nodeStatusUi } from '@/lib/nodeStatusUi';
import {
	buildNodeFlatTree,
	expandableNodeIds,
	expandIdsForHits,
	filterNodeTree,
	groupChildrenByParent,
	type NodeFlatNode,
} from '@/lib/nodeTree';
import { cn } from '@/lib/utils';
import { useAppStore } from '@/stores/appStore';
import type { GraphEdge, GraphNode, NodeStatus } from '@/types/graph';

const ALL_STATUSES: NodeStatus[] = [
	'idle',
	'running',
	'success',
	'error',
	'skipped',
];

type NodeTreePanelProps = {
	nodes: GraphNode[];
	edges: GraphEdge[];
	seedUrl?: string;
};

type TreeRowProps = {
	node: NodeFlatNode;
	depth: number;
	hasVisibleChildren: boolean;
	hasGraphChildren: boolean;
	expanded: boolean;
	selected: boolean;
	isHit: boolean;
	grayed: boolean;
	crawlExclude: boolean;
	isCollapsedOnGraph: boolean;
	onToggleExpand: (id: string) => void;
	onSelect: (id: string, e: React.MouseEvent) => void;
	onCopy: (url: string) => void;
	onMenuOpen: (id: string) => void;
};

function TreeRow({
	node,
	depth,
	hasVisibleChildren,
	hasGraphChildren,
	expanded,
	selected,
	isHit,
	grayed,
	crawlExclude,
	isCollapsedOnGraph,
	onToggleExpand,
	onSelect,
	onCopy,
	onMenuOpen,
}: TreeRowProps) {
	const cfg = nodeStatusUi[node.status] ?? nodeStatusUi.idle;
	const [menuOpen, setMenuOpen] = useState(false);

	return (
		<div
			className={cn(
				'group flex w-full items-start gap-0.5 rounded-md px-0.5 py-0.5',
				selected && 'bg-sidebar-accent',
				isHit && !selected && 'bg-sidebar-accent/50',
				grayed && 'opacity-45 grayscale',
			)}
			data-tree-node-id={node.id}
			style={{ paddingLeft: `${depth * 0.75 + 0.125}rem` }}
		>
			{hasVisibleChildren ? (
				<button
					type='button'
					className='mt-0.5 flex size-4 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted'
					aria-label={
						expanded
							? messages.graph.collapseSubtree
							: messages.graph.expandSubtree
					}
					onClick={(e) => {
						e.stopPropagation();
						onToggleExpand(node.id);
					}}
				>
					{expanded ? (
						<ChevronDown className='size-3' />
					) : (
						<ChevronRight className='size-3' />
					)}
				</button>
			) : (
				<span className='size-4 shrink-0' />
			)}
			<div className='min-w-0 flex-1 overflow-hidden'>
				<Tooltip>
					<TooltipTrigger asChild>
						<button
							type='button'
							className='flex w-full min-w-0 items-start gap-1 rounded-md px-1 py-0.5 text-left hover:bg-sidebar-accent/80'
							onClick={(e) => onSelect(node.id, e)}
						>
							<span className='mt-0.5'>{cfg.icon}</span>
							<span className='min-w-0 flex-1 overflow-hidden'>
								<span
									className={cn(
										'block truncate font-medium text-xs',
										cfg.textClass,
									)}
								>
									{node.label}
								</span>
								<span className='mt-0.5 block truncate text-[10px] text-muted-foreground'>
									{node.urlNormalized}
								</span>
							</span>
						</button>
					</TooltipTrigger>
					<TooltipContent side='top' className='max-w-sm break-all text-left'>
						{node.urlNormalized}
					</TooltipContent>
				</Tooltip>
			</div>
			<ActionTooltip label={messages.nodeTree.copyUrl}>
				<Button
					variant='ghost'
					size='icon-xs'
					className={cn(
						'mt-0.5 shrink-0 bg-sidebar opacity-0 hover:bg-muted focus-visible:opacity-100 group-hover:opacity-100',
						(selected || isHit) && 'bg-sidebar-accent',
					)}
					aria-label={messages.nodeTree.copyUrl}
					onClick={(e) => {
						e.stopPropagation();
						onCopy(node.urlNormalized);
					}}
				>
					<Copy className='size-3' />
				</Button>
			</ActionTooltip>
			<DropdownMenu
				open={menuOpen}
				onOpenChange={(open) => {
					if (open) onMenuOpen(node.id);
					setMenuOpen(open);
				}}
			>
				<DropdownMenuTrigger asChild>
					<Button
						variant='ghost'
						size='icon-xs'
						className={cn(
							'mt-0.5 shrink-0 bg-sidebar hover:bg-muted',
							(selected || isHit) && 'bg-sidebar-accent',
						)}
						aria-label={messages.nodeTree.openMenu}
						onClick={(e) => e.stopPropagation()}
					>
						<Menu className='size-3' />
					</Button>
				</DropdownMenuTrigger>
				<DropdownMenuContent
					align='end'
					sideOffset={6}
					className='min-w-40 border-border p-1 shadow-lg'
				>
					<NodeContextMenuItems
						state={{
							nodeId: node.id,
							hasChildren: hasGraphChildren,
							isCollapsed: isCollapsedOnGraph,
							crawlExclude,
						}}
						onClose={() => setMenuOpen(false)}
					/>
				</DropdownMenuContent>
			</DropdownMenu>
		</div>
	);
}

export function NodeTreePanel({ nodes, edges, seedUrl }: NodeTreePanelProps) {
	const selectedNodeIds = useAppStore((s) => s.selectedNodeIds);
	const selectNode = useAppStore((s) => s.selectNode);
	const requestGraphFocus = useAppStore((s) => s.requestGraphFocus);
	const treeFocusRequest = useAppStore((s) => s.treeFocusRequest);
	const clearTreeFocusRequest = useAppStore((s) => s.clearTreeFocusRequest);
	const collapsedNodeIds = useAppStore(
		(s) => s.getActiveWorkspace()?.collapsedNodeIds ?? [],
	);

	const flat = useMemo(
		() => buildNodeFlatTree(nodes, edges, seedUrl),
		[nodes, edges, seedUrl],
	);
	const byParent = useMemo(() => groupChildrenByParent(flat), [flat]);
	const expandableIds = useMemo(() => expandableNodeIds(flat), [flat]);

	const [query, setQuery] = useState('');
	const [statusFilter, setStatusFilter] = useState<NodeStatus[]>([]);
	const [expandedIds, setExpandedIds] = useState<Set<string>>(
		() => new Set(expandableIds),
	);
	const expandSnapshotRef = useRef<Set<string> | null>(null);
	const wasFilteringRef = useRef(false);
	const knownExpandableRef = useRef<Set<string>>(new Set(expandableIds));
	const listRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const expandable = expandableNodeIds(flat);
		const expandableSet = new Set(expandable);
		setExpandedIds((prev) => {
			const next = new Set<string>();
			for (const id of prev) {
				if (expandableSet.has(id)) next.add(id);
			}
			for (const id of expandable) {
				if (!knownExpandableRef.current.has(id)) next.add(id);
			}
			return next;
		});
		knownExpandableRef.current = expandableSet;
	}, [flat]);

	const filterResult = useMemo(
		() => filterNodeTree(flat, query, statusFilter),
		[flat, query, statusFilter],
	);

	const isFiltering = query.trim().length > 0 || statusFilter.length > 0;

	useEffect(() => {
		if (isFiltering) {
			setExpandedIds((prev) => {
				if (!wasFilteringRef.current) {
					expandSnapshotRef.current = new Set(prev);
					wasFilteringRef.current = true;
				}
				return expandIdsForHits(flat, filterResult.hitIds);
			});
			return;
		}
		if (wasFilteringRef.current) {
			wasFilteringRef.current = false;
			if (expandSnapshotRef.current) {
				setExpandedIds(expandSnapshotRef.current);
				expandSnapshotRef.current = null;
			}
		}
	}, [isFiltering, flat, filterResult.hitIds]);

	useEffect(() => {
		if (!treeFocusRequest || treeFocusRequest.ids.length === 0) return;
		const id = treeFocusRequest.ids[0];
		if (isFiltering && !filterResult.visibleIds.has(id)) {
			clearTreeFocusRequest();
			return;
		}
		const ancestors = expandIdsForHits(flat, new Set([id]));
		setExpandedIds((prev) => {
			let changed = false;
			const next = new Set(prev);
			for (const ancestorId of ancestors) {
				if (!next.has(ancestorId)) {
					next.add(ancestorId);
					changed = true;
				}
			}
			return changed ? next : prev;
		});

		let innerFrame = 0;
		const outerFrame = requestAnimationFrame(() => {
			innerFrame = requestAnimationFrame(() => {
				const el = listRef.current?.querySelector(
					`[data-tree-node-id="${CSS.escape(id)}"]`,
				);
				el?.scrollIntoView({ block: 'center' });
				clearTreeFocusRequest();
			});
		});
		return () => {
			cancelAnimationFrame(outerFrame);
			cancelAnimationFrame(innerFrame);
		};
	}, [
		treeFocusRequest,
		isFiltering,
		filterResult.visibleIds,
		flat,
		clearTreeFocusRequest,
	]);

	const toggleExpand = useCallback((id: string) => {
		setExpandedIds((prev) => {
			const next = new Set(prev);
			if (next.has(id)) next.delete(id);
			else next.add(id);
			return next;
		});
	}, []);

	const expandAll = () => setExpandedIds(new Set(expandableIds));
	const collapseAll = () => setExpandedIds(new Set());

	const toggleStatus = (status: NodeStatus) => {
		setStatusFilter((prev) =>
			prev.includes(status)
				? prev.filter((s) => s !== status)
				: [...prev, status],
		);
	};

	const onSelect = (id: string, e: React.MouseEvent) => {
		const additive = !e.shiftKey && (e.ctrlKey || e.metaKey);
		const range = e.shiftKey;
		selectNode(id, { additive, range });
		if (!additive && !range) {
			requestGraphFocus([id]);
		}
	};

	const onCopy = async (url: string) => {
		try {
			await navigator.clipboard.writeText(url);
		} catch {
			/* ignore */
		}
	};

	const onMenuOpen = (id: string) => {
		selectNode(id);
	};

	const renderBranch = (parentId: string | null, depth: number) => {
		const children = (byParent.get(parentId) ?? []).filter((n) =>
			filterResult.visibleIds.has(n.id),
		);
		return children.map((node) => {
			const childList = byParent.get(node.id) ?? [];
			const visibleChildren = childList.some((c) =>
				filterResult.visibleIds.has(c.id),
			);
			const expanded = expandedIds.has(node.id);
			const grayed = isExcludedSubtree(node.id, nodes, edges);
			const graphNode = nodes.find((n) => n.id === node.id);

			return (
				<div key={node.id}>
					<TreeRow
						node={node}
						depth={depth}
						hasVisibleChildren={visibleChildren}
						hasGraphChildren={hasChildNodes(node.id, edges)}
						expanded={expanded}
						selected={selectedNodeIds.includes(node.id)}
						isHit={filterResult.hitIds.has(node.id)}
						grayed={grayed}
						crawlExclude={graphNode?.crawlExclude ?? false}
						isCollapsedOnGraph={collapsedNodeIds.includes(node.id)}
						onToggleExpand={toggleExpand}
						onSelect={onSelect}
						onCopy={onCopy}
						onMenuOpen={onMenuOpen}
					/>
					{visibleChildren && expanded && renderBranch(node.id, depth + 1)}
				</div>
			);
		});
	};

	if (flat.length === 0) {
		return (
			<p className='px-2 py-2 text-muted-foreground text-xs'>
				{messages.sidebar.emptyNodes}
			</p>
		);
	}

	return (
		<div className='flex min-h-0 flex-1 flex-col gap-1.5'>
			<div className='space-y-1.5 px-1'>
				<div className='flex items-center gap-1'>
					<div className='relative min-w-0 flex-1'>
						<Input
							value={query}
							onChange={(e) => setQuery(e.target.value)}
							placeholder={messages.nodeTree.searchPlaceholder}
							className='h-7 pr-7 text-xs'
							aria-label={messages.nodeTree.searchPlaceholder}
						/>
						{query.length > 0 && (
							<ActionTooltip label={messages.nodeTree.clearSearch}>
								<Button
									variant='ghost'
									size='icon-xs'
									className='absolute top-1/2 right-0.5 -translate-y-1/2'
									aria-label={messages.nodeTree.clearSearch}
									onClick={() => setQuery('')}
								>
									<X className='size-3' />
								</Button>
							</ActionTooltip>
						)}
					</div>
					{isFiltering && (
						<span className='shrink-0 text-[10px] text-muted-foreground'>
							{messages.nodeTree.hitCount(filterResult.hitCount)}
						</span>
					)}
				</div>
				<div className='flex flex-wrap gap-1'>
					{ALL_STATUSES.map((status) => {
						const active = statusFilter.includes(status);
						return (
							<button
								key={status}
								type='button'
								onClick={() => toggleStatus(status)}
								aria-pressed={active}
							>
								<Badge
									variant={active ? 'default' : 'outline'}
									className='cursor-pointer text-[10px]'
								>
									{nodeStatusUi[status].label}
								</Badge>
							</button>
						);
					})}
				</div>
				<div className='flex gap-1'>
					<ActionTooltip label={messages.nodeTree.expandAll}>
						<Button
							variant='outline'
							size='xs'
							aria-label={messages.nodeTree.expandAll}
							onClick={expandAll}
						>
							<UnfoldVertical className='size-3' />
						</Button>
					</ActionTooltip>
					<ActionTooltip label={messages.nodeTree.collapseAll}>
						<Button
							variant='outline'
							size='xs'
							aria-label={messages.nodeTree.collapseAll}
							onClick={collapseAll}
						>
							<FoldVertical className='size-3' />
						</Button>
					</ActionTooltip>
				</div>
			</div>
			<div ref={listRef} className='min-h-0 flex-1 overflow-auto px-0.5 pb-1'>
				{isFiltering && filterResult.hitCount === 0 ? (
					<p className='px-2 py-2 text-muted-foreground text-xs'>
						{messages.nodeTree.noMatches}
					</p>
				) : (
					renderBranch(null, 0)
				)}
			</div>
		</div>
	);
}
