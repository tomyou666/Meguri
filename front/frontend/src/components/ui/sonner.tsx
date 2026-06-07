import {
	CircleCheckIcon,
	InfoIcon,
	Loader2Icon,
	OctagonXIcon,
	TriangleAlertIcon,
} from 'lucide-react';
import { Toaster as Sonner, type ToasterProps } from 'sonner';
import { cn } from '@/lib/utils';

const Toaster = ({ ...props }: ToasterProps) => {
	return (
		<Sonner
			theme='system'
			className='toaster group'
			icons={{
				success: <CircleCheckIcon className='size-4' />,
				info: <InfoIcon className='size-4' />,
				warning: <TriangleAlertIcon className='size-4' />,
				error: <OctagonXIcon className='size-4' />,
				loading: <Loader2Icon className='size-4 animate-spin' />,
			}}
			style={
				{
					'--normal-bg': 'var(--popover)',
					'--normal-text': 'var(--popover-foreground)',
					'--normal-border': 'var(--border)',
					'--border-radius': 'var(--radius)',
				} as React.CSSProperties
			}
			toastOptions={{
				classNames: {
					toast: cn(
						'cn-toast',
						'group toast border-border bg-popover text-popover-foreground shadow-lg',
					),
					description: 'text-muted-foreground',
					actionButton: cn(
						'!ml-auto !mr-0 !size-6 !min-h-6 !min-w-6 shrink-0 !rounded-[min(var(--radius-md),10px)] !border-0 !bg-transparent !p-0 !text-muted-foreground !shadow-none',
						'hover:!bg-muted/50 hover:!text-foreground',
					),
				},
			}}
			{...props}
		/>
	);
};

export { Toaster };
