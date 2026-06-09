import { describe, expect, it } from 'vitest';
import {
	canonicalizeLinksJson,
	canonicalizeMarkdown,
	contentHashFromMarkdown,
} from './contentHash';

// Markdown/リンクの正規化と SHA-256 コンテンツハッシュを検証する。
describe('contentHash', () => {
	it('改行と前後空白を正規化する', () => {
		expect(canonicalizeMarkdown('  a\r\nb  ')).toBe('a\nb');
	});

	it('同一入力から安定した 64 文字の SHA-256 を返す', async () => {
		const h1 = await contentHashFromMarkdown('hello');
		const h2 = await contentHashFromMarkdown('hello');
		expect(h1).toBe(h2);
		expect(h1).toMatch(/^[a-f0-9]{64}$/);
	});

	it('リンク配列の順序差を正規化して同一 JSON にする', () => {
		expect(canonicalizeLinksJson(['b', 'a'])).toBe(
			canonicalizeLinksJson(['a', 'b']),
		);
	});
});
