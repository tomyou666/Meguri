import type * as React from 'react';
import { cn } from '@/lib/utils';

function Label({ className, ...props }: React.ComponentProps<'label'>) {
	return (
		// biome-ignore lint/a11y/noLabelWithoutControl: This reusable component forwards htmlFor/children from props.
		<label
			className={cn('font-medium text-muted-foreground text-xs', className)}
			{...props}
		/>
	);
}

export { Label };
