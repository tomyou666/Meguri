import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import {
	getDismissedUpdateToastVersion,
	isUpdateToastDismissedForVersion,
	setDismissedUpdateToastVersion,
} from './updatePreferences';

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

describe('updatePreferences', () => {
	beforeEach(() => {
		installLocalStorageMock();
	});

	afterEach(() => {
		localStorage.clear();
	});

	it('stores dismissed version', () => {
		setDismissedUpdateToastVersion('1.2.3');
		expect(getDismissedUpdateToastVersion()).toBe('1.2.3');
		expect(isUpdateToastDismissedForVersion('1.2.3')).toBe(true);
		expect(isUpdateToastDismissedForVersion('1.2.4')).toBe(false);
	});
});
