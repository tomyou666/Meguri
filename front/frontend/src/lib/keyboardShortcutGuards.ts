type TextInputLike = {
	tagName: string;
	isContentEditable: boolean;
};

type ElementLike = {
	closest: (selector: string) => Element | null;
};

function isTextInputLike(target: EventTarget | null): boolean {
	return (
		target !== null &&
		typeof target === 'object' &&
		'tagName' in target &&
		'isContentEditable' in target
	);
}

function isElementLike(target: EventTarget | null): boolean {
	return (
		target !== null &&
		typeof target === 'object' &&
		'closest' in target &&
		typeof (target as ElementLike).closest === 'function'
	);
}

/** INPUT / TEXTAREA / contenteditable ではブラウザ標準の編集ショートカットを優先する。 */
export function isTextInputElement(target: EventTarget | null): boolean {
	if (!isTextInputLike(target)) {
		return false;
	}
	const el = target as unknown as TextInputLike;
	const tag = el.tagName;
	if (tag === 'INPUT' || tag === 'TEXTAREA') {
		return true;
	}
	return el.isContentEditable;
}

/** 画面上でテキストが選択されている。 */
export function hasNonEmptyTextSelection(
	selection: Selection | null = typeof window !== 'undefined'
		? window.getSelection()
		: null,
): boolean {
	if (!selection || selection.isCollapsed) {
		return false;
	}
	return selection.toString().length > 0;
}

/** Ctrl+C / Ctrl+V などでテキスト操作を優先すべきか。 */
export function shouldDeferToNativeTextEditing(
	target: EventTarget | null,
	selection: Selection | null = typeof window !== 'undefined'
		? window.getSelection()
		: null,
): boolean {
	return isTextInputElement(target) || hasNonEmptyTextSelection(selection);
}

/** グラフ上でのノード向けショートカット（Ctrl+A 等）の対象か。 */
export function isGraphShortcutTarget(target: EventTarget | null): boolean {
	if (!isElementLike(target)) {
		return false;
	}
	return (target as unknown as ElementLike).closest('.react-flow') !== null;
}
