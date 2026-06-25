import { cn } from '@/lib/utils';

function ScrollArea({
	className,
	children,
}: {
	className?: string;
	children: React.ReactNode;
}) {
	return (
		<div
			className={cn(
				'overflow-y-auto overflow-x-hidden scrollbar-gutter-stable',
				className,
			)}
		>
			{children}
		</div>
	);
}

export { ScrollArea };
