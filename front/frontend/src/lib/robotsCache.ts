export type RobotsStatus = 'loading' | 'found' | 'not_found' | 'error';
export type RobotsInfo = {
	status: RobotsStatus;
	statusCode?: number;
	body?: string;
	error?: string;
};

const STORAGE_KEY = 'meguri.domainStatus.robotsCache';
const TTL_MS = 24 * 60 * 60 * 1000;

type PersistedEntry = { info: RobotsInfo; savedAt: number };
type PersistedCache = Record<string, PersistedEntry>;

function readPersistedCache(): PersistedCache {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (!raw) return {};
		const parsed = JSON.parse(raw) as unknown;
		if (!parsed || typeof parsed !== 'object') return {};
		return parsed as PersistedCache;
	} catch {
		return {};
	}
}

function writePersistedCache(cache: PersistedCache): void {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(cache));
	} catch {
		// ignore quota / private mode
	}
}

function isValidPersistedEntry(
	entry: unknown,
	now: number,
): entry is PersistedEntry {
	if (!entry || typeof entry !== 'object') return false;
	const { info, savedAt } = entry as PersistedEntry;
	if (typeof savedAt !== 'number' || now - savedAt > TTL_MS) return false;
	if (!info || typeof info !== 'object') return false;
	const status = (info as RobotsInfo).status;
	return status === 'found' || status === 'not_found' || status === 'error';
}

/** localStorage から TTL 内の robots キャッシュを読み込む。期限切れは掃除する。 */
export function loadRobotsCache(now = Date.now()): Record<string, RobotsInfo> {
	const persisted = readPersistedCache();
	const result: Record<string, RobotsInfo> = {};
	let changed = false;

	for (const [host, entry] of Object.entries(persisted)) {
		if (isValidPersistedEntry(entry, now)) {
			result[host] = entry.info;
		} else {
			changed = true;
		}
	}

	if (changed) {
		const cleaned: PersistedCache = {};
		for (const [host, info] of Object.entries(result)) {
			const savedAt = persisted[host]?.savedAt;
			if (typeof savedAt === 'number') {
				cleaned[host] = { info, savedAt };
			}
		}
		if (Object.keys(cleaned).length === 0) {
			try {
				localStorage.removeItem(STORAGE_KEY);
			} catch {
				// ignore
			}
		} else {
			writePersistedCache(cleaned);
		}
	}

	return result;
}

/** 確定した robots 結果を localStorage に保存する。loading は無視する。 */
export function saveRobotsCacheEntry(
	host: string,
	info: RobotsInfo,
	now = Date.now(),
): void {
	if (info.status === 'loading') return;

	const persisted = readPersistedCache();
	persisted[host] = { info, savedAt: now };
	writePersistedCache(persisted);
}
