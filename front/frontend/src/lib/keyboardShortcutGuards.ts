type TextInputLike = {
	tagName: string;
	isContentEditable: boolean;
};

type PrimaryModifierKeys = {
	ctrlKey: boolean;
	metaKey: boolean;
};

function isTextInputLike(target: EventTarget | null): boolean {
	return (
		target !== null &&
		typeof target === 'object' &&
		'tagName' in target &&
		'isContentEditable' in target
	);
}

/** Ctrl（Windows）または Cmd（Mac）が押されているか。 */
export function isPrimaryModifier(e: PrimaryModifierKeys): boolean {
	return e.ctrlKey || e.metaKey;
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
