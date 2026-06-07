import type * as React from 'react';
import { cn } from '@/lib/utils';

function Dialog({
	open,
	onOpenChange,
	children,
	size = 'default',
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	children: React.ReactNode;
	size?: 'default' | 'full' | 'fullHeight';
}) {
	if (!open) return null;
	return (
		<div className='fixed inset-0 z-50 flex items-center justify-center'>
			<button
				type='button'
				aria-label='Close dialog'
				className='absolute inset-0 bg-black/60'
				onClick={() => onOpenChange(false)}
			/>
			<div
				className={cn(
					'relative z-10 w-full',
					size === 'default' && 'max-w-md',
					size === 'full' &&
						'mx-2 flex h-[calc(100vh-1rem)] max-h-[calc(100vh-1rem)] max-w-[calc(100vw-1rem)]',
					size === 'fullHeight' &&
						'flex h-[calc(100vh-1rem)] max-h-[calc(100vh-1rem)] max-w-lg w-full',
				)}
			>
				{children}
			</div>
		</div>
	);
}

function DialogContent({
	className,
	children,
}: {
	className?: string;
	children: React.ReactNode;
}) {
	return (
		<div
			className={cn(
				'mx-4 rounded-lg border border-border bg-card p-4 text-card-foreground shadow-lg',
				className,
			)}
		>
			{children}
		</div>
	);
}

function DialogHeader({ children }: { children: React.ReactNode }) {
	return <div className='mb-4 space-y-1'>{children}</div>;
}

function DialogTitle({ children }: { children: React.ReactNode }) {
	return <h2 className='text-sm font-semibold'>{children}</h2>;
}

function DialogFooter({
	className,
	children,
}: {
	className?: string;
	children: React.ReactNode;
}) {
	return (
		<div className={cn('mt-4 flex justify-end gap-2', className)}>
			{children}
		</div>
	);
}

export { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle };
