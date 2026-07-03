import { describe, expect, it } from 'vitest';
import {
	addToken,
	normalizeToken,
	removeLastToken,
	removeTokenAt,
} from '@/components/settings/tagListInputUtils';

// TagListInput の Enter/カンマ追加・削除ロジックを検証する。
describe('tagListInputUtils', () => {
	it('normalizeToken: trim 後に空なら null', () => {
		expect(normalizeToken('  article  ')).toBe('article');
		expect(normalizeToken('   ')).toBeNull();
		expect(normalizeToken('')).toBeNull();
	});

	it('addToken: Enter 相当で末尾に追加する', () => {
		expect(addToken([], 'article')).toEqual(['article']);
		expect(addToken(['article'], 'section')).toEqual(['article', 'section']);
	});

	it('addToken: カンマ確定相当でも trim して追加する', () => {
		expect(addToken([], ' /docs ')).toEqual(['/docs']);
	});

	it('addToken: 重複は黙ってスキップする', () => {
		const values = ['article'];
		expect(addToken(values, 'article')).toBe(values);
		expect(addToken(values, '  article  ')).toBe(values);
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
