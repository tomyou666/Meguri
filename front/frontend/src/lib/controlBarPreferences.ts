import type { RunMode } from '@/types/crawl';

const RUN_MODE_KEY = 'meguri.control.runMode';
const RESCRAPE_EXISTING_KEY = 'meguri.control.rescrapeExisting';

function isValidRunMode(value: number): value is RunMode {
	return value === 1 || value === 2 || value === 3 || value === 4;
}

export function getRunMode(fallback: RunMode): RunMode {
	try {
		const stored = localStorage.getItem(RUN_MODE_KEY);
		if (stored === null) return fallback;
		const parsed = Number(stored);
		return isValidRunMode(parsed) ? parsed : fallback;
	} catch {
		return fallback;
	}
}

export function setRunMode(mode: RunMode): void {
	try {
		localStorage.setItem(RUN_MODE_KEY, String(mode));
	} catch {
		// ignore quota / private mode
	}
}

export function getRescrapeExisting(fallback: boolean): boolean {
	try {
		const stored = localStorage.getItem(RESCRAPE_EXISTING_KEY);
		if (stored === null) return fallback;
		return stored === 'true';
	} catch {
		return fallback;
	}
}

export function setRescrapeExisting(value: boolean): void {
	try {
		localStorage.setItem(RESCRAPE_EXISTING_KEY, value ? 'true' : 'false');
	} catch {
		// ignore quota / private mode
	}
}
