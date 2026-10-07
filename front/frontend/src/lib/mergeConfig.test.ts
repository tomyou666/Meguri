import { describe, expect, it } from 'vitest';
import { mergeConfig } from './mergeConfig';

// 結果プレビュー用マージ。実行時のモード別マージとは別。
describe('mergeConfig', () => {
	it('ノード設定がワークスペース設定を上書きする', () => {
		const merged = mergeConfig(
			{},
			{ crawl: { max_depth: 5 } },
			{
				crawl: { max_depth: 1 },
			},
		);
		expect(merged.crawl?.max_depth).toBe(1);
	});
});
