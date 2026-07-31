import { describe, expect, it } from 'vitest';
import { defaultsForLayer, sanitizeConfigForLayer } from './configFormUtils';

// レイヤー別の保存前サニタイズ（ノードは content のみ、WS は formats 除去）を検証する。
describe('sanitizeConfigForLayer', () => {
	it('ノード層は content（formats 除く）のみ残し他セクションを落とす', () => {
		const out = sanitizeConfigForLayer(
			{
				request: { timeout: '10s' },
				content: {
					formats: ['markdown'],
					only_main_content: true,
					selector: 'main',
				},
				crawl: { max_depth: 2 },
			},
			'node',
		);
		expect(out).toEqual({
			content: { only_main_content: true, selector: 'main' },
		});
	});

	it('ノード層で content が無い場合は空オブジェクトを返す', () => {
		expect(
			sanitizeConfigForLayer({ request: { timeout: '1s' } }, 'node'),
		).toEqual({});
	});

	it('ワークスペース層は formats を落としつつ他セクションは残す', () => {
		const out = sanitizeConfigForLayer(
			{
				request: { timeout: '10s' },
				content: { formats: ['html'], selector: 'body' },
			},
			'workspace',
		);
		expect(out).toEqual({
			request: { timeout: '10s' },
			content: { selector: 'body' },
		});
	});
});

// レイヤー別リセット先を検証する。
describe('defaultsForLayer', () => {
	it('ノード層のリセット先は空オブジェクト', () => {
		expect(
			defaultsForLayer(
				{ request: { timeout: '5s' }, content: { selector: 'main' } },
				'node',
			),
		).toEqual({});
	});
});
