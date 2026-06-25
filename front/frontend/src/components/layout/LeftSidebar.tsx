import {
	Copy,
	GitCompare,
	Menu,
	PanelLeftClose,
	PanelLeftOpen,
	Plus,
	Settings,
	Trash2,
} from 'lucide-react';
import { useState } from 'react';
import { CollapsedSidebarRail } from '@/components/layout/CollapsedSidebarRail';
import { DomainStatusPanel } from '@/components/layout/DomainStatusPanel';
import { ConfigEditor } from '@/components/settings/ConfigEditor';
import { ActionTooltip } from '@/components/ui/action-tooltip';
import { Button } from '@/components/ui/button';
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ScrollArea } from '@/components/ui/scroll-area';
import { messages } from '@/i18n/messages';
import { cn } from '@/lib/utils';
import { useAppStore } from '@/stores/appStore';

export function LeftSidebarContent() {
	const workspaces = useAppStore((s) => s.workspaces);
	const activeWorkspaceId = useAppStore((s) => s.activeWorkspaceId);
	const leftCollapsed = useAppStore((s) => s.leftSidebarCollapsed);
	const workspaceDiffCache = useAppStore((s) => s.workspaceDiffCache);
	const setActiveWorkspace = useAppStore((s) => s.setActiveWorkspace);
	const openDeleteWorkspaceDialog = useAppStore(
		(s) => s.openDeleteWorkspaceDialog,
	);
	const openNewWorkspaceDialog = useAppStore((s) => s.openNewWorkspaceDialog);
	const openDuplicateWorkspaceDialog = useAppStore(
		(s) => s.openDuplicateWorkspaceDialog,
	);
	const fetchWorkspaceDiff = useAppStore((s) => s.fetchWorkspaceDiff);
	const toggleLeftSidebar = useAppStore((s) => s.toggleLeftSidebar);
	const persistWorkspaceSettings = useAppStore(
		(s) => s.persistWorkspaceSettings,
	);
	const activeWorkspace = useAppStore((s) =>
		s.workspaces.find((w) => w.id === s.activeWorkspaceId),
	);
	const appDefaults = useAppStore((s) => s.appDefaults);
	const [wsSettingsOpen, setWsSettingsOpen] = useState(false);
	const [diffDialogWs, setDiffDialogWs] = useState<string | null>(null);

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
						<ActionTooltip label={messages.sidebar.newWorkspace}>
							<Button
								variant='ghost'
								size='icon-xs'
								aria-label={messages.sidebar.newWorkspace}
								onClick={openNewWorkspaceDialog}
							>
								<Plus className='size-3.5' />
							</Button>
						</ActionTooltip>
						<ActionTooltip label={messages.sidebar.closeLeft}>
							<Button
								variant='ghost'
								size='icon-xs'
								aria-label={messages.sidebar.closeLeft}
								onClick={toggleLeftSidebar}
							>
								<PanelLeftClose className='size-3.5' />
							</Button>
						</ActionTooltip>
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
									<ActionTooltip label={messages.sidebar.workspaceSettings}>
										<Button
											variant='ghost'
											size='icon-xs'
											aria-label={messages.sidebar.workspaceSettings}
											onClick={(e) => {
												e.stopPropagation();
												setActiveWorkspace(ws.id);
												setWsSettingsOpen(true);
											}}
										>
											<Settings className='size-3' />
										</Button>
									</ActionTooltip>
									<ActionTooltip label={messages.sidebar.deleteWorkspace}>
										<Button
											variant='ghost'
											size='icon-xs'
											aria-label={messages.sidebar.deleteWorkspace}
											onClick={(e) => {
												e.stopPropagation();
												openDeleteWorkspaceDialog(ws.id);
											}}
										>
											<Trash2 className='size-3' />
										</Button>
									</ActionTooltip>
									<DropdownMenu>
										<DropdownMenuTrigger asChild>
											<Button
												variant='ghost'
												size='icon-xs'
												aria-label={messages.sidebar.openWorkspaceMenu}
												onClick={(e) => e.stopPropagation()}
											>
												<Menu className='size-3' />
											</Button>
										</DropdownMenuTrigger>
										<DropdownMenuContent
											align='end'
											sideOffset={6}
											className='min-w-44 w-auto border-border p-1 shadow-lg'
										>
											<DropdownMenuLabel className='max-w-44 truncate px-2 py-1 text-xs font-normal text-muted-foreground'>
												{ws.name}
											</DropdownMenuLabel>
											<DropdownMenuSeparator className='my-1' />
											<DropdownMenuItem
												className='gap-2 px-2 py-1.5 text-xs'
												onClick={() => openDuplicateWorkspaceDialog(ws.id)}
											>
												<Copy className='size-3.5 text-muted-foreground' />
												{messages.sidebar.duplicateWorkspace}
											</DropdownMenuItem>
											<DropdownMenuItem
												className='gap-2 px-2 py-1.5 text-xs'
												onClick={async () => {
													await fetchWorkspaceDiff(ws.id);
													setDiffDialogWs(ws.id);
												}}
											>
												<GitCompare className='size-3.5 text-muted-foreground' />
												<span className='flex-1'>
													{messages.sidebar.diffSummary}
												</span>
												{diff?.hasDiff && (
													<span
														className='size-1.5 shrink-0 rounded-full bg-amber-500'
														aria-hidden
													/>
												)}
											</DropdownMenuItem>
										</DropdownMenuContent>
									</DropdownMenu>
								</div>
							);
						})
					)}
				</ScrollArea>

				<div className='flex flex-1 flex-col border-t border-sidebar-border'>
					<div className='px-2 py-2'>
						<span className='text-xs font-semibold'>
							{messages.sidebar.domainStatus}
						</span>
					</div>
					<ScrollArea className='flex-1 px-1 pb-2'>
						{activeWorkspace ? (
							<DomainStatusPanel
								nodes={activeWorkspace.nodes}
								appDefaults={appDefaults}
								wsSettings={activeWorkspace.settings}
							/>
						) : (
							<p className='px-2 py-2 text-xs text-muted-foreground'>
								{messages.sidebar.emptyDomains}
							</p>
						)}
					</ScrollArea>
				</div>
			</aside>

			{activeWorkspace && (
				<Dialog
					open={wsSettingsOpen}
					onOpenChange={setWsSettingsOpen}
					size='fullHeight'
				>
					<DialogContent className='flex h-full w-full min-w-0 flex-col overflow-hidden'>
						<DialogHeader>
							<DialogTitle>{messages.sidebar.workspaceSettings}</DialogTitle>
						</DialogHeader>
						<div className='min-h-0 flex-1'>
							<ConfigEditor
								layer='workspace'
								settings={activeWorkspace.settings}
								onSave={(settings) => persistWorkspaceSettings(settings)}
							/>
						</div>
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
