import type { z } from 'zod';
import { messages } from '@/i18n/messages';

export type DurationUnit = 's' | 'ms';

export type DurationFieldKey =
	| 'timeout'
	| 'retry_interval'
	| 'request_delay'
	| 'wait_timeout'
	| 'network_idle_duration';

export type DurationFormValue = {
	amount: number | undefined;
	unit: DurationUnit;
};

const DURATION_TOKEN_RE = /([-+]?(?:\d+\.?\d*|\.\d+))(ns|us|µs|ms|s|m|h)/gi;

const UNIT_TO_MS: Record<string, number> = {
	ns: 1 / 1_000_000,
	us: 1 / 1_000,
	µs: 1 / 1_000,
	ms: 1,
	s: 1_000,
	m: 60_000,
	h: 3_600_000,
};

export const DURATION_LIMITS: Record<
	DurationFieldKey,
	{ minMs: number; maxMs: number }
> = {
	timeout: { minMs: 1_000, maxMs: 300_000 },
	retry_interval: { minMs: 100, maxMs: 60_000 },
	request_delay: { minMs: 0, maxMs: 60_000 },
	wait_timeout: { minMs: 0, maxMs: 120_000 },
	network_idle_duration: { minMs: 100, maxMs: 30_000 },
};

const DURATION_RANGE_MESSAGES: Record<DurationFieldKey, string> = {
	timeout: messages.settings.validation.timeoutRange,
	retry_interval: messages.settings.validation.retryIntervalRange,
	request_delay: messages.settings.validation.requestDelayRange,
	wait_timeout: messages.settings.validation.waitTimeoutRange,
	network_idle_duration: messages.settings.validation.networkIdleDurationRange,
};

/** Go time.ParseDuration 互換の文字列をミリ秒に変換する */
export function parseGoDurationToMs(raw: string): number | null {
	const s = raw.trim();
	if (!s) return null;

	let totalMs = 0;
	let consumed = 0;
	const re = new RegExp(DURATION_TOKEN_RE.source, 'gi');
	let match: RegExpExecArray | null = re.exec(s);
	while (match !== null) {
		const n = Number.parseFloat(match[1]);
		const unit =
			match[2].toLowerCase() === 'µs' ? 'µs' : match[2].toLowerCase();
		const factor = UNIT_TO_MS[unit];
		if (!Number.isFinite(n) || factor === undefined) return null;
		totalMs += n * factor;
		consumed = re.lastIndex;
		match = re.exec(s);
	}

	if (consumed !== s.length) return null;
	return Math.round(totalMs);
}

/** 保存済み duration 文字列をフォーム表示用に分解する */
export function parseDurationForForm(raw: string): DurationFormValue {
	const trimmed = raw.trim();
	if (!trimmed) {
		return { amount: undefined, unit: 's' };
	}

	const totalMs = parseGoDurationToMs(trimmed);
	if (totalMs === null) {
		return { amount: undefined, unit: 's' };
	}

	if (totalMs % 1_000 === 0 && totalMs >= 1_000) {
		return { amount: totalMs / 1_000, unit: 's' };
	}
	return { amount: totalMs, unit: 'ms' };
}

/** フォーム値を Go duration 文字列に戻す */
export function formatDurationForSave(
	amount: number | undefined,
	unit: DurationUnit,
): string | undefined {
	if (amount === undefined || Number.isNaN(amount)) return undefined;
	if (!Number.isInteger(amount) || amount < 0) return undefined;
	return `${amount}${unit}`;
}

export function isValidGoDuration(raw: string): boolean {
	return parseGoDurationToMs(raw) !== null;
}

export function durationInRange(
	field: DurationFieldKey,
	valueMs: number,
): boolean {
	const { minMs, maxMs } = DURATION_LIMITS[field];
	return valueMs >= minMs && valueMs <= maxMs;
}

export function getDurationRangeMessage(field: DurationFieldKey): string {
	return DURATION_RANGE_MESSAGES[field];
}

/** zod superRefine 用: フィールド別レンジ検証 */
export function createDurationRangeRefine(field: DurationFieldKey) {
	return (val: string | undefined, ctx: z.RefinementCtx) => {
		if (val === undefined) return;
		const ms = parseGoDurationToMs(val);
		if (ms === null) return;
		if (!durationInRange(field, ms)) {
			ctx.addIssue({
				code: 'custom',
				message: getDurationRangeMessage(field),
			});
		}
	};
}
