import type { CrawlResultPreview } from '@/types/crawl';

/** アプリ全体の本文キャッシュ合計上限（文字数概算）。 */
export const NODE_RESULT_BODY_CACHE_MAX_BYTES = 32 * 1024 * 1024;

/** サイズ超過時もこれ未満なら追い出さない件数。 */
export const NODE_RESULT_BODY_CACHE_MIN_ENTRIES = 10;

type CacheEntry = {
	workspaceId: string;
	nodeId: string;
	preview: CrawlResultPreview;
	size: number;
};

export type NodeResultBodyCache = {
	get: (workspaceId: string, nodeId: string) => CrawlResultPreview | null;
	set: (
		workspaceId: string,
		nodeId: string,
		preview: CrawlResultPreview,
	) => void;
	drop: (workspaceId: string, nodeId: string) => void;
	dropWorkspace: (workspaceId: string) => void;
	/** テスト用: 現在の合計サイズ（文字数概算）。 */
	totalSize: () => number;
	/** テスト用: エントリ数。 */
	size: () => number;
};

function cacheKey(workspaceId: string, nodeId: string): string {
	return `${workspaceId}\0${nodeId}`;
}

/** CrawlResultPreview の本文相当フィールドの文字数合計。 */
export function estimatePreviewSize(preview: CrawlResultPreview): number {
	let n = 0;
	n += preview.markdown?.length ?? 0;
	n += preview.html?.length ?? 0;
	n += preview.raw_html?.length ?? 0;
	n += preview.json?.length ?? 0;
	if (preview.links) {
		for (const link of preview.links) {
			n += link.length;
		}
	}
	if (preview.metadata) {
		for (const [k, v] of Object.entries(preview.metadata)) {
			n += k.length + v.length;
		}
	}
	return n;
}

/**
 * ノード本文のセッション内 LRU キャッシュを作る。
 *
 * maxBytes を超えても minEntries 未満なら保持する。
 * minEntries 以上で超えるときは LRU で空きを作り、それでも入らなければ set しない
 *（既存エントリは消さない）。
 */
export function createNodeResultBodyCache(
	maxBytes = NODE_RESULT_BODY_CACHE_MAX_BYTES,
	minEntries = NODE_RESULT_BODY_CACHE_MIN_ENTRIES,
): NodeResultBodyCache {
	/** Map の挿入順を LRU に使う（先頭が最古）。 */
	const map = new Map<string, CacheEntry>();
	let total = 0;

	const touch = (key: string, entry: CacheEntry) => {
		map.delete(key);
		map.set(key, entry);
	};

	const removeKey = (key: string) => {
		const entry = map.get(key);
		if (!entry) return;
		map.delete(key);
		total -= entry.size;
	};

	return {
		get(workspaceId, nodeId) {
			const key = cacheKey(workspaceId, nodeId);
			const entry = map.get(key);
			if (!entry) return null;
			touch(key, entry);
			return entry.preview;
		},

		set(workspaceId, nodeId, preview) {
			const key = cacheKey(workspaceId, nodeId);
			const size = estimatePreviewSize(preview);
			const previous = map.get(key);

			// 上書き対象を除いた状態で予算をシミュレーションする（拒否時は previous を残す）。
			let simulatedTotal = total - (previous?.size ?? 0);
			let simulatedCount = map.size - (previous ? 1 : 0);

			// 件数優先: まだ最低件数に満たないなら予算超過でも入れる。
			if (simulatedCount < minEntries) {
				if (previous) {
					removeKey(key);
				}
				map.set(key, { workspaceId, nodeId, preview, size });
				total += size;
				return;
			}

			// 入るまで LRU をシミュレーション（最終件数が minEntries を下回らない範囲）。
			// previous は既に「無いもの」として数えているので、追い出し候補から除外する。
			const evictKeys: string[] = [];
			for (const oldest of map.keys()) {
				if (oldest === key) continue;
				if (simulatedTotal + size <= maxBytes) break;
				if (simulatedCount < minEntries) break;
				const entry = map.get(oldest);
				if (!entry) break;
				evictKeys.push(oldest);
				simulatedTotal -= entry.size;
				simulatedCount -= 1;
			}

			if (simulatedTotal + size > maxBytes) {
				// 空きが足りないので新しい方はキャッシュしない（既存・previous は触らない）。
				return;
			}

			if (previous) {
				removeKey(key);
			}
			for (const k of evictKeys) {
				removeKey(k);
			}
			map.set(key, { workspaceId, nodeId, preview, size });
			total += size;
		},

		drop(workspaceId, nodeId) {
			removeKey(cacheKey(workspaceId, nodeId));
		},

		dropWorkspace(workspaceId) {
			for (const [key, entry] of [...map.entries()]) {
				if (entry.workspaceId === workspaceId) {
					removeKey(key);
				}
			}
		},

		totalSize: () => total,
		size: () => map.size,
	};
}

/** アプリ全体で共有する本文キャッシュ。 */
export const nodeResultBodyCache = createNodeResultBodyCache();
