import { Events } from '@wailsio/runtime';
import { useEffect } from 'react';
import { crawlResultFromDTO } from '@/lib/wailsMappers';
import { useAppStore } from '@/stores/appStore';
import type { CrawlResultDTO } from '../../bindings/scraperbot-front/internal/model/models.js';

const TOPIC_NODE_RESULT_UPDATED = 'node-result:updated';

function parseNodeResultUpdated(data: unknown): {
	workspaceId: string;
	nodeId: string;
	result: ReturnType<typeof crawlResultFromDTO>;
} | null {
	if (!data || typeof data !== 'object') return null;
	const raw = data as Record<string, unknown>;
	if (
		typeof raw.workspaceId !== 'string' ||
		typeof raw.nodeId !== 'string' ||
		!raw.result
	) {
		return null;
	}
	return {
		workspaceId: raw.workspaceId,
		nodeId: raw.nodeId,
		result: crawlResultFromDTO(raw.result as CrawlResultDTO),
	};
}

/** メインウィンドウでノード結果の手動編集を他ウィンドウと同期する。 */
export function useNodeResultSync() {
	const applyNodeResultFromSync = useAppStore((s) => s.applyNodeResultFromSync);

	useEffect(() => {
		const off = Events.On(TOPIC_NODE_RESULT_UPDATED, (ev) => {
			const payload = parseNodeResultUpdated(ev.data);
			if (!payload) return;
			applyNodeResultFromSync(
				payload.workspaceId,
				payload.nodeId,
				payload.result,
			);
		});
		return off;
	}, [applyNodeResultFromSync]);
}
