import { cn } from '@/lib/utils';
import type { PartialConfig } from '@/types/config';

export type ConfigLayer = 'app' | 'workspace' | 'node';

export type FieldErrors = Record<string, string>;

/** WS・ノードでは content.formats を編集・保存しない（上位で導出）。 */
export function stripContentFormats(config: PartialConfig): PartialConfig {
	if (!config.content) return config;
	const { formats: _formats, ...contentRest } = config.content;
	const hasContent = Object.keys(contentRest).length > 0;
	if (!hasContent) {
		const { content: _c, ...rest } = config;
		return rest;
	}
	return { ...config, content: contentRest };
}

export function sanitizeConfigForLayer(
	config: PartialConfig,
	layer: ConfigLayer,
): PartialConfig {
	if (layer === 'app') {
		return config;
	}
	return stripContentFormats(config);
}

export function fieldError(
	errors: FieldErrors,
	path: string,
): string | undefined {
	return errors[path];
}

export function fieldInvalid(errors: FieldErrors, path: string): boolean {
	return Boolean(errors[path]);
}

export function inputClassName(
	invalid: boolean,
	base = 'mt-1 h-8 text-xs',
): string {
	return cn(
		base,
		invalid &&
			'border-destructive bg-destructive/5 focus-visible:border-destructive focus-visible:ring-destructive/40',
	);
}

export function selectClassName(
	invalid: boolean,
	base = 'mt-1 h-8 w-full rounded-lg border border-input bg-background px-2 text-xs',
): string {
	return cn(
		base,
		invalid &&
			'border-destructive bg-destructive/5 focus-visible:border-destructive focus-visible:ring-destructive/40',
	);
}

export function textareaClassName(
	invalid: boolean,
	base = 'min-h-16 w-full rounded-lg border border-input bg-background p-2 font-mono text-xs',
): string {
	return cn(
		base,
		invalid &&
			'border-destructive bg-destructive/5 focus-visible:border-destructive focus-visible:ring-destructive/40',
	);
}

/** 空欄は undefined。数字以外は NaN（zod で「数値を入力してください」） */
export function parseOptionalNumber(raw: string): number | undefined {
	const trimmed = raw.trim();
	if (trimmed === '') return undefined;
	if (!/^-?\d+$/.test(trimmed)) return Number.NaN;
	return Number(trimmed);
}

export function formatOptionalNumber(value: number | undefined): string {
	if (value === undefined || Number.isNaN(value)) return '';
	return String(value);
}
