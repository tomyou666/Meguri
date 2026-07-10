import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import {
	loadRobotsCache,
	type RobotsInfo,
	saveRobotsCacheEntry,
} from './robotsCache';

function installLocalStorageMock(): void {
	const store = new Map<string, string>();
	const mock = {
		getItem: (key: string) => store.get(key) ?? null,
		setItem: (key: string, value: string) => {
			store.set(key, value);
		},
		removeItem: (key: string) => {
			store.delete(key);
		},
		clear: () => {
			store.clear();
		},
	};
	Object.defineProperty(globalThis, 'localStorage', {
		value: mock,
		configurable: true,
	});
}

const STORAGE_KEY = 'meguri.domainStatus.robotsCache';
const TTL_MS = 24 * 60 * 60 * 1000;

// robots キャッシュの保存・読み込み・TTL を検証する。
describe('robotsCache', () => {
	beforeEach(() => {
		installLocalStorageMock();
	});

	afterEach(() => {
		localStorage.clear();
	});

	it('保存した結果を即座に読み込める', () => {
		const info: RobotsInfo = {
			status: 'found',
			statusCode: 200,
			body: 'User-agent: *\nDisallow:',
		};
		const now = 1_700_000_000_000;

		saveRobotsCacheEntry('example.com', info, now);
		expect(loadRobotsCache(now)).toEqual({ 'example.com': info });
	});

	it('24時間以上経過したエントリは読み込まずストレージからも消える', () => {
		const info: RobotsInfo = { status: 'not_found', statusCode: 404 };
		const savedAt = 1_700_000_000_000;
		const expiredAt = savedAt + TTL_MS + 1;

		saveRobotsCacheEntry('example.com', info, savedAt);
		expect(loadRobotsCache(expiredAt)).toEqual({});
		expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
	});

	it('loading 状態は保存されない', () => {
		const now = 1_700_000_000_000;

		saveRobotsCacheEntry('example.com', { status: 'loading' }, now);
		expect(loadRobotsCache(now)).toEqual({});
		expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
	});

	it('複数host混在時、期限切れのhostだけを除いてblobを書き直す', () => {
		const expiredInfo: RobotsInfo = { status: 'error', error: 'timeout' };
		const validInfo: RobotsInfo = { status: 'found', statusCode: 200 };
		const expiredSavedAt = 1_700_000_000_000;
		const validSavedAt = expiredSavedAt + 1_000;
		const now = expiredSavedAt + TTL_MS + 1;

		saveRobotsCacheEntry('expired.example', expiredInfo, expiredSavedAt);
		saveRobotsCacheEntry('valid.example', validInfo, validSavedAt);

		expect(loadRobotsCache(now)).toEqual({ 'valid.example': validInfo });

		const rewritten = JSON.parse(
			localStorage.getItem(STORAGE_KEY) ?? '{}',
		) as Record<string, unknown>;
		expect(Object.keys(rewritten)).toEqual(['valid.example']);

		// 書き直された blob からも、期限切れの host を除いて正しく再読み込みできる
		expect(loadRobotsCache(now)).toEqual({ 'valid.example': validInfo });
	});
});
