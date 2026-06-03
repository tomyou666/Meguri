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
import { useAppStore } from '@/stores/appStore';

export function MenuBar() {
	const appDefaults = useAppStore((s) => s.appDefaults);
	const persistAppDefaults = useAppStore((s) => s.persistAppDefaults);
	const [settingsOpen, setSettingsOpen] = useState(false);

	return (
		<>
			<div className='flex h-8 items-center gap-1 border-b border-border bg-card px-2 text-xs'>
				<span className='px-2 font-medium'>{messages.menu.file}</span>
				<Button variant='ghost' size='xs' disabled title={messages.scrbPhase2}>
					{messages.menu.openScrb}
				</Button>
				<Button variant='ghost' size='xs' disabled title={messages.scrbPhase2}>
					{messages.menu.saveScrb}
				</Button>
				<span className='mx-1 text-muted-foreground'>|</span>
				<span className='px-2 font-medium'>{messages.menu.settings}</span>
				<Button variant='ghost' size='xs' onClick={() => setSettingsOpen(true)}>
					{messages.menu.appDefaults}
				</Button>
			</div>

			<Dialog open={settingsOpen} onOpenChange={setSettingsOpen}>
				<DialogContent className='max-h-[85vh] max-w-lg'>
					<DialogHeader>
						<DialogTitle>{messages.menu.appDefaults}</DialogTitle>
					</DialogHeader>
					<ScrollArea className='max-h-[60vh] pr-2'>
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
