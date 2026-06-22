import { describe, expect, it } from 'vitest';
import {
	isEditableFormat,
	resultTextForFormat,
	updatePatchForFormat,
} from '@/lib/resultFormatText';
import type { CrawlResultPreview } from '@/types/crawl';

const sample: CrawlResultPreview = {
	url: 'https://example.com',
	markdown: '# Title',
	html: '<p>hi</p>',
	raw_html: '<div>raw</div>',
	json: '{"a":1}',
	links: ['https://a.test', 'https://b.test'],
	metadata: { title: 'Example' },
};

describe('resultFormatText', () => {
	it('isEditableFormat は編集可能フォーマットのみ true', () => {
		expect(isEditableFormat('markdown')).toBe(true);
		expect(isEditableFormat('links')).toBe(false);
	});

	it('resultTextForFormat はフォーマット別テキストを返す', () => {
		expect(resultTextForFormat(sample, 'markdown')).toBe('# Title');
		expect(resultTextForFormat(sample, 'links')).toBe(
			'https://a.test\nhttps://b.test',
		);
		expect(resultTextForFormat(sample, 'metadata')).toBe(
			JSON.stringify({ title: 'Example' }, null, 2),
		);
	});

	it('updatePatchForFormat は保存パッチを構築する', () => {
		expect(updatePatchForFormat('raw_html', 'x')).toEqual({ raw_html: 'x' });
		expect(updatePatchForFormat('links', 'x')).toBeNull();
	});
});
