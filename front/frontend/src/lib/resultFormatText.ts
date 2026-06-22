import type { UpdateNodeResultPatch } from '@/types/adapter';
import type { ContentFormat } from '@/types/config';
import type { CrawlResultPreview } from '@/types/crawl';

const EDITABLE_FORMATS = new Set<ContentFormat>([
	'markdown',
	'html',
	'raw_html',
	'json',
]);

/** フォーマットが手動編集可能かどうか。 */
export function isEditableFormat(format: ContentFormat): boolean {
	return EDITABLE_FORMATS.has(format);
}

/** コピー・最大化用のプレーンテキストを返す。 */
export function resultTextForFormat(
	result: CrawlResultPreview,
	format: ContentFormat,
): string {
	switch (format) {
		case 'markdown':
			return result.markdown ?? '';
		case 'html':
			return result.html ?? '';
		case 'raw_html':
			return result.raw_html ?? '';
		case 'json':
			return result.json ?? '';
		case 'links':
			return (result.links ?? []).join('\n');
		case 'metadata':
			return JSON.stringify(result.metadata ?? {}, null, 2);
		default:
			return '';
	}
}

/** 保存用パッチオブジェクトを構築する。 */
export function updatePatchForFormat(
	format: ContentFormat,
	value: string,
): UpdateNodeResultPatch | null {
	switch (format) {
		case 'markdown':
			return { markdown: value };
		case 'html':
			return { html: value };
		case 'raw_html':
			return { raw_html: value };
		case 'json':
			return { json: value };
		default:
			return null;
	}
}

/** 編集フォームの初期値を返す。 */
export function editableValueForFormat(
	result: CrawlResultPreview,
	format: ContentFormat,
): string {
	return resultTextForFormat(result, format);
}
