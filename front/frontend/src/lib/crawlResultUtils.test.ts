import { describe, expect, it } from 'vitest';
import { isPdfResourceResult } from './crawlResultUtils';

// metadata.content_type から PDF リソースか判定する。
describe('isPdfResourceResult', () => {
	it('application/pdf のとき true', () => {
		expect(
			isPdfResourceResult({
				url: 'https://example.com/doc.pdf',
				metadata: { content_type: 'application/pdf' },
			}),
		).toBe(true);
	});

	it('application/pdf; charset=binary のとき true', () => {
		expect(
			isPdfResourceResult({
				url: 'https://example.com/doc.pdf',
				metadata: { content_type: 'application/pdf; charset=binary' },
			}),
		).toBe(true);
	});

	it('text/html のとき false', () => {
		expect(
			isPdfResourceResult({
				url: 'https://example.com/',
				metadata: { content_type: 'text/html; charset=utf-8' },
			}),
		).toBe(false);
	});

	it('metadata なし・結果なしのとき false', () => {
		expect(isPdfResourceResult({ url: 'https://example.com/' })).toBe(false);
		expect(isPdfResourceResult(undefined)).toBe(false);
		expect(isPdfResourceResult(null)).toBe(false);
	});
});
