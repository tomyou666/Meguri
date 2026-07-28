import { describe, expect, it, vi } from 'vitest';
import { createCrawlEventGate } from './crawlEventQueue';

// runId 確定前後の Event キュー再生を検証する。
describe('createCrawlEventGate', () => {
	it('runId 確定前のイベントを、確定後に一致するものだけ再生する', () => {
		const received: string[] = [];
		const gate = createCrawlEventGate<{ runId: string; nodeId: string }>(
			(p) => {
				received.push(p.nodeId);
			},
		);

		gate.dispatch({ runId: 'run-a', nodeId: 'n1' });
		gate.dispatch({ runId: 'run-b', nodeId: 'n-other' });
		expect(received).toEqual([]);

		gate.setRunId('run-a');
		expect(received).toEqual(['n1']);

		gate.dispatch({ runId: 'run-a', nodeId: 'n2' });
		gate.dispatch({ runId: 'run-b', nodeId: 'n3' });
		expect(received).toEqual(['n1', 'n2']);
	});

	it('runId 確定後はキューを経由せず即時処理する', () => {
		const handler = vi.fn();
		const gate = createCrawlEventGate<{ runId: string }>(handler);
		gate.setRunId('run-1');
		gate.dispatch({ runId: 'run-1' });
		expect(handler).toHaveBeenCalledTimes(1);
	});
});
