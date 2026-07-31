import type { Stat } from 'he-tree-react';
import { sortFlatData, useHeTree } from 'he-tree-react';
import { X } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ActionTooltip } from '@/components/ui/action-tooltip';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ScrollArea } from '@/components/ui/scroll-area';
import { messages } from '@/i18n/messages';
import {
	computeSemiCheckedIds,
	type ExportFlatNode,
	toggleExportNodeCheck,
} from '@/lib/exportTree';
import { filterNodeTree } from '@/lib/nodeTree';
import { cn } from '@/lib/utils';

const FLAT_KEYS = { idKey: 'id' as const, parentIdKey: 'parent_id' as const };

type ExportTreeNodeProps = {
	stat: Stat<ExportFlatNode>;
	checkedIds: string[];
	semiCheckedIds: string[];
	isHit: boolean;
	onToggle: (id: string, checked: boolean) => void;
};

function ExportTreeNode({
	stat,
	checkedIds,
	semiCheckedIds,
	isHit,
	onToggle,
}: ExportTreeNodeProps) {
	const node = stat.node;
	const isChecked = checkedIds.includes(node.id);
	const isSemiChecked = !isChecked && semiCheckedIds.includes(node.id);
	const inputRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		if (inputRef.current) {
			inputRef.current.indeterminate = isSemiChecked;
		}
	}, [isSemiChecked]);

	return (
		<div
			className={cn(
				'flex items-start gap-2 rounded px-1 py-0.5 text-xs',
				isHit && 'bg-muted',
				!isChecked && !isSemiChecked && 'text-muted-foreground opacity-50',
			)}
		>
			<Checkbox
				ref={inputRef}
				checked={isChecked}
				onCheckedChange={() => onToggle(node.id, !isChecked)}
				onClick={(e) => e.stopPropagation()}
				aria-label={node.label}
			/>
			<div className='min-w-0 flex-1'>
				<p className='truncate font-medium'>{node.label}</p>
				<p className='truncate text-[10px] text-muted-foreground'>
					{node.urlNormalized}
				</p>
			</div>
		</div>
	);
}

type ExportOrderSidebarProps = {
	flatData: ExportFlatNode[];
	onFlatDataChange: (data: ExportFlatNode[]) => void;
	checkedIds: string[];
	onCheckedIdsChange: (ids: string[]) => void;
	cascadeCheck: boolean;
	onCascadeCheckChange: (value: boolean) => void;
};

export function ExportOrderSidebar({
	flatData,
	onFlatDataChange,
	checkedIds,
	onCheckedIdsChange,
	cascadeCheck,
	onCascadeCheckChange,
}: ExportOrderSidebarProps) {
	const [query, setQuery] = useState('');

	const filterResult = useMemo(
		() => filterNodeTree(flatData, query, []),
		[flatData, query],
	);

	const isFiltering = query.trim().length > 0;

	const displayData = useMemo(() => {
		if (!isFiltering) return flatData;
		return flatData.filter((n) => filterResult.visibleIds.has(n.id));
	}, [flatData, isFiltering, filterResult.visibleIds]);

	const semiCheckedIds = useMemo(
		() => computeSemiCheckedIds(flatData, checkedIds),
		[flatData, checkedIds],
	);

	const handleChecked = useCallback(
		(id: string, checked: boolean) => {
			onCheckedIdsChange(
				toggleExportNodeCheck(flatData, checkedIds, id, checked, cascadeCheck),
			);
		},
		[flatData, checkedIds, onCheckedIdsChange, cascadeCheck],
	);

	const selectAll = () => {
		if (isFiltering) {
			const next = new Set(checkedIds);
			for (const id of filterResult.hitIds) next.add(id);
			onCheckedIdsChange([...next]);
			return;
		}
		onCheckedIdsChange(flatData.map((n) => n.id));
	};

	const deselectAll = () => {
		if (isFiltering) {
			onCheckedIdsChange(
				checkedIds.filter((id) => !filterResult.hitIds.has(id)),
			);
			return;
		}
		onCheckedIdsChange([]);
	};

	const renderNode = useCallback(
		(stat: Stat<ExportFlatNode>) => (
			<ExportTreeNode
				stat={stat}
				checkedIds={checkedIds}
				semiCheckedIds={semiCheckedIds}
				isHit={filterResult.hitIds.has(stat.node.id)}
				onToggle={handleChecked}
			/>
		),
		[checkedIds, semiCheckedIds, filterResult.hitIds, handleChecked],
	);

	const { renderTree } = useHeTree({
		...FLAT_KEYS,
		data: displayData,
		dataType: 'flat',
		checkedIds,
		onChange: (next) => {
			if (isFiltering) return;
			onFlatDataChange(sortFlatData(next, FLAT_KEYS) as ExportFlatNode[]);
		},
		canDrag: () => !isFiltering,
		isFunctionReactive: true,
		renderNode,
	});

	if (flatData.length === 0) {
		return (
			<aside className='flex h-full flex-col border-border border-r bg-card'>
				<div className='border-border border-b px-3 py-2 font-semibold text-xs'>
					{messages.export.orderTitle}
				</div>
				<p className='p-3 text-muted-foreground text-xs'>
					{messages.export.noNodesInTree}
				</p>
			</aside>
		);
	}

	return (
		<aside className='flex h-full min-w-0 flex-col border-border border-r bg-card'>
			<div className='border-border border-b px-3 py-2 font-semibold text-xs'>
				{messages.export.orderTitle}
			</div>
			<div className='space-y-1.5 border-border border-b p-2'>
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
					<Button size='xs' variant='outline' onClick={selectAll}>
						{messages.export.selectAll}
					</Button>
					<Button size='xs' variant='outline' onClick={deselectAll}>
						{messages.export.deselectAll}
					</Button>
				</div>
				<div className='flex w-full items-center gap-2'>
					<Checkbox
						id='export-cascade-check'
						checked={cascadeCheck}
						onCheckedChange={(checked) =>
							onCascadeCheckChange(checked === true)
						}
					/>
					<Label
						htmlFor='export-cascade-check'
						className='font-normal text-[10px]'
					>
						{messages.export.cascadeCheck}
					</Label>
				</div>
			</div>
			<ScrollArea className='min-h-0 flex-1 p-2'>
				{isFiltering && filterResult.hitCount === 0 ? (
					<p className='px-1 py-2 text-muted-foreground text-xs'>
						{messages.nodeTree.noMatches}
					</p>
				) : (
					renderTree()
				)}
			</ScrollArea>
		</aside>
	);
}
