import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import {
	getRescrapeExisting,
	getRunMode,
	setRescrapeExisting,
	setRunMode,
} from './controlBarPreferences';

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

// localStorage.getItem/setItem が例外を投げる状態を再現し、フォールバック分岐を検証するためのモック
function installThrowingLocalStorageMock(): void {
	const mock = {
		getItem: () => {
			throw new Error('blocked');
		},
		setItem: () => {
			throw new Error('blocked');
		},
		removeItem: () => {},
		clear: () => {},
	};
	Object.defineProperty(globalThis, 'localStorage', {
		value: mock,
		configurable: true,
	});
}

describe('controlBarPreferences', () => {
	beforeEach(() => {
		installLocalStorageMock();
	});

	afterEach(() => {
		localStorage.clear();
	});

	describe('runMode', () => {
		it('保存した値を読み込める', () => {
			setRunMode(3);
			expect(getRunMode(1)).toBe(3);
		});

		it('未保存の場合はフォールバック値を返す', () => {
			expect(getRunMode(2)).toBe(2);
		});

		it('不正な値が保存されている場合はフォールバック値を返す', () => {
			localStorage.setItem('meguri.control.runMode', '99');
			expect(getRunMode(1)).toBe(1);
		});

		it('localStorage アクセスが例外を投げる場合はフォールバック値を返す', () => {
			installThrowingLocalStorageMock();
			expect(getRunMode(4)).toBe(4);
			expect(() => setRunMode(2)).not.toThrow();
		});
	});

	describe('rescrapeExisting', () => {
		it('保存した値を読み込める', () => {
			setRescrapeExisting(true);
			expect(getRescrapeExisting(false)).toBe(true);
		});

		it('未保存の場合はフォールバック値を返す', () => {
			expect(getRescrapeExisting(true)).toBe(true);
		});

		it('localStorage アクセスが例外を投げる場合はフォールバック値を返す', () => {
			installThrowingLocalStorageMock();
			expect(getRescrapeExisting(true)).toBe(true);
			expect(() => setRescrapeExisting(false)).not.toThrow();
		});
	});
});
