/** セレクトの「カスタム」用 sentinel。config には保存しない。 */
export const LOCALE_PRESET_CUSTOM = '__custom__';

export type LocalePresetField = 'lang' | 'accept_language';

export type LocalePreset = {
	/** 安定キー（表示ラベルの lookup にも使う） */
	id: string;
	lang: string;
	acceptLanguage: string;
};

/** 主な国の BCP47 / Accept-Language プリセット（手書きヘッダ） */
export const LOCALE_PRESETS: readonly LocalePreset[] = [
	{
		id: 'ja-JP',
		lang: 'ja-JP',
		acceptLanguage: 'ja,en-US;q=0.9,en;q=0.8',
	},
	{
		id: 'en-US',
		lang: 'en-US',
		acceptLanguage: 'en-US,en;q=0.9',
	},
	{
		id: 'en-GB',
		lang: 'en-GB',
		acceptLanguage: 'en-GB,en;q=0.9',
	},
	{
		id: 'de-DE',
		lang: 'de-DE',
		acceptLanguage: 'de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7',
	},
	{
		id: 'fr-FR',
		lang: 'fr-FR',
		acceptLanguage: 'fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7',
	},
	{
		id: 'ko-KR',
		lang: 'ko-KR',
		acceptLanguage: 'ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7',
	},
	{
		id: 'zh-CN',
		lang: 'zh-CN',
		acceptLanguage: 'zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7',
	},
	{
		id: 'zh-TW',
		lang: 'zh-TW',
		acceptLanguage: 'zh-TW,zh;q=0.9,en-US;q=0.8,en;q=0.7',
	},
	{
		id: 'es-ES',
		lang: 'es-ES',
		acceptLanguage: 'es-ES,es;q=0.9,en-US;q=0.8,en;q=0.7',
	},
	{
		id: 'pt-BR',
		lang: 'pt-BR',
		acceptLanguage: 'pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7',
	},
] as const;

/** フィールド種別に応じた option value（厳密一致用）を返す。 */
export function localePresetOptionValue(
	preset: LocalePreset,
	field: LocalePresetField,
): string {
	return field === 'lang' ? preset.lang : preset.acceptLanguage;
}

/**
 * 現在の設定値からセレクトの value を解決する。
 * 空 → ""、プリセット厳密一致 → その value、それ以外 → カスタム。
 */
export function resolveLocaleSelectValue(
	current: string | undefined,
	field: LocalePresetField,
): string {
	const v = current ?? '';
	if (v === '') return '';
	const match = LOCALE_PRESETS.find(
		(p) => localePresetOptionValue(p, field) === v,
	);
	return match ? localePresetOptionValue(match, field) : LOCALE_PRESET_CUSTOM;
}
