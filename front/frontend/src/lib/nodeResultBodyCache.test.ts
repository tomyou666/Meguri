import { describe, expect, it } from 'vitest';
import type { CrawlResultPreview } from '@/types/crawl';
import {
	createNodeResultBodyCache,
	estimatePreviewSize,
} from './nodeResultBodyCache';

function preview(
	partial: Partial<CrawlResultPreview> = {},
): CrawlResultPreview {
	return {
		url: partial.url ?? '',
		markdown: partial.markdown,
		html: partial.html,
		raw_html: partial.raw_html,
		json: partial.json,
		links: partial.links,
		metadata: partial.metadata,
		manuallyEdited: partial.manuallyEdited,
	};
}

// ノード本文セッションキャッシュの hit / 追い出し / drop を検証する。
describe('nodeResultBodyCache', () => {
	it('同じキーを get すると preview が返り LRU が更新される', () => {
		// 予算 25。各 10 なので 2 件で 20、3 件目を入れるには 1 件捨てる。
		const cache = createNodeResultBodyCache(25, 2);
		cache.set('ws', 'a', preview({ markdown: 'aaaaaaaaaa' }));
		cache.set('ws', 'b', preview({ markdown: 'bbbbbbbbbb' }));
		expect(cache.get('ws', 'a')?.markdown).toBe('aaaaaaaaaa');
		// a を触ったので最古は b。次の set で b が落ちる。
		cache.set('ws', 'c', preview({ markdown: 'cccccccccc' }));
		expect(cache.get('ws', 'b')).toBeNull();
		expect(cache.get('ws', 'a')?.markdown).toBe('aaaaaaaaaa');
		expect(cache.get('ws', 'c')?.markdown).toBe('cccccccccc');
	});

	it('同じキーの set は古いサイズを差し替える', () => {
		const cache = createNodeResultBodyCache(1000, 2);
		cache.set('ws', 'n1', preview({ markdown: 'old' }));
		const before = cache.totalSize();
		cache.set('ws', 'n1', preview({ markdown: 'newer-body' }));
		expect(cache.size()).toBe(1);
		expect(cache.totalSize()).toBe(
			estimatePreviewSize(preview({ markdown: 'newer-body' })),
		);
		expect(cache.totalSize()).not.toBe(before);
		expect(cache.get('ws', 'n1')?.markdown).toBe('newer-body');
	});

	it('drop でノード単位、dropWorkspace で WS 単位に消える', () => {
		const cache = createNodeResultBodyCache();
		cache.set('ws1', 'n1', preview({ markdown: 'a' }));
		cache.set('ws1', 'n2', preview({ markdown: 'b' }));
		cache.set('ws2', 'n1', preview({ markdown: 'c' }));
		cache.drop('ws1', 'n1');
		expect(cache.get('ws1', 'n1')).toBeNull();
		expect(cache.get('ws1', 'n2')?.markdown).toBe('b');
		cache.dropWorkspace('ws1');
		expect(cache.get('ws1', 'n2')).toBeNull();
		expect(cache.get('ws2', 'n1')?.markdown).toBe('c');
	});

	it('最低件数未満なら予算超過でも保持する', () => {
		const cache = createNodeResultBodyCache(10, 3);
		cache.set('ws', 'a', preview({ markdown: 'xxxxxxxxxx' }));
		cache.set('ws', 'b', preview({ markdown: 'yyyyyyyyyy' }));
		expect(cache.size()).toBe(2);
		expect(cache.totalSize()).toBeGreaterThan(10);
		expect(cache.get('ws', 'a')?.markdown).toBe('xxxxxxxxxx');
	});

	it('最低件数以上で予算超過なら LRU で空きを作り、足りなければ新しい方を入れず既存は残す', () => {
		const maxBytes = 25;
		const minEntries = 2;
		const cache = createNodeResultBodyCache(maxBytes, minEntries);
		cache.set('ws', 'a', preview({ markdown: 'aaaaaaaaaa' }));
		cache.set('ws', 'b', preview({ markdown: 'bbbbbbbbbb' }));
		expect(cache.size()).toBe(2);
		// 80 文字は 1 件捨てても入らない → 既存は触らない
		cache.set('ws', 'c', preview({ markdown: 'x'.repeat(80) }));
		expect(cache.get('ws', 'c')).toBeNull();
		expect(cache.size()).toBe(minEntries);
		expect(cache.get('ws', 'a')?.markdown).toBe('aaaaaaaaaa');
		expect(cache.get('ws', 'b')?.markdown).toBe('bbbbbbbbbb');
	});

	it('最低件数以上でも LRU で空きが出れば新しい方を入れる', () => {
		const cache = createNodeResultBodyCache(25, 2);
		cache.set('ws', 'a', preview({ markdown: 'aaaaaaaaaa' }));
		cache.set('ws', 'b', preview({ markdown: 'bbbbbbbbbb' }));
		cache.set('ws', 'c', preview({ markdown: 'cccccccccc' }));
		expect(cache.get('ws', 'a')).toBeNull();
		expect(cache.get('ws', 'b')?.markdown).toBe('bbbbbbbbbb');
		expect(cache.get('ws', 'c')?.markdown).toBe('cccccccccc');
		expect(cache.size()).toBe(2);
	});

	it('上書き set が予算超過で拒否されても古いキーは残る', () => {
		const cache = createNodeResultBodyCache(35, 2);
		cache.set('ws', 'a', preview({ markdown: 'aaaaaaaaaa' }));
		cache.set('ws', 'b', preview({ markdown: 'bbbbbbbbbb' }));
		cache.set('ws', 'c', preview({ markdown: 'cccccccccc' }));
		expect(cache.size()).toBe(3);
		// c を巨大本文で上書きしようとすると入らない → c の旧値と他エントリは残る
		cache.set('ws', 'c', preview({ markdown: 'x'.repeat(80) }));
		expect(cache.get('ws', 'c')?.markdown).toBe('cccccccccc');
		expect(cache.get('ws', 'a')?.markdown).toBe('aaaaaaaaaa');
		expect(cache.get('ws', 'b')?.markdown).toBe('bbbbbbbbbb');
		expect(cache.size()).toBe(3);
	});
});
