import { useEffect } from 'react';
import { flushAllDebouncedSaves } from '@/lib/debouncedWorkspaceSave';
import { useAppStore } from '@/stores/appStore';

/** アプリ終了前に debounce 中のワークスペース保存をフラッシュする。 */
export function useFlushPendingSavesOnUnload() {
	useEffect(() => {
		const flush = () => {
			const ws = useAppStore.getState().getActiveWorkspace();
			if (ws) flushAllDebouncedSaves(ws);
		};
		window.addEventListener('pagehide', flush);
		window.addEventListener('beforeunload', flush);
		return () => {
			window.removeEventListener('pagehide', flush);
			window.removeEventListener('beforeunload', flush);
		};
	}, []);
}
