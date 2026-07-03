import { describe, expect, it } from 'vitest';
import {
	getConfigFieldErrors,
	validatePartialConfig,
} from './configValidation';

// 部分設定のバリデーションとフィールド別エラーマップを検証する。
describe('validatePartialConfig', () => {
	it('空の部分設定は受理する', () => {
		const r = validatePartialConfig({});
		expect(r.ok).toBe(true);
	});

	it('範囲外の retry_count を日本語メッセージで拒否する', () => {
		const r = validatePartialConfig({
			request: { retry_count: 99 },
		});
		expect(r.ok).toBe(false);
		if (r.ok === false) {
			expect(r.errors[0]).toContain('再試行回数');
			expect(r.errors[0]).toContain('10以下');
		}
	});

	it('パス単位でフィールドエラーを取得できる', () => {
		const errors = getConfigFieldErrors({ request: { retry_count: 99 } });
		expect(errors['request.retry_count']).toContain('10以下');
	});

	it('空の formats 配列は受理する', () => {
		const r = validatePartialConfig({
			content: { formats: [] },
		});
		expect(r.ok).toBe(true);
	});

	it('timeout がレンジ外ならフィールドエラーを返す', () => {
		const errors = getConfigFieldErrors({
			request: { timeout: '500ms' },
		});
		expect(errors['request.timeout']).toContain('1秒以上300秒以下');
	});

	it('retry_interval がレンジ外ならフィールドエラーを返す', () => {
		const errors = getConfigFieldErrors({
			request: { retry_interval: '50ms' },
		});
		expect(errors['request.retry_interval']).toContain('100ミリ秒以上60秒以下');
	});

	it('request_delay がレンジ外ならフィールドエラーを返す', () => {
		const errors = getConfigFieldErrors({
			crawl: { request_delay: '90s' },
		});
		expect(errors['crawl.request_delay']).toContain('0秒以上60秒以下');
	});
});
