import type { ContentFormat, PartialConfig } from '@/types/config';
import { DEFAULT_APP_CONFIG } from './defaults';
import { mergeConfig } from './mergeConfig';

const TRANSFORMER_FORMATS = [
	'markdown',
	'html',
	'raw_html',
	'json',
] as const satisfies readonly ContentFormat[];

export type TransformerFormat = (typeof TRANSFORMER_FORMATS)[number];

/** 設定から transformer 本文形式を返す。 */
export function getTransformerFormat(
	settings: PartialConfig,
): TransformerFormat {
	const t = settings.plugins?.transformer ?? 'markdown';
	if ((TRANSFORMER_FORMATS as readonly string[]).includes(t)) {
		return t as TransformerFormat;
	}
	return 'markdown';
}

/** transformer + extract フラグから content.formats を導出する。 */
export function deriveContentFormats(settings: PartialConfig): ContentFormat[] {
	const formats: ContentFormat[] = [getTransformerFormat(settings)];
	if (settings.content?.extract_metadata !== false) {
		formats.push('metadata');
	}
	if (settings.content?.extract_links !== false) {
		formats.push('links');
	}
	return formats;
}

/** 結果プレビュー用タブ（deriveContentFormats と同じ）。 */
export function getPreviewTabs(settings: PartialConfig): ContentFormat[] {
	return deriveContentFormats(settings);
}

/** content.formats を導出値で上書きした設定を返す。 */
export function withDerivedContentFormats(
	settings: PartialConfig,
	base?: PartialConfig,
): PartialConfig {
	const merged = base ? mergeConfig(base, settings) : mergeConfig(settings);
	return {
		...settings,
		content: {
			...settings.content,
			formats: deriveContentFormats(merged),
		},
	};
}

/** プレビュー用に app / ws / domain / node をマージした設定。 */
export function mergedPreviewSettings(
	appDefaults: PartialConfig,
	wsSettings?: PartialConfig,
	domainSettings?: PartialConfig,
	nodeSettings?: PartialConfig,
): PartialConfig {
	return mergeConfig(
		appDefaults ?? DEFAULT_APP_CONFIG,
		wsSettings,
		domainSettings,
		nodeSettings,
	);
}

/** CrawlResultPreview から transformer 形式の本文スニペットを返す。 */
export function bodySnippetForFormat(
	result: {
		markdown?: string;
		html?: string;
		raw_html?: string;
	},
	format: TransformerFormat,
	maxLen = 200,
): string {
	const body =
		format === 'html'
			? result.html
			: format === 'raw_html'
				? result.raw_html
				: result.markdown;
	if (!body) return '—';
	return body.length <= maxLen ? body : `${body.slice(0, maxLen)}…`;
}

export function previewTabLabel(format: ContentFormat): string {
	switch (format) {
		case 'markdown':
			return 'Markdown';
		case 'html':
			return 'HTML';
		case 'raw_html':
			return 'Raw HTML';
		case 'json':
			return 'JSON';
		case 'links':
			return 'Links';
		case 'metadata':
			return 'Metadata';
		default:
			return format;
	}
}
