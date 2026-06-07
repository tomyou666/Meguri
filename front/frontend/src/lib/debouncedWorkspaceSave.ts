import { workspaceToDTO } from '@/lib/wailsMappers';
import type { Workspace } from '@/types/workspace';
import * as StoreService from '../../bindings/scraperbot-front/internal/usecase/wails_service/storeservice';

let timer: ReturnType<typeof setTimeout> | null = null;

export function debouncedSaveWorkspace(ws: Workspace, delayMs = 500): void {
	if (timer) clearTimeout(timer);
	timer = setTimeout(() => {
		timer = null;
		void StoreService.SaveWorkspace(workspaceToDTO(ws));
	}, delayMs);
}

export function flushDebouncedWorkspaceSave(ws: Workspace): void {
	if (timer) {
		clearTimeout(timer);
		timer = null;
	}
	void StoreService.SaveWorkspace(workspaceToDTO(ws));
}
