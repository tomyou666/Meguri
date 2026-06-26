import { workspaceToDTO } from '@/lib/wailsMappers';
import type { Workspace } from '@/types/workspace';
import * as StoreService from '../../bindings/meguri-app/internal/usecase/wails_service/storeservice';

export type NodePositionPatch = {
	nodeId: string;
	position: { x: number; y: number };
	userPositioned: boolean;
};

let workspaceTimer: ReturnType<typeof setTimeout> | null = null;
let positionTimer: ReturnType<typeof setTimeout> | null = null;
const pendingPositionPatches = new Map<
	string,
	Map<string, NodePositionPatch>
>();

function cancelPendingPositionPatches(): void {
	if (positionTimer) {
		clearTimeout(positionTimer);
		positionTimer = null;
	}
	pendingPositionPatches.clear();
}

function flushPendingPositionPatches(): void {
	for (const [workspaceId, patches] of pendingPositionPatches) {
		const updates = [...patches.values()].map((p) => ({
			nodeId: p.nodeId,
			position: p.position,
			userPositioned: p.userPositioned,
		}));
		if (updates.length === 0) continue;
		void StoreService.PatchGraphNodePositions({
			workspaceId,
			updates,
		});
	}
	pendingPositionPatches.clear();
}

export function debouncedSaveWorkspace(ws: Workspace, delayMs = 500): void {
	cancelPendingPositionPatches();
	if (workspaceTimer) clearTimeout(workspaceTimer);
	workspaceTimer = setTimeout(() => {
		workspaceTimer = null;
		void StoreService.SaveWorkspace(workspaceToDTO(ws));
	}, delayMs);
}

export function debouncedPatchNodePositions(
	workspaceId: string,
	update: NodePositionPatch,
	delayMs = 500,
): void {
	let wsPatches = pendingPositionPatches.get(workspaceId);
	if (!wsPatches) {
		wsPatches = new Map();
		pendingPositionPatches.set(workspaceId, wsPatches);
	}
	wsPatches.set(update.nodeId, update);

	if (positionTimer) clearTimeout(positionTimer);
	positionTimer = setTimeout(() => {
		positionTimer = null;
		flushPendingPositionPatches();
	}, delayMs);
}

export function flushDebouncedWorkspaceSave(ws: Workspace): void {
	cancelPendingPositionPatches();
	if (workspaceTimer) {
		clearTimeout(workspaceTimer);
		workspaceTimer = null;
	}
	void StoreService.SaveWorkspace(workspaceToDTO(ws));
}
