import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const badgeVariants = cva(
	'inline-flex items-center rounded-md border px-1.5 py-0.5 font-medium text-xs',
	{
		variants: {
			variant: {
				default: 'border-transparent bg-primary text-primary-foreground',
				secondary: 'border-transparent bg-secondary text-secondary-foreground',
				outline: 'text-foreground',
				destructive: 'border-transparent bg-destructive/20 text-destructive',
				success: 'border-transparent bg-emerald-500/20 text-emerald-400',
				warning: 'border-transparent bg-amber-500/20 text-amber-400',
			},
		},
		defaultVariants: { variant: 'default' },
	},
);

function Badge({
	className,
	variant,
	...props
}: React.ComponentProps<'span'> & VariantProps<typeof badgeVariants>) {
	return (
		<span className={cn(badgeVariants({ variant }), className)} {...props} />
	);
}

export { Badge, badgeVariants };
