import { Browser } from '@wailsio/runtime';

export const PREVIEW_BASE_URL_ATTR = 'data-preview-base-url';

export function isBrowsableHttpUrl(href: string): boolean {
	try {
		const url = new URL(href);
		return url.protocol === 'http:' || url.protocol === 'https:';
	} catch {
		return false;
	}
}

function shouldOpenInExternalBrowser(href: string): boolean {
	try {
		const url = new URL(href, window.location.href);
		if (url.protocol !== 'http:' && url.protocol !== 'https:') {
			return false;
		}
		return url.origin !== window.location.origin;
	} catch {
		return false;
	}
}

function isNonNavigableHref(rawHref: string): boolean {
	const trimmed = rawHref.trim();
	if (!trimmed || trimmed.startsWith('#')) {
		return true;
	}
	const lower = trimmed.toLowerCase();
	return (
		lower.startsWith('javascript:') ||
		lower.startsWith('mailto:') ||
		lower.startsWith('tel:')
	);
}

/** プレビュー基準 URL に対して相対・絶対 href を閲覧可能な http(s) URL に解決する。 */
export function resolvePreviewBrowsableUrl(
	rawHref: string,
	baseUrl: string,
): string | null {
	if (isNonNavigableHref(rawHref)) {
		return null;
	}
	try {
		const resolved = new URL(rawHref.trim(), baseUrl).href;
		return isBrowsableHttpUrl(resolved) ? resolved : null;
	} catch {
		return null;
	}
}

export async function openExternalBrowserUrl(url: string): Promise<void> {
	try {
		await Browser.OpenURL(url);
	} catch {
		window.open(url, '_blank', 'noopener,noreferrer');
	}
}

function resolveExternalLinkUrl(anchor: HTMLAnchorElement): string | null {
	const rawHref = anchor.getAttribute('href');
	if (!rawHref) {
		return null;
	}

	if (isBrowsableHttpUrl(rawHref)) {
		return shouldOpenInExternalBrowser(rawHref) ? rawHref : null;
	}

	const previewBase = anchor
		.closest(`[${PREVIEW_BASE_URL_ATTR}]`)
		?.getAttribute(PREVIEW_BASE_URL_ATTR);
	if (!previewBase) {
		return null;
	}

	return resolvePreviewBrowsableUrl(rawHref, previewBase);
}

function handleDocumentClick(event: MouseEvent): void {
	if (event.defaultPrevented || event.button !== 0) {
		return;
	}

	const anchor = (event.target as Element | null)?.closest('a');
	if (!anchor) {
		return;
	}

	const externalUrl = resolveExternalLinkUrl(anchor);
	if (!externalUrl) {
		return;
	}

	event.preventDefault();
	void openExternalBrowserUrl(externalUrl);
}

/**
 * 外部 http(s) リンクをクリックしたとき、WebView 内遷移ではなく既定ブラウザで開く。
 */
export function installExternalLinkDelegation(): () => void {
	document.addEventListener('click', handleDocumentClick);
	return () => document.removeEventListener('click', handleDocumentClick);
}
