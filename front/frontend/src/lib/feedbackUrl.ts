import { isBrowsableHttpUrl } from '@/lib/externalLinkDelegation';

export function getFeedbackUrl(): string | null {
	const url = import.meta.env.VITE_FEEDBACK_URL?.trim();
	if (!url || !isBrowsableHttpUrl(url)) {
		return null;
	}
	return url;
}
