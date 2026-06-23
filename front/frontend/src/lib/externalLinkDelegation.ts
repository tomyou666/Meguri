import { Browser } from '@wailsio/runtime';

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

async function openExternalUrl(url: string): Promise<void> {
	try {
		await Browser.OpenURL(url);
	} catch {
		window.open(url, '_blank', 'noopener,noreferrer');
	}
}

function handleDocumentClick(event: MouseEvent): void {
	if (event.defaultPrevented || event.button !== 0) {
		return;
	}

	const anchor = (event.target as Element | null)?.closest('a');
	if (!anchor?.href || !shouldOpenInExternalBrowser(anchor.href)) {
		return;
	}

	event.preventDefault();
	void openExternalUrl(anchor.href);
}

/**
 * 外部 http(s) リンクをクリックしたとき、WebView 内遷移ではなく既定ブラウザで開く。
 */
export function installExternalLinkDelegation(): () => void {
	document.addEventListener('click', handleDocumentClick);
	return () => document.removeEventListener('click', handleDocumentClick);
}
