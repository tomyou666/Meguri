import { describe, expect, it } from 'vitest';
import { messages } from '@/i18n/messages';
import { formatNotifyMessage } from '@/lib/notify';

/** 通知文言の整形（JSON 展開・既知エラーの日本語化）を検証する。 */
describe('formatNotifyMessage', () => {
	it('RuntimeError JSON から message を取り出して日本語化する', () => {
		const raw = JSON.stringify({
			message: 'export preview session expired',
			cause: {},
			kind: 'RuntimeError',
		});
		expect(formatNotifyMessage(raw)).toBe(
			messages.export.previewSessionExpired,
		);
	});

	it('通常の文字列はそのまま返す', () => {
		expect(formatNotifyMessage('something failed')).toBe('something failed');
	});

	it('空文字は不明エラーにする', () => {
		expect(formatNotifyMessage('   ')).toBe(messages.error.unknown);
	});
});
