/** runId 確定前に届いた crawl Event を一時保持し、確定後に再生する。 */

type QueuedItem<T> = {
	payload: T;
	handler: (payload: T) => void;
};

/**
 * runId 未確定中は (payload, handler) をキューし、確定後に一致するものだけ再生する。
 * runId が空の間にイベントを捨てない。
 */
export function createRunIdEventBuffer<T extends { runId: string }>() {
	let runId = '';
	const queue: QueuedItem<T>[] = [];

	const accept = (payload: T, handler: (payload: T) => void) => {
		if (runId === '') {
			queue.push({ payload, handler });
			return;
		}
		if (payload.runId !== runId) return;
		handler(payload);
	};

	const setRunId = (id: string) => {
		runId = id;
		if (runId === '') return;
		const pending = queue.splice(0, queue.length);
		for (const item of pending) {
			if (item.payload.runId !== runId) continue;
			item.handler(item.payload);
		}
	};

	return { accept, setRunId, getRunId: () => runId };
}

/**
 * 単一 handler 向けの薄いラッパ（単体テスト用）。
 */
export function createCrawlEventGate<T extends { runId: string }>(
	handler: (payload: T) => void,
) {
	const buffer = createRunIdEventBuffer<T>();
	return {
		dispatch: (payload: T) => buffer.accept(payload, handler),
		setRunId: buffer.setRunId,
		getRunId: buffer.getRunId,
	};
}
