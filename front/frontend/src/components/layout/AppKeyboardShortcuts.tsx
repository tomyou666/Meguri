import { useEffect } from 'react';
import {
	isPrimaryModifier,
	isTextInputElement,
	shouldDeferToNativeTextEditing,
} from '@/lib/keyboardShortcutGuards';
import { useAppStore } from '@/stores/appStore';

export function AppKeyboardShortcuts() {
	useEffect(() => {
		const onKeyDown = (e: KeyboardEvent) => {
			const store = useAppStore.getState();
			const target = e.target;
			const mod = isPrimaryModifier(e);

			if (mod && e.key === 'z' && !e.shiftKey) {
				if (isTextInputElement(target)) return;
				e.preventDefault();
				store.undo();
				return;
			}
			if (mod && (e.key === 'y' || (e.key === 'z' && e.shiftKey))) {
				if (isTextInputElement(target)) return;
				e.preventDefault();
				store.redo();
				return;
			}
			if (mod && e.key === 'c') {
				if (shouldDeferToNativeTextEditing(target)) return;
				e.preventDefault();
				store.copySelectedNodes();
				return;
			}
			if (mod && e.key === 'v') {
				if (shouldDeferToNativeTextEditing(target)) return;
				e.preventDefault();
				store.pasteNodes();
				return;
			}
			if (mod && e.key === 'a') {
				if (shouldDeferToNativeTextEditing(target)) return;
				e.preventDefault();
				store.selectAllNodes();
				return;
			}
			if (e.key === 'Delete' && store.selectedNodeIds.length > 0) {
				if (isTextInputElement(target)) return;
				store.deleteSelectedNodes();
				return;
			}
			// ツール切替は入力中のみ除外。React Flow クリック後も focus が BODY のまま。
			if ((e.key === 'h' || e.key === 'H') && !mod && !e.altKey) {
				if (isTextInputElement(target)) return;
				e.preventDefault();
				store.setGraphToolMode('pan');
				return;
			}
			if ((e.key === 'v' || e.key === 'V') && !mod && !e.altKey) {
				if (isTextInputElement(target)) return;
				e.preventDefault();
				store.setGraphToolMode('select');
			}
		};
		window.addEventListener('keydown', onKeyDown);
		return () => window.removeEventListener('keydown', onKeyDown);
	}, []);
	return null;
}
