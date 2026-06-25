import { cn } from '@/lib/utils';

function ScrollArea({
	className,
	children,
	scrollbarGutter = 'auto',
}: {
	className?: string;
	children: React.ReactNode;
	scrollbarGutter?: 'stable' | 'auto';
}) {
	return (
		<div
			className={cn(
				'overflow-y-auto overflow-x-hidden',
				scrollbarGutter === 'stable' && 'scrollbar-gutter-stable',
				className,
			)}
		>
			{children}
		</div>
	);
}

export { ScrollArea };
