import type { ReactNode } from 'react';
import { isBrowsableHttpUrl } from '@/lib/externalLinkDelegation';
import { cn } from '@/lib/utils';

type ExternalLinkProps = {
	href: string;
	children?: ReactNode;
	className?: string;
	stopPropagation?: boolean;
};

export function ExternalLink({
	href,
	children,
	className,
	stopPropagation = false,
}: ExternalLinkProps) {
	const label = children ?? href;

	if (!isBrowsableHttpUrl(href)) {
		return <span className={className}>{label}</span>;
	}

	return (
		<a
			href={href}
			target='_blank'
			rel='noopener noreferrer'
			className={cn('text-primary underline underline-offset-2', className)}
			onClick={stopPropagation ? (e) => e.stopPropagation() : undefined}
		>
			{label}
		</a>
	);
}
