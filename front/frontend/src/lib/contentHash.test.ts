import { describe, expect, it } from 'vitest';
import {
	canonicalizeLinksJson,
	canonicalizeMarkdown,
	contentHashFromMarkdown,
} from './contentHash';

describe('contentHash', () => {
	it('canonicalizes markdown line endings and trim', () => {
		expect(canonicalizeMarkdown('  a\r\nb  ')).toBe('a\nb');
	});

	it('produces stable SHA-256 hex', async () => {
		const h1 = await contentHashFromMarkdown('hello');
		const h2 = await contentHashFromMarkdown('hello');
		expect(h1).toBe(h2);
		expect(h1).toMatch(/^[a-f0-9]{64}$/);
	});

	it('canonicalizes links for diff', () => {
		expect(canonicalizeLinksJson(['b', 'a'])).toBe(
			canonicalizeLinksJson(['a', 'b']),
		);
	});
});
