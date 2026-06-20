import { describe, expect, it } from 'vitest';
import {
	deriveContentFormats,
	getPreviewTabs,
	getTransformerFormat,
	withDerivedContentFormats,
} from './previewFormats';

// transformer と extract フラグから formats / プレビュータブを導出する。
describe('previewFormats', () => {
	it('transformer のみ ON のとき metadata と links を含む', () => {
		expect(
			deriveContentFormats({
				plugins: { transformer: 'html' },
				content: {},
			}),
		).toEqual(['html', 'metadata', 'links']);
	});

	it('extract_metadata OFF のとき metadata を除外する', () => {
		expect(
			getPreviewTabs({
				plugins: { transformer: 'raw_html' },
				content: { extract_metadata: false },
			}),
		).toEqual(['raw_html', 'links']);
	});

	it('extract_links OFF のとき links を除外する', () => {
		expect(
			getPreviewTabs({
				plugins: { transformer: 'markdown' },
				content: { extract_links: false },
			}),
		).toEqual(['markdown', 'metadata']);
	});

	it('withDerivedContentFormats が content.formats を書き込む', () => {
		const out = withDerivedContentFormats({
			plugins: { transformer: 'html' },
			content: { extract_links: false },
		});
		expect(out.content?.formats).toEqual(['html', 'metadata']);
	});

	it('未知 transformer は markdown にフォールバックする', () => {
		expect(getTransformerFormat({ plugins: { transformer: 'unknown' } })).toBe(
			'markdown',
		);
	});
});
