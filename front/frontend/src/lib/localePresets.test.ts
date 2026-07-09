import { describe, expect, it } from 'vitest';
import {
	LOCALE_PRESET_CUSTOM,
	LOCALE_PRESETS,
	localePresetOptionValue,
	resolveLocaleSelectValue,
} from './localePresets';

// locale プリセットの option value 解決と厳密一致を検証する。
describe('localePresets', () => {
	it('LOCALE_PRESETS は 10 件で id が一意', () => {
		expect(LOCALE_PRESETS).toHaveLength(10);
		const ids = LOCALE_PRESETS.map((p) => p.id);
		expect(new Set(ids).size).toBe(ids.length);
	});

	it('localePresetOptionValue は field ごとに正しい値を返す', () => {
		const ja = LOCALE_PRESETS[0];
		expect(localePresetOptionValue(ja, 'lang')).toBe('ja-JP');
		expect(localePresetOptionValue(ja, 'accept_language')).toBe(
			'ja,en-US;q=0.9,en;q=0.8',
		);
	});

	it('resolveLocaleSelectValue: 空は未設定', () => {
		expect(resolveLocaleSelectValue(undefined, 'lang')).toBe('');
		expect(resolveLocaleSelectValue('', 'accept_language')).toBe('');
	});

	it('resolveLocaleSelectValue: 厳密一致でプリセット value', () => {
		expect(resolveLocaleSelectValue('ja-JP', 'lang')).toBe('ja-JP');
		expect(
			resolveLocaleSelectValue('ja,en-US;q=0.9,en;q=0.8', 'accept_language'),
		).toBe('ja,en-US;q=0.9,en;q=0.8');
	});

	it('resolveLocaleSelectValue: 不一致はカスタム', () => {
		expect(resolveLocaleSelectValue('ja', 'lang')).toBe(LOCALE_PRESET_CUSTOM);
		expect(resolveLocaleSelectValue('ja,en-US;q=0.9', 'accept_language')).toBe(
			LOCALE_PRESET_CUSTOM,
		);
	});
});
