const STORAGE_KEY = 'meguri.graph.minimapCollapsed';

export function getMinimapCollapsed(): boolean {
	try {
		return localStorage.getItem(STORAGE_KEY) === 'true';
	} catch {
		return false;
	}
}

export function setMinimapCollapsed(collapsed: boolean): void {
	try {
		localStorage.setItem(STORAGE_KEY, collapsed ? 'true' : 'false');
	} catch {
		// ignore quota / private mode
	}
}
