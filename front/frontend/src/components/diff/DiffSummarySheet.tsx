import { CheckCheck } from 'lucide-react';
import { useMemo, useState } from 'react';
import {
	filterNodesByKind,
	summaryBadgeLabels,
} from '@/components/diff/diffSummaryUtils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
	Sheet,
	SheetContent,
	SheetHeader,
	SheetTitle,
} from '@/components/ui/sheet';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { messages } from '@/i18n/messages';
import { cn } from '@/lib/utils';
import { useAppStore } from '@/stores/appStore';
import type { DiffKind } from '@/types/adapter';

type FilterKind = DiffKind | 'all';

export function DiffSummarySheet() {
	const open = useAppStore((s) => s.diffSummaryOpen);
	const workspaceId = useAppStore((s) => s.diffSummaryWorkspaceId);
	const setDiffSummaryOpen = useAppStore((s) => s.setDiffSummaryOpen);
	const workspaces = useAppStore((s) => s.workspaces);
	const workspaceDiffCache = useAppStore((s) => s.workspaceDiffCache);
	const openNodeDiff = useAppStore((s) => s.openNodeDiff);
	const updateBaselineToCurrent = useAppStore((s) => s.updateBaselineToCurrent);

	const [filter, setFilter] = useState<FilterKind>('all');

	const ws = workspaces.find((w) => w.id === workspaceId);
	const diff = workspaceId ? workspaceDiffCache[workspaceId] : undefined;

	const filteredNodes = useMemo(
		() => filterNodesByKind(diff?.nodes ?? [], filter),
		[diff?.nodes, filter],
	);

	const badges = diff ? summaryBadgeLabels(diff.summary) : null;

	return (
		<Sheet
			open={open}
			onOpenChange={(v) => {
				setDiffSummaryOpen(v);
				if (!v) setFilter('all');
			}}
		>
			<SheetContent>
				<SheetHeader>
					<SheetTitle>{ws?.name ?? messages.diff.summaryTitle}</SheetTitle>
					{badges && (
						<div className='flex flex-wrap gap-1 pt-1'>
							<Badge variant='outline' className='text-[10px]'>
								{badges.content}
							</Badge>
							<Badge variant='outline' className='text-[10px]'>
								{badges.links}
							</Badge>
							<Badge variant='outline' className='text-[10px]'>
								{badges.fetch}
							</Badge>
						</div>
					)}
					{diff?.hasDiff && (
						<Button
							size='sm'
							variant='outline'
							className='mt-2 w-full border-amber-500 text-amber-600 hover:bg-amber-500/10 hover:text-amber-700'
							onClick={async () => {
								await updateBaselineToCurrent();
								setDiffSummaryOpen(false);
							}}
						>
							<CheckCheck className='size-3.5' />
							{messages.diff.markReviewed}
						</Button>
					)}
				</SheetHeader>
				<Tabs
					value={filter}
					onValueChange={(v) => setFilter(v as FilterKind)}
					className='flex min-h-0 flex-1 flex-col px-4 py-2'
				>
					<TabsList className='shrink-0'>
						<TabsTrigger value='all'>{messages.diff.filterAll}</TabsTrigger>
						<TabsTrigger value='content'>
							{messages.diff.kindContent}
						</TabsTrigger>
						<TabsTrigger value='links'>{messages.diff.kindLinks}</TabsTrigger>
						<TabsTrigger value='fetch'>{messages.diff.kindFetch}</TabsTrigger>
					</TabsList>
					<TabsContent value={filter} className='min-h-0 flex-1'>
						<ScrollArea className='h-[calc(100vh-10rem)]'>
							{filteredNodes.length === 0 ? (
								<p className='py-4 text-xs text-muted-foreground'>
									{messages.diff.emptyNodes}
								</p>
							) : (
								<ul className='space-y-1 py-1'>
									{filteredNodes.map((node) => (
										<li key={node.nodeId}>
											<button
												type='button'
												className={cn(
													'w-full rounded-md px-2 py-2 text-left text-xs hover:bg-muted',
												)}
												onClick={() =>
													void openNodeDiff(node.nodeId, node.kinds[0])
												}
											>
												<p className='truncate font-medium'>{node.url}</p>
												<div className='mt-1 flex flex-wrap gap-1'>
													{node.kinds.map((k) => (
														<Badge
															key={k}
															variant='secondary'
															className='text-[9px]'
														>
															{k}
														</Badge>
													))}
												</div>
											</button>
										</li>
									))}
								</ul>
							)}
						</ScrollArea>
					</TabsContent>
				</Tabs>
			</SheetContent>
		</Sheet>
	);
}
