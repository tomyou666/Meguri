import { useOnSelectionChange } from '@xyflow/react';
import { useAppStore } from '@/stores/appStore';

/** ReactFlow 内部でのみマウントすること（Provider 必須） */
export function GraphSelectionSync() {
	const selectNodes = useAppStore((s) => s.selectNodes);
	const clearNodeSelection = useAppStore((s) => s.clearNodeSelection);

	useOnSelectionChange({
		onChange: ({ nodes }) => {
			if (useAppStore.getState()._suppressSelectionSync) return;
			const ids = nodes.map((n) => n.id);
			if (ids.length === 0) {
				clearNodeSelection();
				return;
			}
			selectNodes(ids);
		},
	});

	return null;
}
