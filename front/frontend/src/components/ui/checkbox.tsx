import type * as React from 'react';
import { forwardRef } from 'react';
import { cn } from '@/lib/utils';

const Checkbox = forwardRef<
	HTMLInputElement,
	Omit<React.ComponentProps<'input'>, 'type' | 'onChange'> & {
		onCheckedChange?: (checked: boolean) => void;
	}
>(function Checkbox({ className, checked, onCheckedChange, ...props }, ref) {
	return (
		<input
			ref={ref}
			type='checkbox'
			checked={checked}
			onChange={(e) => onCheckedChange?.(e.target.checked)}
			className={cn('size-4 accent-primary', className)}
			{...props}
		/>
	);
});

export { Checkbox };
