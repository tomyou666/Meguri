import {
	Copy,
	GitCompare,
	PanelLeftClose,
	PanelLeftOpen,
	Plus,
	Trash2,
} from 'lucide-react';
import { useMemo, useState } from 'react';
import { CollapsedSidebarRail } from '@/components/layout/CollapsedSidebarRail';
import { ConfigEditor } from '@/components/settings/ConfigEditor';
import { Button } from '@/components/ui/button';
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import { ScrollArea } from '@/components/ui/scroll-area';
import { messages } from '@/i18n/messages';
import { hostFromUrl } from '@/lib/normalizeUrl';
import { cn } from '@/lib/utils';
import { useAppStore } from '@/stores/appStore';

export function LeftSidebarContent() {
	const workspaces = useAppStore((s) => s.workspaces);
	const activeWorkspaceId = useAppStore((s) => s.activeWorkspaceId);
	const leftCollapsed = useAppStore((s) => s.leftSidebarCollapsed);
	const workspaceDiffCache = useAppStore((s) => s.workspaceDiffCache);
	const setActiveWorkspace = useAppStore((s) => s.setActiveWorkspace);
	const deleteWorkspace = useAppStore((s) => s.deleteWorkspace);
	const openNewWorkspaceDialog = useAppStore((s) => s.openNewWorkspaceDialog);
	const openAddNodeDialog = useAppStore((s) => s.openAddNodeDialog);
	const openDeleteNodeDialog = useAppStore((s) => s.openDeleteNodeDialog);
	const duplicateWorkspace = useAppStore((s) => s.duplicateWorkspace);
	const fetchWorkspaceDiff = useAppStore((s) => s.fetchWorkspaceDiff);
	const toggleLeftSidebar = useAppStore((s) => s.toggleLeftSidebar);
	const persistWorkspaceSettings = useAppStore(
		(s) => s.persistWorkspaceSettings,
	);
	const selectedNodeId = useAppStore((s) => s.selectedNodeId);
	const selectedDomain = useAppStore((s) => s.selectedDomain);
	const selectDomain = useAppStore((s) => s.selectDomain);
	const activeWorkspace = useAppStore((s) =>
		s.workspaces.find((w) => w.id === s.activeWorkspaceId),
	);
	const crawlStatus = useAppStore((s) => s.crawlStatus);
	const [wsSettingsOpen, setWsSettingsOpen] = useState(false);
	const [diffDialogWs, setDiffDialogWs] = useState<string | null>(null);

	const domains = useMemo(() => {
		if (!activeWorkspace) return [];
		const hosts = new Set(
			activeWorkspace.nodes.map((n) => hostFromUrl(n.urlNormalized)),
		);
		return [...hosts].sort();
	}, [activeWorkspace]);

	if (leftCollapsed) {
		return (
			<CollapsedSidebarRail
				icon={PanelLeftOpen}
				label={messages.sidebar.openLeft}
				onClick={toggleLeftSidebar}
				borderSide='left'
			/>
		);
	}

	return (
		<>
			<aside className='flex h-full w-full min-w-[14rem] flex-col overflow-hidden border-r border-border bg-sidebar'>
				<div className='flex items-center justify-between border-b border-sidebar-border px-2 py-2'>
					<span className='text-xs font-semibold'>
						{messages.sidebar.workspaces}
					</span>
					<div className='flex gap-0.5'>
						<Button
							variant='ghost'
							size='icon-xs'
							onClick={openNewWorkspaceDialog}
						>
							<Plus className='size-3.5' />
						</Button>
						<Button variant='ghost' size='icon-xs' onClick={toggleLeftSidebar}>
							<PanelLeftClose className='size-3.5' />
						</Button>
					</div>
				</div>
				<ScrollArea className='max-h-40 flex-none px-1 py-1'>
					{workspaces.length === 0 ? (
						<p className='px-2 py-2 text-xs text-muted-foreground'>
							{messages.sidebar.emptyWorkspaces}
						</p>
					) : (
						workspaces.map((ws) => {
							const diff = workspaceDiffCache[ws.id];
							return (
								<div
									key={ws.id}
									className={cn(
										'flex w-full items-center gap-0.5 rounded-md px-1 py-0.5',
										activeWorkspaceId === ws.id && 'bg-sidebar-accent',
									)}
								>
									<button
										type='button'
										className='min-w-0 flex-1 truncate px-1 py-1 text-left text-xs hover:underline'
										onClick={() => setActiveWorkspace(ws.id)}
									>
										{ws.name}
										{diff?.hasDiff && (
											<span className='ml-1 text-amber-500'>●</span>
										)}
									</button>
									<Button
										variant='ghost'
										size='icon-xs'
										title='差分確認'
										onClick={async (e) => {
											e.stopPropagation();
											await fetchWorkspaceDiff(ws.id);
											setDiffDialogWs(ws.id);
										}}
									>
										<GitCompare className='size-3' />
									</Button>
									<Button
										variant='ghost'
										size='icon-xs'
										title='ワークスペースをコピー'
										onClick={(e) => {
											e.stopPropagation();
											void duplicateWorkspace(ws.id);
										}}
									>
										<Copy className='size-3' />
									</Button>
									<Button
										variant='ghost'
										size='icon-xs'
										onClick={(e) => {
											e.stopPropagation();
											deleteWorkspace(ws.id);
										}}
									>
										<Trash2 className='size-3' />
									</Button>
								</div>
							);
						})
					)}
				</ScrollArea>

				{activeWorkspace && (
					<Button
						variant='outline'
						size='xs'
						className='mx-2 my-1'
						onClick={() => setWsSettingsOpen(true)}
					>
						WS 設定
					</Button>
				)}

				<div className='flex flex-1 flex-col border-t border-sidebar-border'>
					<div className='flex items-center justify-between px-2 py-2'>
						<span className='text-xs font-semibold'>
							{messages.sidebar.domains}
						</span>
						<div className='flex gap-0.5'>
							<Button
								variant='ghost'
								size='icon-xs'
								onClick={() => openAddNodeDialog()}
							>
								<Plus className='size-3.5' />
							</Button>
							<Button
								variant='ghost'
								size='icon-xs'
								disabled={!selectedNodeId || crawlStatus !== 'idle'}
								onClick={openDeleteNodeDialog}
							>
								<Trash2 className='size-3.5' />
							</Button>
						</div>
					</div>
					<ScrollArea className='flex-1 px-1 pb-2'>
						{domains.length === 0 ? (
							<p className='px-2 py-2 text-xs text-muted-foreground'>
								{messages.sidebar.emptyDomains}
							</p>
						) : (
							domains.map((host) => (
								<button
									key={host}
									type='button'
									className={cn(
										'block w-full truncate rounded-md px-2 py-1.5 text-left text-xs hover:bg-sidebar-accent',
										selectedDomain === host && 'bg-sidebar-accent font-medium',
									)}
									onClick={() => selectDomain(host)}
								>
									{host}
								</button>
							))
						)}
					</ScrollArea>
				</div>
			</aside>

			{activeWorkspace && (
				<Dialog open={wsSettingsOpen} onOpenChange={setWsSettingsOpen}>
					<DialogContent className='max-h-[85vh] max-w-lg overflow-y-auto'>
						<DialogHeader>
							<DialogTitle>{messages.sidebar.workspaceSettings}</DialogTitle>
						</DialogHeader>
						<ConfigEditor
							layer='workspace'
							settings={activeWorkspace.settings}
							onSave={(settings) => persistWorkspaceSettings(settings)}
						/>
					</DialogContent>
				</Dialog>
			)}

			{diffDialogWs && workspaceDiffCache[diffDialogWs] && (
				<Dialog
					open={!!diffDialogWs}
					onOpenChange={() => setDiffDialogWs(null)}
				>
					<DialogContent>
						<DialogHeader>
							<DialogTitle>{messages.sidebar.diffSummary}</DialogTitle>
						</DialogHeader>
						<pre className='text-xs'>
							{JSON.stringify(workspaceDiffCache[diffDialogWs], null, 2)}
						</pre>
					</DialogContent>
				</Dialog>
			)}
		</>
	);
}

/** @deprecated AppShell 内の Panel でラップするため LeftSidebarContent を使用 */
export const LeftSidebar = LeftSidebarContent;
