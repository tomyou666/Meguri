import { describe, expect, it } from 'vitest';
import {
	acceptBodyResponse,
	computeBodyWindow,
	type ExportPreviewRowBody,
	type ExportPreviewRowMeta,
	estimateExportRowHeight,
	pruneBodyMap,
	shouldFetchBodies,
} from '@/lib/exportPreviewVirtual';

/** 高さ見積もりと本文窓の挙動を検証する。 */
describe('exportPreviewVirtual', () => {
	it('estimates height from byte length with first-row extra', () => {
		expect(estimateExportRowHeight(0, false)).toBe(64);
		expect(estimateExportRowHeight(8192, false)).toBe(240);
		expect(estimateExportRowHeight(8192, true)).toBe(288);
	});

	it('computes body window with one viewport before and after', () => {
		expect(computeBodyWindow(10, 20, 100, 10)).toEqual({
			start: 0,
			end: 30,
		});
		expect(computeBodyWindow(0, 5, 10, 5)).toEqual({ start: 0, end: 10 });
	});

	it('fetches only when needed range leaves the held window', () => {
		const held = { start: 0, end: 20 };
		expect(shouldFetchBodies({ start: 5, end: 15 }, held)).toBe(false);
		expect(shouldFetchBodies({ start: 0, end: 25 }, held)).toBe(true);
		expect(shouldFetchBodies({ start: 0, end: 10 }, null)).toBe(true);
	});

	it('prunes bodies outside the window', () => {
		const bodies = new Map<number, ExportPreviewRowBody>([
			[0, { id: 'a:0', nodeId: 'a', url: 'u', body: '0' }],
			[1, { id: 'a:1', nodeId: 'a', url: 'u', body: '1' }],
			[2, { id: 'b:0', nodeId: 'b', url: 'u', body: '2' }],
		]);
		const next = pruneBodyMap(bodies, { start: 1, end: 2 });
		expect([...next.keys()]).toEqual([1]);
	});

	it('rejects body responses with mismatched generation or row ids', () => {
		const rows: ExportPreviewRowMeta[] = [
			{ id: 'a:0', nodeId: 'a', byteLength: 10, isFirst: true },
			{ id: 'a:5', nodeId: 'a', byteLength: 10, isFirst: false },
		];
		const ok = acceptBodyResponse({
			responseGeneration: 1,
			currentGeneration: 1,
			responseStart: 0,
			responseEnd: 2,
			requestedStart: 0,
			requestedEnd: 2,
			rows,
			responseRows: [
				{ id: 'a:0', nodeId: 'a', url: 'u', body: 'x' },
				{ id: 'a:5', nodeId: 'a', url: 'u', body: 'y' },
			],
		});
		expect(ok).toBe(true);

		expect(
			acceptBodyResponse({
				responseGeneration: 2,
				currentGeneration: 1,
				responseStart: 0,
				responseEnd: 1,
				requestedStart: 0,
				requestedEnd: 1,
				rows,
				responseRows: [{ id: 'a:0', nodeId: 'a', url: 'u', body: 'x' }],
			}),
		).toBe(false);

		expect(
			acceptBodyResponse({
				responseGeneration: 1,
				currentGeneration: 1,
				responseStart: 0,
				responseEnd: 1,
				requestedStart: 0,
				requestedEnd: 1,
				rows,
				responseRows: [{ id: 'wrong', nodeId: 'a', url: 'u', body: 'x' }],
			}),
		).toBe(false);
	});
});
