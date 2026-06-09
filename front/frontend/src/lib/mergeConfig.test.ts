import { describe, expect, it } from 'vitest';
import { DEFAULT_APP_CONFIG } from './defaults';
import { configForMode2, mergeConfig } from './mergeConfig';

// 設定レイヤーのマージ優先順位とモード2のデフォルト適用を検証する。
describe('mergeConfig', () => {
	it('ノード設定がワークスペース設定を上書きする', () => {
		const merged = mergeConfig({}, { crawl: { max_depth: 5 } }, undefined, {
			crawl: { max_depth: 1 },
		});
		expect(merged.crawl?.max_depth).toBe(1);
	});

	it('モード2はアプリデフォルトのみをベースに部分上書きする', () => {
		const m2 = configForMode2({ crawl: { max_depth: 9 } });
		expect(m2.crawl?.max_depth).toBe(9);
		expect(m2.crawl?.max_pages).toBe(DEFAULT_APP_CONFIG.crawl?.max_pages);
	});
});
