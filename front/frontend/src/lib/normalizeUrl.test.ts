import { describe, expect, it } from 'vitest';
import { normalizeUrl } from './normalizeUrl';

// クロール用 URL の正規化ルールを検証する。
describe('normalizeUrl', () => {
	it('ホストを小文字化しフラグメントを除去する', () => {
		expect(normalizeUrl('HTTPS://Example.COM/path#frag')).toBe(
			'https://example.com/path',
		);
	});

	it('クエリキーをソートする', () => {
		expect(normalizeUrl('https://ex.com/?b=2&a=1')).toBe(
			'https://ex.com/?a=1&b=2',
		);
	});

	it('デフォルト HTTPS ポートを除去する', () => {
		expect(normalizeUrl('https://ex.com:443/')).toBe('https://ex.com/');
	});
});
