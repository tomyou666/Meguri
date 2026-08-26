import { describe, expect, it } from 'vitest';
import {
	addToken,
	addTokens,
	normalizeToken,
	removeLastToken,
	removeTokenAt,
	splitTokens,
	tokensToCopyText,
} from '@/components/settings/tagListInputUtils';

// TagListInput の Enter 追加・改行一括貼り付け・削除ロジックを検証する。
describe('tagListInputUtils', () => {
	it('normalizeToken: trim 後に空なら null', () => {
		expect(normalizeToken('  article  ')).toBe('article');
		expect(normalizeToken('   ')).toBeNull();
		expect(normalizeToken('')).toBeNull();
	});

	it('splitTokens: 改行区切りで分割し空行を捨てる', () => {
		expect(splitTokens('xlsx$\n\ndoc$')).toEqual(['xlsx$', 'doc$']);
	});

	it('splitTokens: カンマはトークンの一部として残す', () => {
		expect(splitTokens('xlsx$,doc$')).toEqual(['xlsx$,doc$']);
		expect(splitTokens('div:has(span, p)')).toEqual(['div:has(span, p)']);
	});

	it('splitTokens: 前後空白を trim する', () => {
		expect(splitTokens(' xlsx$ \n doc$ ')).toEqual(['xlsx$', 'doc$']);
	});

	it('splitTokens: 空入力は空配列', () => {
		expect(splitTokens('')).toEqual([]);
		expect(splitTokens('  \n  ')).toEqual([]);
	});

	it('addToken: Enter 相当で末尾に追加する', () => {
		expect(addToken([], 'article')).toEqual(['article']);
		expect(addToken(['article'], 'section')).toEqual(['article', 'section']);
	});

	it('addToken: trim して追加する', () => {
		expect(addToken([], ' /docs ')).toEqual(['/docs']);
	});

	it('addToken: 重複は黙ってスキップする', () => {
		const values = ['article'];
		expect(addToken(values, 'article')).toBe(values);
		expect(addToken(values, '  article  ')).toBe(values);
	});

	it('addTokens: 複数行を追加する', () => {
		expect(addTokens([], 'xlsx$\n\ndoc$')).toEqual(['xlsx$', 'doc$']);
	});

	it('addTokens: カンマは割らず1件として追加する', () => {
		expect(addTokens([], 'xlsx$,doc$')).toEqual(['xlsx$,doc$']);
	});

	it('addTokens: 重複は黙ってスキップする', () => {
		expect(addTokens(['xlsx$'], 'xlsx$\ndoc$')).toEqual(['xlsx$', 'doc$']);
	});

	it('tokensToCopyText: 1行1件の改行区切り', () => {
		expect(tokensToCopyText(['xlsx$', 'doc$'])).toBe('xlsx$\ndoc$');
		expect(tokensToCopyText([])).toBe('');
	});

	it('removeLastToken: 入力空の Backspace 相当で末尾を削除する', () => {
		expect(removeLastToken(['a', 'b'])).toEqual(['a']);
		expect(removeLastToken([])).toEqual([]);
	});

	it('removeTokenAt: × クリック相当で指定インデックスを削除する', () => {
		expect(removeTokenAt(['a', 'b', 'c'], 1)).toEqual(['a', 'c']);
		expect(removeTokenAt(['a'], -1)).toEqual(['a']);
	});
});
