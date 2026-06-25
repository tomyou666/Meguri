import * as React from 'react';
import { cn } from '@/lib/utils';

const TabsContext = React.createContext<{
	value: string;
	onValueChange: (v: string) => void;
} | null>(null);

function Tabs({
	value,
	onValueChange,
	children,
	className,
}: {
	value: string;
	onValueChange: (v: string) => void;
	children: React.ReactNode;
	className?: string;
}) {
	return (
		<TabsContext.Provider value={{ value, onValueChange }}>
			<div className={className}>{children}</div>
		</TabsContext.Provider>
	);
}

function TabsList({
	className,
	children,
}: {
	className?: string;
	children: React.ReactNode;
}) {
	return (
		<div className={cn('flex gap-1 border-b border-border', className)}>
			{children}
		</div>
	);
}

function TabsTrigger({
	value,
	children,
}: {
	value: string;
	children: React.ReactNode;
}) {
	const ctx = React.useContext(TabsContext)!;
	const active = ctx.value === value;
	return (
		<button
			type='button'
			onClick={() => ctx.onValueChange(value)}
			className={cn(
				'px-2 py-1.5 text-xs font-medium transition-colors',
				active
					? 'border-b-2 border-primary text-foreground'
					: 'text-muted-foreground hover:text-foreground',
			)}
		>
			{children}
		</button>
	);
}

function TabsContent({
	value,
	children,
	className,
}: {
	value: string;
	children: React.ReactNode;
	className?: string;
}) {
	const ctx = React.useContext(TabsContext)!;
	if (ctx.value !== value) return null;
	return <div className={cn(className)}>{children}</div>;
}

export { Tabs, TabsContent, TabsList, TabsTrigger };
