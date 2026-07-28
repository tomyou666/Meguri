import { describe, expect, it } from 'vitest';
import {
	buildNodeStatusPatches,
	collectRunningNodeIds,
	nodeSucceededStatusPatch,
} from './nodeStatusReconcile';

// 完了後 reconcile 用の running 収集と status パッチ変換を検証する。
describe('nodeStatusReconcile', () => {
	it('UI 上 running のノード ID だけを集める', () => {
		expect(
			collectRunningNodeIds([
				{ id: 'n1', status: 'running' },
				{ id: 'n2', status: 'success' },
				{ id: 'n3', status: 'running' },
			]),
		).toEqual(['n1', 'n3']);
	});

	it('API 結果を status / lastError パッチに変換する', () => {
		expect(
			buildNodeStatusPatches([
				{ nodeId: 'n1', status: 'success' },
				{ nodeId: 'n2', status: 'error', lastError: 'boom' },
				{ nodeId: 'n3', status: 'running' },
				{ nodeId: 'n4', status: 'unknown' },
			]),
		).toEqual([
			{ nodeId: 'n1', status: 'success', lastError: undefined },
			{ nodeId: 'n2', status: 'error', lastError: 'boom' },
			{ nodeId: 'n3', status: 'running', lastError: undefined },
		]);
	});

	it('result なしの nodeSucceeded で status が success になる', () => {
		expect(nodeSucceededStatusPatch('n1')).toEqual({
			nodeId: 'n1',
			status: 'success',
			lastError: undefined,
		});
	});
});
