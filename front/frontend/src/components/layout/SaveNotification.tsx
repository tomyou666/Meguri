import { Alert } from '@/components/ui/alert';
import { messages } from '@/i18n/messages';
import { cn } from '@/lib/utils';
import { useAppStore } from '@/stores/appStore';

export function SaveNotification() {
	const notification = useAppStore((s) => s.saveNotification);
	const clearSaveNotification = useAppStore((s) => s.clearSaveNotification);

	if (!notification) return null;

	return (
		<div className='pointer-events-none fixed bottom-4 right-4 z-100 flex max-w-sm flex-col gap-2'>
			<Alert
				className={cn(
					'pointer-events-auto border px-3 py-2 text-xs shadow-lg',
					notification.type === 'success'
						? 'border-emerald-500/50 bg-emerald-500/10 text-emerald-100'
						: 'border-destructive/50 bg-destructive/10',
				)}
			>
				<div className='flex items-start justify-between gap-2'>
					<span>
						{notification.type === 'success'
							? messages.settings.saveSuccess
							: messages.settings.saveFailed}
						{notification.detail ? `: ${notification.detail}` : ''}
					</span>
					<button
						type='button'
						className='text-muted-foreground hover:text-foreground'
						onClick={clearSaveNotification}
						aria-label={messages.dialog.cancel}
					>
						×
					</button>
				</div>
			</Alert>
		</div>
	);
}
