import { useEffect } from 'react';
import { Skeleton } from '@/components/ui/skeleton';
import { messages } from '@/i18n/messages';
import { useAppStore } from '@/stores/appStore';

export function AppBootstrap({ children }: { children: React.ReactNode }) {
	const bootstrapped = useAppStore((s) => s.bootstrapped);
	const bootstrap = useAppStore((s) => s.bootstrap);

	useEffect(() => {
		bootstrap();
	}, [bootstrap]);

	if (!bootstrapped) {
		return (
			<div className='flex h-screen flex-col gap-4 bg-background p-4'>
				<Skeleton className='h-8 w-full' />
				<Skeleton className='h-10 w-full' />
				<div className='flex flex-1 gap-4'>
					<Skeleton className='h-full w-56' />
					<Skeleton className='h-full flex-1' />
					<Skeleton className='h-full w-72' />
				</div>
				<p className='text-center text-muted-foreground text-xs'>
					{messages.bootstrapLoading}
				</p>
			</div>
		);
	}

	return <>{children}</>;
}
