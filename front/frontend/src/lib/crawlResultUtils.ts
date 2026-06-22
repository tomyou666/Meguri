import type { CrawlResultPreview } from '@/types/crawl';

/** 取得結果の metadata.content_type が PDF リソースかどうか。 */
export function isPdfResourceResult(
	result?: CrawlResultPreview | null,
): boolean {
	const ct = result?.metadata?.content_type?.toLowerCase() ?? '';
	return ct.includes('application/pdf');
}
