import { describe, expect, it } from 'vitest';
import defaultsJson from '../../../shared/defaults.json';
import { DEFAULT_APP_CONFIG } from './defaults';

// shared defaults.json と TS 側 DEFAULT_APP_CONFIG の一致を検証する。
describe('defaults', () => {
	it('DEFAULT_APP_CONFIG は shared defaults.json と同一内容', () => {
		expect(DEFAULT_APP_CONFIG).toEqual(defaultsJson);
		expect(DEFAULT_APP_CONFIG.crawl?.max_depth).toBe(2);
		expect(DEFAULT_APP_CONFIG.plugins?.stealth?.chromium?.headless).toBe(true);
		expect(DEFAULT_APP_CONFIG.plugins?.stealth?.chromium?.hide_automation).toBe(
			true,
		);
	});
});
