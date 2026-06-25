import type { Stat } from 'he-tree-react';
import { sortFlatData, useHeTree } from 'he-tree-react';
import { useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { ScrollArea } from '@/components/ui/scroll-area';
import { messages } from '@/i18n/messages';
import { type ExportFlatNode, toggleExportNodeCheck } from '@/lib/exportTree';
import { cn } from '@/lib/utils';

const FLAT_KEYS = { idKey: 'id' as const, parentIdKey: 'parent_id' as const };

type ExportTreeNodeProps = {
	stat: Stat<ExportFlatNode>;
	checkedIds: string[];
	onToggle: (id: string, checked: boolean) => void;
};

function ExportTreeNode({ stat, checkedIds, onToggle }: ExportTreeNodeProps) {
	const node = stat.node;
	const isChecked = checkedIds.includes(node.id);

	return (
		<div
			className={cn(
				'flex items-start gap-2 rounded px-1 py-0.5 text-xs',
				!isChecked && 'opacity-50 text-muted-foreground',
			)}
		>
			<Checkbox
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
};

export function ExportOrderSidebar({
	flatData,
	onFlatDataChange,
	checkedIds,
	onCheckedIdsChange,
}: ExportOrderSidebarProps) {
	const handleChecked = useCallback(
		(id: string, checked: boolean) => {
			onCheckedIdsChange(
				toggleExportNodeCheck(flatData, checkedIds, id, checked),
			);
		},
		[flatData, checkedIds, onCheckedIdsChange],
	);

	const selectAll = () => {
		onCheckedIdsChange(flatData.map((n) => n.id));
	};

	const deselectAll = () => {
		onCheckedIdsChange([]);
	};

	const renderNode = useCallback(
		(stat: Stat<ExportFlatNode>) => (
			<ExportTreeNode
				stat={stat}
				checkedIds={checkedIds}
				onToggle={handleChecked}
			/>
		),
		[checkedIds, handleChecked],
	);

	const { renderTree } = useHeTree({
		...FLAT_KEYS,
		data: flatData,
		dataType: 'flat',
		checkedIds,
		onChange: (next) => {
			onFlatDataChange(sortFlatData(next, FLAT_KEYS) as ExportFlatNode[]);
		},
		isFunctionReactive: true,
		renderNode,
	});

	if (flatData.length === 0) {
		return (
			<aside className='flex h-full flex-col border-r border-border bg-card'>
				<div className='border-b border-border px-3 py-2 text-xs font-semibold'>
					{messages.export.orderTitle}
				</div>
				<p className='p-3 text-xs text-muted-foreground'>
					{messages.export.noNodesInTree}
				</p>
			</aside>
		);
	}

	return (
		<aside className='flex h-full min-w-0 flex-col border-r border-border bg-card'>
			<div className='border-b border-border px-3 py-2 text-xs font-semibold'>
				{messages.export.orderTitle}
			</div>
			<div className='flex gap-1 border-b border-border p-2'>
				<Button size='xs' variant='outline' onClick={selectAll}>
					{messages.export.selectAll}
				</Button>
				<Button size='xs' variant='outline' onClick={deselectAll}>
					{messages.export.deselectAll}
				</Button>
			</div>
			<ScrollArea className='min-h-0 flex-1 p-2'>{renderTree()}</ScrollArea>
		</aside>
	);
}
