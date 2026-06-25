import { describe, expect, it, vi } from 'vitest';

vi.mock('@wailsio/runtime', () => ({
	Browser: { OpenURL: vi.fn() },
}));

import { isBrowsableHttpUrl } from '@/lib/externalLinkDelegation';

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
