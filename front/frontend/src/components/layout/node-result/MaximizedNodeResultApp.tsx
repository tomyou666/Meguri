import { Events } from '@wailsio/runtime';
import { useEffect, useState } from 'react';
import { scraperPort } from '@/adapters';
import { NodeResultPanel } from '@/components/layout/node-result/NodeResultPanel';
import { Badge } from '@/components/ui/badge';
import { Toaster } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import { messages } from '@/i18n/messages';
import { crawlResultFromDTO } from '@/lib/wailsMappers';
import type { MaximizedNodeResultSnapshot } from '@/types/adapter';
import type { ContentFormat } from '@/types/config';
import type { CrawlResultPreview } from '@/types/crawl';
import type { CrawlResultDTO } from '../../../../bindings/scraperbot-front/internal/model/models.js';

const TOPIC_NODE_RESULT_MAXIMIZE = 'node-result:maximize';
const TOPIC_NODE_RESULT_UPDATED = 'node-result:updated';

function snapshotFromEventData(
	data: unknown,
): MaximizedNodeResultSnapshot | null {
	if (!data || typeof data !== 'object') return null;
	const raw = data as Record<string, unknown>;
	if (!raw.result || typeof raw.title !== 'string') return null;
	return {
		title: raw.title,
		workspaceId: String(raw.workspaceId ?? ''),
		nodeId: String(raw.nodeId ?? ''),
		activeFormat: String(raw.activeFormat ?? 'markdown'),
		markdownView: raw.markdownView === 'source' ? 'source' : 'preview',
		formats: Array.isArray(raw.formats)
			? raw.formats.map(String)
			: ['markdown'],
		result: crawlResultFromDTO(raw.result as CrawlResultDTO),
	};
}

function updatedResultFromEventData(data: unknown): {
	nodeId: string;
	result: CrawlResultPreview;
} | null {
	if (!data || typeof data !== 'object') return null;
	const raw = data as Record<string, unknown>;
	if (!raw.result || typeof raw.nodeId !== 'string') return null;
	return {
		nodeId: raw.nodeId,
		result: crawlResultFromDTO(raw.result as CrawlResultDTO),
	};
}

export function MaximizedNodeResultApp() {
	const [snapshot, setSnapshot] = useState<MaximizedNodeResultSnapshot | null>(
		null,
	);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		let cancelled = false;

		void scraperPort.getMaximizedNodeResult().then((initial) => {
			if (!cancelled && initial) setSnapshot(initial);
			if (!cancelled) setLoading(false);
		});

		const offMaximize = Events.On(TOPIC_NODE_RESULT_MAXIMIZE, (ev) => {
			const next = snapshotFromEventData(ev.data);
			if (next) setSnapshot(next);
		});

		const offUpdated = Events.On(TOPIC_NODE_RESULT_UPDATED, (ev) => {
			const updated = updatedResultFromEventData(ev.data);
			if (!updated) return;
			setSnapshot((prev) => {
				if (!prev || prev.nodeId !== updated.nodeId) return prev;
				return { ...prev, result: updated.result };
			});
		});

		return () => {
			cancelled = true;
			offMaximize();
			offUpdated();
		};
	}, []);

	if (loading) {
		return (
			<div className='flex h-screen items-center justify-center bg-card text-sm text-muted-foreground'>
				{messages.right.noResultApi}
			</div>
		);
	}

	if (!snapshot) {
		return (
			<div className='flex h-screen items-center justify-center bg-card text-sm text-muted-foreground'>
				{messages.right.noResultApi}
			</div>
		);
	}

	return (
		<TooltipProvider>
			<div className='flex h-screen flex-col bg-card text-foreground'>
				<header className='border-b border-border px-4 py-3'>
					<p className='text-sm font-semibold'>{messages.right.nodeResult}</p>
					<p className='truncate text-sm text-muted-foreground'>
						{snapshot.title}
					</p>
					{snapshot.result.manuallyEdited && (
						<Badge variant='secondary' className='mt-1 text-[10px] font-normal'>
							{messages.right.manuallyEdited}
						</Badge>
					)}
				</header>
				<NodeResultPanel
					key={`${snapshot.nodeId}-${snapshot.activeFormat}-${snapshot.markdownView}`}
					panelMode='maximized'
					workspaceId={snapshot.workspaceId}
					nodeId={snapshot.nodeId}
					formats={snapshot.formats as ContentFormat[]}
					result={snapshot.result}
					initialTab={snapshot.activeFormat as ContentFormat}
					initialMarkdownView={snapshot.markdownView}
					onResultChange={(result) =>
						setSnapshot((prev) => (prev ? { ...prev, result } : prev))
					}
				/>
				<Toaster duration={5000} />
			</div>
		</TooltipProvider>
	);
}
