import {
	ChevronDown,
	FolderOpen,
	MessageSquare,
	RefreshCw,
	Save,
	Settings,
} from 'lucide-react';
import { useState } from 'react';
import { ConfigEditor } from '@/components/settings/ConfigEditor';
import { Badge } from '@/components/ui/badge';
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
import { messages } from '@/i18n/messages';
import { openExternalBrowserUrl } from '@/lib/externalLinkDelegation';
import { getFeedbackUrl } from '@/lib/feedbackUrl';
import { notifyError, notifySuccess } from '@/lib/notify';
import { handleUpdatePromptResult } from '@/lib/updateFlow';
import { useAppStore } from '@/stores/appStore';
import * as ProjectService from '../../../bindings/meguri-app/internal/usecase/wails_service/projectservice';
import * as UpdateService from '../../../bindings/meguri-app/internal/usecase/wails_service/updateservice';

type MenuBarProps = {
	updateAvailable: boolean;
	refreshUpdateStatus: () => Promise<void>;
};

export function MenuBar({
	updateAvailable,
	refreshUpdateStatus,
}: MenuBarProps) {
	const appDefaults = useAppStore((s) => s.appDefaults);
	const persistAppDefaults = useAppStore((s) => s.persistAppDefaults);
	const activeWorkspaceId = useAppStore((s) => s.activeWorkspaceId);
	const loadWorkspace = useAppStore((s) => s.loadWorkspaceFromServer);
	const [settingsOpen, setSettingsOpen] = useState(false);
	const [checkingUpdates, setCheckingUpdates] = useState(false);
	const feedbackUrl = getFeedbackUrl();

	const handleOpenScrb = async () => {
		try {
			const res = await ProjectService.OpenScrb();
			if (res?.workspaceId) {
				await loadWorkspace(res.workspaceId);
			}
		} catch (e) {
			const errMessage = e instanceof Error ? e.message : String(e);
			if (errMessage.includes('cancelled by user')) return;
			notifyError(errMessage);
		}
	};

	const handleSaveScrb = async () => {
		if (!activeWorkspaceId) {
			notifyError('ワークスペースが選択されていません');
			return;
		}
		try {
			await ProjectService.SaveScrb(activeWorkspaceId);
		} catch (e) {
			const errMessage = e instanceof Error ? e.message : String(e);
			if (errMessage.includes('cancelled by user')) return;
			notifyError(errMessage);
		}
	};

	const handleCheckForUpdates = async () => {
		if (checkingUpdates) {
			return;
		}
		setCheckingUpdates(true);
		try {
			const result = await UpdateService.CheckForUpdates();
			await refreshUpdateStatus();
			if (result.status === 'up_to_date') {
				notifySuccess(messages.update.upToDate);
				return;
			}
			await handleUpdatePromptResult(result.action, result.releaseURL);
			await refreshUpdateStatus();
		} catch (e) {
			const msg = e instanceof Error ? e.message : String(e);
			if (msg.includes('updater unavailable')) {
				notifyError(messages.update.unavailable);
				return;
			}
			notifyError(messages.update.checkFailed, { description: msg });
		} finally {
			setCheckingUpdates(false);
		}
	};

	return (
		<>
			<div className='flex h-8 items-center gap-1 border-b border-border bg-card px-2 text-xs'>
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button
							variant='ghost'
							size='xs'
							aria-label={messages.menu.openFileMenu}
						>
							{messages.menu.file}
							<ChevronDown className='size-3.5' />
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent
						align='start'
						sideOffset={6}
						className='min-w-44 w-auto border-border p-1 shadow-lg'
					>
						<DropdownMenuLabel className='px-2 py-1 text-xs font-normal text-muted-foreground'>
							{messages.menu.file}
						</DropdownMenuLabel>
						<DropdownMenuSeparator className='my-1' />
						<DropdownMenuItem
							className='gap-2 px-2 py-1.5 text-xs'
							onClick={() => void handleOpenScrb()}
						>
							<FolderOpen className='size-3.5 text-muted-foreground' />
							{messages.menu.openScrb}
						</DropdownMenuItem>
						<DropdownMenuItem
							className='gap-2 px-2 py-1.5 text-xs'
							disabled={!activeWorkspaceId}
							onClick={() => void handleSaveScrb()}
						>
							<Save className='size-3.5 text-muted-foreground' />
							{messages.menu.saveScrb}
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
				<span className='mx-1 text-muted-foreground'>|</span>
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button
							variant='ghost'
							size='xs'
							aria-label={messages.menu.openSettingsMenu}
						>
							{messages.menu.settings}
							<ChevronDown className='size-3.5' />
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent
						align='start'
						sideOffset={6}
						className='min-w-44 w-auto border-border p-1 shadow-lg'
					>
						<DropdownMenuLabel className='px-2 py-1 text-xs font-normal text-muted-foreground'>
							{messages.menu.settings}
						</DropdownMenuLabel>
						<DropdownMenuSeparator className='my-1' />
						<DropdownMenuItem
							className='gap-2 px-2 py-1.5 text-xs'
							onClick={() => setSettingsOpen(true)}
						>
							<Settings className='size-3.5 text-muted-foreground' />
							{messages.menu.appDefaults}
						</DropdownMenuItem>
						<DropdownMenuItem
							className='gap-2 px-2 py-1.5 text-xs'
							disabled={checkingUpdates}
							onClick={() => void handleCheckForUpdates()}
						>
							<RefreshCw
								className={`size-3.5 text-muted-foreground${checkingUpdates ? ' animate-spin' : ''}`}
							/>
							<span className='flex flex-1 items-center justify-between gap-2'>
								{messages.menu.checkForUpdates}
								{updateAvailable ? (
									<Badge
										variant='destructive'
										className='h-4 min-w-4 px-1 text-[10px] leading-none'
										aria-label={messages.update.badgeAria}
									>
										!
									</Badge>
								) : null}
							</span>
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
				{feedbackUrl ? (
					<Button
						variant='ghost'
						size='xs'
						className='ml-auto gap-1'
						aria-label={messages.menu.openFeedback}
						onClick={() => void openExternalBrowserUrl(feedbackUrl)}
					>
						<MessageSquare className='size-3.5' />
						{messages.menu.feedback}
					</Button>
				) : null}
			</div>

			<Dialog
				open={settingsOpen}
				onOpenChange={setSettingsOpen}
				size='fullHeight'
			>
				<DialogContent className='flex h-full flex-col overflow-hidden'>
					<DialogHeader>
						<DialogTitle>{messages.menu.appDefaults}</DialogTitle>
					</DialogHeader>
					<div className='min-h-0 flex-1'>
						<ConfigEditor
							layer='app'
							settings={appDefaults}
							onSave={(config) => persistAppDefaults(config)}
						/>
					</div>
				</DialogContent>
			</Dialog>
		</>
	);
}
