import { Events } from '@wailsio/runtime';
import { useEffect, useState } from 'react';
import { scraperPort } from '@/adapters';
import { NodeDiffViewerPanel } from '@/components/diff/NodeDiffViewerPanel';
import { Toaster } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import type { DiffKind, NodeDiffViewerSnapshot } from '@/types/adapter';

const TOPIC_NODE_DIFF_OPEN = 'node-diff:open';

function snapshotFromEventData(data: unknown): NodeDiffViewerSnapshot | null {
	if (!data || typeof data !== 'object') return null;
	const raw = data as Record<string, unknown>;
	if (!raw.workspaceId || !raw.nodeId) return null;
	return {
		workspaceId: String(raw.workspaceId),
		nodeId: String(raw.nodeId),
		initialKind: raw.initialKind
			? (String(raw.initialKind) as DiffKind)
			: undefined,
		title: String(raw.title ?? ''),
	};
}

export function NodeDiffApp() {
	const [snapshot, setSnapshot] = useState<NodeDiffViewerSnapshot | null>(null);

	useEffect(() => {
		void scraperPort.getNodeDiffViewerSession().then((initial) => {
			if (initial) setSnapshot(initial);
		});

		const off = Events.On(TOPIC_NODE_DIFF_OPEN, (ev) => {
			const next = snapshotFromEventData(ev.data);
			if (next) setSnapshot(next);
		});
		return () => off();
	}, []);

	if (!snapshot) {
		return (
			<div className='flex h-screen items-center justify-center text-sm text-muted-foreground'>
				Loading…
			</div>
		);
	}

	return (
		<TooltipProvider>
			<NodeDiffViewerPanel
				workspaceId={snapshot.workspaceId}
				nodeId={snapshot.nodeId}
				initialKind={snapshot.initialKind}
			/>
			<Toaster duration={5000} />
		</TooltipProvider>
	);
}
