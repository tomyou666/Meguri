import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';

function Sheet({
	open,
	onOpenChange,
	children,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	children: ReactNode;
}) {
	if (!open) return null;
	return (
		<div className='fixed inset-0 z-50 flex justify-end'>
			<button
				type='button'
				aria-label='Close sheet'
				className='absolute inset-0 bg-black/40'
				onClick={() => onOpenChange(false)}
			/>
			{children}
		</div>
	);
}

function SheetContent({
	className,
	children,
}: {
	className?: string;
	children: ReactNode;
}) {
	return (
		<div
			className={cn(
				'relative z-10 flex h-full w-full max-w-md flex-col border-l border-border bg-card shadow-xl',
				className,
			)}
		>
			{children}
		</div>
	);
}

function SheetHeader({
	className,
	children,
}: {
	className?: string;
	children: ReactNode;
}) {
	return (
		<div
			className={cn(
				'flex flex-col gap-1 border-b border-border px-4 py-3',
				className,
			)}
		>
			{children}
		</div>
	);
}

function SheetTitle({
	className,
	children,
}: {
	className?: string;
	children: ReactNode;
}) {
	return <h2 className={cn('text-sm font-semibold', className)}>{children}</h2>;
}

export { Sheet, SheetContent, SheetHeader, SheetTitle };
