import { describe, expect, it } from 'vitest';
import {
	hasNonEmptyTextSelection,
	isPrimaryModifier,
	isTextInputElement,
	shouldDeferToNativeTextEditing,
} from '@/lib/keyboardShortcutGuards';

function el(
	tagName: string,
	opts?: { contentEditable?: boolean },
): HTMLElement {
	return {
		tagName: tagName.toUpperCase(),
		isContentEditable: opts?.contentEditable ?? false,
	} as unknown as HTMLElement;
}

describe('keyboardShortcutGuards', () => {
	it('INPUT / TEXTAREA / contenteditable をテキスト入力として判定する', () => {
		expect(isTextInputElement(el('input'))).toBe(true);
		expect(isTextInputElement(el('textarea'))).toBe(true);
		expect(isTextInputElement(el('div', { contentEditable: true }))).toBe(true);
		expect(isTextInputElement(el('div'))).toBe(false);
		expect(isTextInputElement(null)).toBe(false);
	});

	it('折りたたまれた選択・空文字はテキスト選択なしとする', () => {
		expect(
			hasNonEmptyTextSelection({
				isCollapsed: true,
				toString: () => 'hello',
			} as Selection),
		).toBe(false);
		expect(
			hasNonEmptyTextSelection({
				isCollapsed: false,
				toString: () => '',
			} as Selection),
		).toBe(false);
		expect(
			hasNonEmptyTextSelection({
				isCollapsed: false,
				toString: () => 'selected',
			} as Selection),
		).toBe(true);
	});

	it('非空のテキスト選択または入力欄ではネイティブ編集を優先する', () => {
		const selection = {
			isCollapsed: false,
			toString: () => 'selected text',
		} as Selection;

		expect(shouldDeferToNativeTextEditing(el('div'), selection)).toBe(true);
		expect(shouldDeferToNativeTextEditing(el('input'), null)).toBe(true);
		expect(shouldDeferToNativeTextEditing(el('div'), null)).toBe(false);
	});

	it('ctrlKey または metaKey を主修飾キーとして判定する', () => {
		expect(isPrimaryModifier({ ctrlKey: true, metaKey: false })).toBe(true);
		expect(isPrimaryModifier({ ctrlKey: false, metaKey: true })).toBe(true);
		expect(isPrimaryModifier({ ctrlKey: true, metaKey: true })).toBe(true);
		expect(isPrimaryModifier({ ctrlKey: false, metaKey: false })).toBe(false);
	});
});
