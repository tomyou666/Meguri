import {
	CircleCheckIcon,
	InfoIcon,
	Loader2Icon,
	OctagonXIcon,
	TriangleAlertIcon,
} from 'lucide-react';
import { Toaster as Sonner, type ToasterProps } from 'sonner';
import { messages } from '@/i18n/messages';
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
				closeButtonAriaLabel: messages.toast.dismissAria,
				classNames: {
					toast: cn(
						'cn-toast',
						'group toast min-w-[320px] items-start border-border bg-popover text-popover-foreground shadow-lg',
					),
					title: 'font-medium',
					description: 'text-foreground/70',
					actionButton: cn(
						'!ml-2 shrink-0 !rounded-[min(var(--radius-md),10px)] !border !border-border !bg-transparent !px-2.5 !py-1 !text-sm !font-medium !text-foreground !shadow-none',
						'whitespace-nowrap',
						'hover:!bg-muted/50',
					),
					cancelButton: cn(
						'!ml-auto !mr-0 !size-6 !min-h-6 !min-w-6 shrink-0 !rounded-[min(var(--radius-md),10px)] !border-0 !bg-transparent !p-0 !text-muted-foreground !shadow-none',
						'hover:!bg-muted/50 hover:!text-foreground',
					),
					warning: 'border-amber-500/30 [&_[data-icon]]:text-amber-500',
				},
			}}
			{...props}
		/>
	);
};

export { Toaster };
