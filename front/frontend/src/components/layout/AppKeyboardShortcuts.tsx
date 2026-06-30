import { useEffect } from 'react';
import {
	isGraphShortcutTarget,
	isTextInputElement,
	shouldDeferToNativeTextEditing,
} from '@/lib/keyboardShortcutGuards';
import { useAppStore } from '@/stores/appStore';

export function AppKeyboardShortcuts() {
	useEffect(() => {
		const onKeyDown = (e: KeyboardEvent) => {
			const store = useAppStore.getState();
			const target = e.target;

			if (e.ctrlKey && e.key === 'z' && !e.shiftKey) {
				if (isTextInputElement(target)) return;
				e.preventDefault();
				store.undo();
				return;
			}
			if (e.ctrlKey && (e.key === 'y' || (e.key === 'z' && e.shiftKey))) {
				if (isTextInputElement(target)) return;
				e.preventDefault();
				store.redo();
				return;
			}
			if (e.ctrlKey && e.key === 'c') {
				if (shouldDeferToNativeTextEditing(target)) return;
				e.preventDefault();
				store.copySelectedNodes();
				return;
			}
			if (e.ctrlKey && e.key === 'v') {
				if (shouldDeferToNativeTextEditing(target)) return;
				e.preventDefault();
				store.pasteNodes();
				return;
			}
			if (e.ctrlKey && e.key === 'a') {
				if (shouldDeferToNativeTextEditing(target)) return;
				if (!isGraphShortcutTarget(target)) return;
				e.preventDefault();
				store.selectAllNodes();
				return;
			}
			if (e.key === 'Delete' && store.selectedNodeIds.length > 0) {
				if (isTextInputElement(target)) return;
				store.deleteSelectedNodes();
			}
		};
		window.addEventListener('keydown', onKeyDown);
		return () => window.removeEventListener('keydown', onKeyDown);
	}, []);
	return null;
}
