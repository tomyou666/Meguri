import { temporal } from 'zundo';
import { create } from 'zustand';
import type { Workspace } from '@/types/workspace';

export interface GraphHistorySlice {
	workspaces: Workspace[];
	activeWorkspaceId: string | null;
}

export const useGraphHistoryStore = create(
	temporal<GraphHistorySlice>(
		() => ({
			workspaces: [],
			activeWorkspaceId: null,
		}),
		{
			partialize: (state) => ({
				workspaces: state.workspaces,
				activeWorkspaceId: state.activeWorkspaceId,
			}),
			limit: 50,
		},
	),
);

export function syncGraphHistory(
	workspaces: Workspace[],
	activeWorkspaceId: string | null,
) {
	useGraphHistoryStore.setState({ workspaces, activeWorkspaceId });
}

export function undoGraph() {
	useGraphHistoryStore.temporal.getState().undo();
	const { workspaces, activeWorkspaceId } = useGraphHistoryStore.getState();
	return { workspaces, activeWorkspaceId };
}

export function redoGraph() {
	useGraphHistoryStore.temporal.getState().redo();
	const { workspaces, activeWorkspaceId } = useGraphHistoryStore.getState();
	return { workspaces, activeWorkspaceId };
}
