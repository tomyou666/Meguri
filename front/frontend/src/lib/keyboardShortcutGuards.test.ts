import { describe, expect, it } from 'vitest';
import {
	hasNonEmptyTextSelection,
	isGraphShortcutTarget,
	isTextInputElement,
	shouldDeferToNativeTextEditing,
} from '@/lib/keyboardShortcutGuards';

function el(
	tagName: string,
	opts?: { contentEditable?: boolean; className?: string },
): HTMLElement {
	return {
		tagName: tagName.toUpperCase(),
		isContentEditable: opts?.contentEditable ?? false,
		className: opts?.className ?? '',
		closest: (selector: string) => {
			if (selector === '.react-flow' && opts?.className === 'react-flow') {
				return {} as Element;
			}
			return null;
		},
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

	it('.react-flow 配下のみグラフ向けショートカット対象とする', () => {
		const graph = el('div', { className: 'react-flow' });
		const pane = {
			closest: (selector: string) =>
				selector === '.react-flow' ? graph : null,
		} as unknown as Element;

		expect(isGraphShortcutTarget(pane)).toBe(true);
		expect(isGraphShortcutTarget(el('div'))).toBe(false);
		expect(isGraphShortcutTarget(null)).toBe(false);
	});
});
