import type * as React from 'react';

import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from '@/components/ui/tooltip';

type ActionTooltipProps = {
	label: string;
	children: React.ReactElement;
} & Pick<
	React.ComponentProps<typeof TooltipContent>,
	'side' | 'align' | 'sideOffset'
>;

/** アイコンボタン等に Radix ツールチップを付ける */
export function ActionTooltip({
	label,
	children,
	side,
	align,
	sideOffset,
}: ActionTooltipProps) {
	return (
		<Tooltip>
			<TooltipTrigger asChild>{children}</TooltipTrigger>
			<TooltipContent side={side} align={align} sideOffset={sideOffset}>
				{label}
			</TooltipContent>
		</Tooltip>
	);
}
