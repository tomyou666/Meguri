import { useEffect } from 'react';
import { useAppStore } from '@/stores/appStore';

export function AppKeyboardShortcuts() {
	useEffect(() => {
		const onKeyDown = (e: KeyboardEvent) => {
			const target = e.target as HTMLElement;
			if (
				target.tagName === 'INPUT' ||
				target.tagName === 'TEXTAREA' ||
				target.isContentEditable
			) {
				return;
			}
			const store = useAppStore.getState();
			if (e.ctrlKey && e.key === 'z' && !e.shiftKey) {
				e.preventDefault();
				store.undo();
			}
			if (e.ctrlKey && (e.key === 'y' || (e.key === 'z' && e.shiftKey))) {
				e.preventDefault();
				store.redo();
			}
			if (e.ctrlKey && e.key === 'c') {
				e.preventDefault();
				store.copySelectedNodes();
			}
			if (e.ctrlKey && e.key === 'v') {
				e.preventDefault();
				store.pasteNodes();
			}
			if (e.ctrlKey && e.key === 'a') {
				e.preventDefault();
				store.selectAllNodes();
			}
			if (e.key === 'Delete' && store.selectedNodeIds.length > 0) {
				store.deleteSelectedNodes();
			}
		};
		window.addEventListener('keydown', onKeyDown);
		return () => window.removeEventListener('keydown', onKeyDown);
	}, []);
	return null;
}
