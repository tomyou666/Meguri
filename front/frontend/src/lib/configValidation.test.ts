import { describe, expect, it } from 'vitest';
import {
	getConfigFieldErrors,
	validatePartialConfig,
} from './configValidation';

describe('validatePartialConfig', () => {
	it('accepts empty partial config', () => {
		const r = validatePartialConfig({});
		expect(r.ok).toBe(true);
	});

	it('rejects invalid retry_count with Japanese message', () => {
		const r = validatePartialConfig({
			request: { retry_count: 99 },
		});
		expect(r.ok).toBe(false);
		if (r.ok === false) {
			expect(r.errors[0]).toContain('再試行回数');
			expect(r.errors[0]).toContain('10以下');
		}
	});

	it('maps field errors by path', () => {
		const errors = getConfigFieldErrors({ request: { retry_count: 99 } });
		expect(errors['request.retry_count']).toContain('10以下');
	});

	it('requires at least one content format when formats set', () => {
		const r = validatePartialConfig({
			content: { formats: [] },
		});
		expect(r.ok).toBe(false);
	});
});
