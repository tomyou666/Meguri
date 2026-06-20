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
});
