import type { LucideIcon } from 'lucide-react';
import { cn } from '@/lib/utils';

type CollapsedSidebarRailProps = {
	icon: LucideIcon;
	label: string;
	onClick: () => void;
	borderSide: 'left' | 'right';
	className?: string;
};

/** 折りたたみ時のサイドバー — 全体をクリック可能な広いヒット領域 */
export function CollapsedSidebarRail({
	icon: Icon,
	label,
	onClick,
	borderSide,
	className,
}: CollapsedSidebarRailProps) {
	return (
		<button
			type='button'
			className={cn(
				'flex h-full w-full min-w-[2.75rem] cursor-pointer flex-col items-center justify-start gap-2 border-border bg-sidebar px-1 py-3 transition-colors hover:bg-sidebar-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
				borderSide === 'left' ? 'border-r' : 'border-l',
				className,
			)}
			onClick={onClick}
			title={label}
			aria-label={label}
		>
			<Icon className='size-5 shrink-0' strokeWidth={2} />
			<span
				className='text-[9px] font-medium text-muted-foreground [writing-mode:vertical-rl]'
				aria-hidden
			>
				{label}
			</span>
		</button>
	);
}
