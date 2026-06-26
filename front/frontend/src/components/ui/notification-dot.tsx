import { cn } from '@/lib/utils';

export function NotificationDot({ className }: { className?: string }) {
	return (
		<span
			className={cn('size-2 shrink-0 rounded-full bg-destructive', className)}
			aria-hidden
		/>
	);
}
