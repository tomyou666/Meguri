import { useState } from 'react';
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
import { notifyError } from '@/lib/notify';
import { useAppStore } from '@/stores/appStore';
import * as ProjectService from '../../../bindings/scraperbot-front/internal/usecase/wails_service/projectservice';

export function MenuBar() {
	const appDefaults = useAppStore((s) => s.appDefaults);
	const persistAppDefaults = useAppStore((s) => s.persistAppDefaults);
	const activeWorkspaceId = useAppStore((s) => s.activeWorkspaceId);
	const loadWorkspace = useAppStore((s) => s.loadWorkspaceFromServer);
	const [settingsOpen, setSettingsOpen] = useState(false);

	const handleOpenScrb = async () => {
		try {
			const res = await ProjectService.OpenScrb();
			if (res?.workspaceId) {
				await loadWorkspace(res.workspaceId);
			}
		} catch (e) {
			notifyError(e instanceof Error ? e.message : String(e));
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
			notifyError(e instanceof Error ? e.message : String(e));
		}
	};

	return (
		<>
			<div className='flex h-8 items-center gap-1 border-b border-border bg-card px-2 text-xs'>
				<span className='px-2 font-medium'>{messages.menu.file}</span>
				<Button variant='ghost' size='xs' onClick={() => void handleOpenScrb()}>
					{messages.menu.openScrb}
				</Button>
				<Button
					variant='ghost'
					size='xs'
					onClick={() => void handleSaveScrb()}
					disabled={!activeWorkspaceId}
				>
					{messages.menu.saveScrb}
				</Button>
				<span className='mx-1 text-muted-foreground'>|</span>
				<span className='px-2 font-medium'>{messages.menu.settings}</span>
				<Button variant='ghost' size='xs' onClick={() => setSettingsOpen(true)}>
					{messages.menu.appDefaults}
				</Button>
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
					<ScrollArea className='min-h-0 flex-1 pr-2'>
						<ConfigEditor
							layer='app'
							settings={appDefaults}
							onSave={(config) => persistAppDefaults(config)}
						/>
					</ScrollArea>
				</DialogContent>
			</Dialog>
		</>
	);
}
