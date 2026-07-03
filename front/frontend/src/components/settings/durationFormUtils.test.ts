import { describe, expect, it } from 'vitest';
import {
	durationInRange,
	formatDurationForSave,
	parseDurationForForm,
	parseGoDurationToMs,
} from './durationFormUtils';

// Go duration 文字列のパース・フォーム表示・保存形式を検証する。
describe('durationFormUtils', () => {
	it('秒・ミリ秒・分をミリ秒に変換する', () => {
		expect(parseGoDurationToMs('30s')).toBe(30_000);
		expect(parseGoDurationToMs('500ms')).toBe(500);
		expect(parseGoDurationToMs('2m')).toBe(120_000);
	});

	it('小数 duration はミリ秒に丸める', () => {
		expect(parseGoDurationToMs('1.5s')).toBe(1_500);
	});

	it('不正な文字列は null を返す', () => {
		expect(parseGoDurationToMs('abc')).toBeNull();
		expect(parseGoDurationToMs('30x')).toBeNull();
	});

	it('フォーム表示は s/ms に正規化する', () => {
		expect(parseDurationForForm('30s')).toEqual({ amount: 30, unit: 's' });
		expect(parseDurationForForm('500ms')).toEqual({ amount: 500, unit: 'ms' });
		expect(parseDurationForForm('2m')).toEqual({ amount: 120, unit: 's' });
		expect(parseDurationForForm('1.5s')).toEqual({ amount: 1500, unit: 'ms' });
		expect(parseDurationForForm('')).toEqual({ amount: undefined, unit: 's' });
	});

	it('保存形式は整数 + 単位サフィックス', () => {
		expect(formatDurationForSave(30, 's')).toBe('30s');
		expect(formatDurationForSave(500, 'ms')).toBe('500ms');
		expect(formatDurationForSave(undefined, 's')).toBeUndefined();
	});

	it('フィールド別レンジを判定する', () => {
		expect(durationInRange('timeout', 30_000)).toBe(true);
		expect(durationInRange('timeout', 500)).toBe(false);
		expect(durationInRange('request_delay', 0)).toBe(true);
		expect(durationInRange('retry_interval', 50)).toBe(false);
	});
});
