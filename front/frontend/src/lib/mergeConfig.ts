import type { AppConfig, PartialConfig } from '@/types/config';
import { DEFAULT_APP_CONFIG } from './defaults';

function deepMerge<T extends Record<string, unknown>>(
	base: T,
	override?: Partial<T>,
): T {
	if (!override) return { ...base };
	const out = { ...base } as T;
	for (const key of Object.keys(override) as (keyof T)[]) {
		const val = override[key];
		if (val === undefined) continue;
		const baseVal = base[key];
		if (
			val &&
			typeof val === 'object' &&
			!Array.isArray(val) &&
			baseVal &&
			typeof baseVal === 'object' &&
			!Array.isArray(baseVal)
		) {
			out[key] = deepMerge(
				baseVal as Record<string, unknown>,
				val as Record<string, unknown>,
			) as T[keyof T];
		} else {
			out[key] = val as T[keyof T];
		}
	}
	return out;
}

export function mergeConfig(
	app: PartialConfig,
	ws?: PartialConfig,
	node?: PartialConfig,
): AppConfig {
	let cfg = deepMerge(
		DEFAULT_APP_CONFIG as unknown as Record<string, unknown>,
		app as Record<string, unknown>,
	) as AppConfig;
	if (ws) {
		cfg = deepMerge(
			cfg as unknown as Record<string, unknown>,
			ws as Record<string, unknown>,
		) as AppConfig;
	}
	if (node) {
		cfg = deepMerge(
			cfg as unknown as Record<string, unknown>,
			node as Record<string, unknown>,
		) as AppConfig;
	}
	return cfg;
}

/** モード2: アプリデフォルトのみ（WS/ノード上書きなし） */
export function configForMode2(app: PartialConfig): AppConfig {
	return mergeConfig(app);
}
