import { useEffect, useMemo, useState } from 'react';
import ReactDiffViewer from 'react-diff-viewer-continued';
import { scraperPort } from '@/adapters';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { messages } from '@/i18n/messages';
import type { DiffKind, NodeDiffDetail } from '@/types/adapter';

type NodeDiffViewerPanelProps = {
	workspaceId: string;
	nodeId: string;
	initialKind?: DiffKind;
};

function useDarkTheme(): boolean {
	const [dark, setDark] = useState(() =>
		document.documentElement.classList.contains('dark'),
	);
	useEffect(() => {
		const el = document.documentElement;
		const obs = new MutationObserver(() => {
			setDark(el.classList.contains('dark'));
		});
		obs.observe(el, { attributes: true, attributeFilter: ['class'] });
		return () => obs.disconnect();
	}, []);
	return dark;
}

function fetchLabel(value: string): string {
	const labels = messages.diff.fetchState as Record<string, string>;
	return labels[value] ?? value;
}

export function NodeDiffViewerPanel({
	workspaceId,
	nodeId,
	initialKind,
}: NodeDiffViewerPanelProps) {
	const [detail, setDetail] = useState<NodeDiffDetail | null>(null);
	const [loading, setLoading] = useState(true);
	const [tab, setTab] = useState<DiffKind>('content');
	const dark = useDarkTheme();

	useEffect(() => {
		let cancelled = false;
		setLoading(true);
		void scraperPort.getNodeDiffDetail(workspaceId, nodeId).then((d) => {
			if (cancelled) return;
			setDetail(d);
			const first = initialKind ?? d.kinds[0] ?? 'content';
			setTab(first);
			setLoading(false);
		});
		return () => {
			cancelled = true;
		};
	}, [workspaceId, nodeId, initialKind]);

	const kinds = detail?.kinds ?? [];
	const title = detail?.url ?? nodeId;

	const pairForTab = useMemo(() => {
		if (!detail) return null;
		if (tab === 'content' && detail.content) {
			return { old: detail.content.old, new: detail.content.new };
		}
		if (tab === 'links' && detail.links) {
			return { old: detail.links.old, new: detail.links.new };
		}
		if (tab === 'fetch' && detail.fetch) {
			return {
				old: fetchLabel(detail.fetch.old),
				new: fetchLabel(detail.fetch.new),
			};
		}
		return null;
	}, [detail, tab]);

	return (
		<div className='flex h-screen flex-col overflow-hidden bg-background p-4'>
			<h1 className='mb-3 truncate text-sm font-semibold'>{title}</h1>
			{loading && (
				<p className='text-xs text-muted-foreground'>{messages.diff.loading}</p>
			)}
			{!loading && detail && kinds.length > 0 && (
				<Tabs
					value={tab}
					onValueChange={(v) => setTab(v as DiffKind)}
					className='flex min-h-0 flex-1 flex-col'
				>
					<TabsList className='shrink-0'>
						{kinds.includes('content') && (
							<TabsTrigger value='content'>
								{messages.diff.kindContent}
							</TabsTrigger>
						)}
						{kinds.includes('links') && (
							<TabsTrigger value='links'>{messages.diff.kindLinks}</TabsTrigger>
						)}
						{kinds.includes('fetch') && (
							<TabsTrigger value='fetch'>{messages.diff.kindFetch}</TabsTrigger>
						)}
					</TabsList>
					<TabsContent
						value={tab}
						className='mt-2 min-h-0 flex-1 overflow-auto'
					>
						{pairForTab && (
							<ReactDiffViewer
								oldValue={pairForTab.old}
								newValue={pairForTab.new}
								splitView
								useDarkTheme={dark}
							/>
						)}
					</TabsContent>
				</Tabs>
			)}
		</div>
	);
}
