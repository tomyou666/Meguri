const STORAGE_KEY = 'meguri.update.toastDismissedVersion';

export function getDismissedUpdateToastVersion(): string | null {
	try {
		return localStorage.getItem(STORAGE_KEY);
	} catch {
		return null;
	}
}

export function setDismissedUpdateToastVersion(version: string): void {
	try {
		localStorage.setItem(STORAGE_KEY, version);
	} catch {
		// ignore quota / private mode
	}
}

export function isUpdateToastDismissedForVersion(version: string): boolean {
	const dismissed = getDismissedUpdateToastVersion();
	return dismissed !== null && dismissed === version;
}
