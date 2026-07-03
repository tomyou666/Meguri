import { describe, expect, it, vi } from 'vitest';

vi.mock('@wailsio/runtime', () => ({
	Browser: { OpenURL: vi.fn() },
}));

import {
	isBrowsableHttpUrl,
	resolvePreviewBrowsableUrl,
} from '@/lib/externalLinkDelegation';

describe('isBrowsableHttpUrl', () => {
	it('http/https の絶対 URL は true', () => {
		expect(isBrowsableHttpUrl('https://example.com/page')).toBe(true);
		expect(isBrowsableHttpUrl('http://example.com')).toBe(true);
	});

	it('相対パス・javascript・空文字は false', () => {
		expect(isBrowsableHttpUrl('/about')).toBe(false);
		expect(isBrowsableHttpUrl('javascript:alert(1)')).toBe(false);
		expect(isBrowsableHttpUrl('')).toBe(false);
		expect(isBrowsableHttpUrl('mailto:test@example.com')).toBe(false);
	});
});

describe('resolvePreviewBrowsableUrl', () => {
	it('相対パスを基準 URL に対して絶対化する', () => {
		expect(resolvePreviewBrowsableUrl('/foo', 'https://example.com/page')).toBe(
			'https://example.com/foo',
		);
		expect(resolvePreviewBrowsableUrl('bar', 'https://example.com/page/')).toBe(
			'https://example.com/page/bar',
		);
	});

	it('絶対 http(s) URL はそのまま返す', () => {
		expect(
			resolvePreviewBrowsableUrl(
				'https://other.com/path',
				'https://example.com/page',
			),
		).toBe('https://other.com/path');
	});

	it('#anchor・空・javascript・不正 base は null', () => {
		expect(resolvePreviewBrowsableUrl('#section', 'https://example.com')).toBe(
			null,
		);
		expect(resolvePreviewBrowsableUrl('', 'https://example.com')).toBe(null);
		expect(
			resolvePreviewBrowsableUrl('javascript:alert(1)', 'https://example.com'),
		).toBe(null);
		expect(resolvePreviewBrowsableUrl('/foo', 'not-a-url')).toBe(null);
	});
});
