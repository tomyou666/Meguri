import { PanelRightClose, PanelRightOpen, Settings } from 'lucide-react';
import { useMemo, useState } from 'react';
import { CollapsedSidebarRail } from '@/components/layout/CollapsedSidebarRail';
import { ConfigEditor } from '@/components/settings/ConfigEditor';
import { ActionTooltip } from '@/components/ui/action-tooltip';
import { Alert } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { messages } from '@/i18n/messages';
import {
	bodySnippetForFormat,
	getPreviewTabs,
	getTransformerFormat,
	mergedPreviewSettings,
	previewTabLabel,
	type TransformerFormat,
} from '@/lib/previewFormats';
import { useAppStore } from '@/stores/appStore';
import type { ContentFormat } from '@/types/config';
import type { CrawlResultPreview } from '@/types/crawl';
import type { GraphNode } from '@/types/graph';

function CloseRightSidebarButton({ onClick }: { onClick: () => void }) {
	return (
		<ActionTooltip label={messages.sidebar.closeRight}>
			<Button
				variant='ghost'
				size='icon-xs'
				aria-label={messages.sidebar.closeRight}
				onClick={onClick}
			>
				<PanelRightClose className='size-3.5' />
			</Button>
		</ActionTooltip>
	);
}

export function RightSidebarContent() {
	const ws = useAppStore((s) => s.getActiveWorkspace());
	const node = useAppStore((s) => s.getSelectedNode());
	const selectedNodeIds = useAppStore((s) => s.selectedNodeIds);
	const rightCollapsed = useAppStore((s) => s.rightSidebarCollapsed);
	const runHistory = useAppStore((s) => s.runHistory);
	const crawlLogs = useAppStore((s) => s.crawlLogs);
	const crawlStatus = useAppStore((s) => s.crawlStatus);
	const crawlError = useAppStore((s) => s.crawlError);
	const loadedNodeResult = useAppStore((s) => s.loadedNodeResult);
	const resultPreview = useAppStore((s) => s.resultPreview);
	const clearCrawlError = useAppStore((s) => s.clearCrawlError);
	const appDefaults = useAppStore((s) => s.appDefaults);
	const previewSelectedResults = useAppStore((s) => s.previewSelectedResults);
	const saveSelectedResults = useAppStore((s) => s.saveSelectedResults);
	const deleteSelectedResults = useAppStore((s) => s.deleteSelectedResults);
	const toggleRightSidebar = useAppStore((s) => s.toggleRightSidebar);

	const formats = useMemo(() => {
		if (!ws) return getPreviewTabs(appDefaults);
		if (node) {
			return getPreviewTabs(
				mergedPreviewSettings(appDefaults, ws.settings, node.nodeSettings),
			);
		}
		return getPreviewTabs(mergedPreviewSettings(appDefaults, ws.settings));
	}, [appDefaults, ws, node]);

	const transformerFormat = useMemo((): TransformerFormat => {
		if (!ws) return getTransformerFormat(appDefaults);
		if (node) {
			return getTransformerFormat(
				mergedPreviewSettings(appDefaults, ws.settings, node.nodeSettings),
			);
		}
		return getTransformerFormat(
			mergedPreviewSettings(appDefaults, ws.settings),
		);
	}, [appDefaults, ws, node]);

	if (rightCollapsed) {
		return (
			<CollapsedSidebarRail
				icon={PanelRightOpen}
				label={messages.sidebar.openRight}
				onClick={toggleRightSidebar}
				borderSide='right'
				className='bg-card hover:bg-muted/50'
			/>
		);
	}

	const shellClass =
		'flex h-full w-full min-w-[16rem] flex-col overflow-hidden border-l border-border bg-card';

	const resultForDisplay: CrawlResultPreview | null =
		selectedNodeIds.length === 1 ? loadedNodeResult : null;

	if (selectedNodeIds.length > 1) {
		return (
			<aside className={shellClass}>
				<div className='flex items-center justify-between border-b border-border px-3 py-2'>
					<p className='text-xs font-semibold'>
						{messages.right.multiSelectCount(selectedNodeIds.length)}
					</p>
					<CloseRightSidebarButton onClick={toggleRightSidebar} />
				</div>
				<div className='flex flex-wrap gap-1 p-3'>
					<Button size='xs' onClick={() => previewSelectedResults()}>
						{messages.right.preview}
					</Button>
					<Button
						size='xs'
						variant='outline'
						onClick={() => saveSelectedResults()}
					>
						{messages.right.save}
					</Button>
					<Button
						size='xs'
						variant='destructive'
						onClick={() => deleteSelectedResults()}
					>
						{messages.right.delete}
					</Button>
				</div>
				<ScrollArea className='flex-1 p-3'>
					{resultPreview?.map((r) => (
						<div key={r.url} className='mb-3 rounded border p-2 text-xs'>
							<p className='font-medium'>{r.url}</p>
							<pre className='mt-1 whitespace-pre-wrap text-[10px]'>
								{bodySnippetForFormat(r, transformerFormat)}
							</pre>
						</div>
					))}
				</ScrollArea>
			</aside>
		);
	}

	if (node) {
		return (
			<aside className={shellClass}>
				<div className='flex items-center justify-between border-b border-border px-3 py-2'>
					<div className='min-w-0'>
						<p className='text-xs font-semibold'>{messages.right.nodeResult}</p>
						<p className='truncate text-xs text-muted-foreground'>
							{node.urlNormalized}
						</p>
						<Badge variant='outline' className='mt-1 text-[10px] font-normal'>
							{messages.right.transformerBadge(transformerFormat)}
						</Badge>
					</div>
					<CloseRightSidebarButton onClick={toggleRightSidebar} />
				</div>
				{node.status === 'error' && node.lastError && (
					<Alert variant='destructive' className='m-2 text-xs'>
						{messages.error.nodeFailed}: {node.lastError}
					</Alert>
				)}
				<NodeResultPanel
					key={node.id}
					node={node}
					formats={formats}
					result={resultForDisplay}
				/>
			</aside>
		);
	}

	return (
		<aside className={shellClass}>
			<div className='flex items-center justify-between border-b border-border px-3 py-2 text-xs font-semibold'>
				{messages.right.runSummary}
				<CloseRightSidebarButton onClick={toggleRightSidebar} />
			</div>
			{crawlError && (
				<Alert variant='destructive' className='m-2 text-xs'>
					<div className='flex justify-between gap-2'>
						<span>
							{messages.error.crawlFailed}: {crawlError.message}
						</span>
						<button type='button' onClick={clearCrawlError}>
							×
						</button>
					</div>
				</Alert>
			)}
			<ScrollArea className='flex-1 p-3'>
				<div className='space-y-4'>
					{(crawlStatus !== 'idle' || crawlLogs.length > 0) && (
						<div className='space-y-2'>
							<p className='text-xs font-medium text-muted-foreground'>
								{messages.right.crawlLog}
							</p>
							{crawlLogs.length === 0 ? (
								<p className='text-xs text-muted-foreground'>
									{messages.right.crawlLogEmpty}
								</p>
							) : (
								<ul className='max-h-40 space-y-1 overflow-y-auto text-xs text-muted-foreground'>
									{crawlLogs.map((entry) => (
										<li
											key={`${entry.at}-${entry.parentUrl}-${entry.targetUrl}-${entry.reason}`}
										>
											{messages.right.linkSkipLine(
												entry.parentUrl,
												entry.targetUrl,
												messages.right.linkSkipReason(entry.reason),
											)}
										</li>
									))}
								</ul>
							)}
						</div>
					)}
					{runHistory.length === 0 ? (
						crawlStatus === 'idle' &&
						crawlLogs.length === 0 && (
							<p className='text-xs text-muted-foreground'>
								{messages.right.noSelection}
							</p>
						)
					) : (
						<div className='space-y-2'>
							<p className='text-xs font-medium text-muted-foreground'>
								{messages.right.history}
							</p>
							{runHistory.map((run) => (
								<div
									key={run.id}
									className='rounded-lg border border-border p-2 text-xs'
								>
									<div className='flex items-center justify-between'>
										<Badge variant='secondary'>
											{messages.right.runModeBadge(run.mode)}
										</Badge>
										<span className='text-muted-foreground'>
											{run.stoppedReason ?? '—'}
										</span>
									</div>
									<p className='mt-1 text-muted-foreground'>
										{new Date(run.startedAt).toLocaleString()}
									</p>
									<p className='mt-1'>
										{messages.right.runStats(
											run.succeeded,
											run.failed,
											run.skipped,
										)}
									</p>
									{(run.skippedDuplicateLinks ?? 0) > 0 && (
										<p className='mt-1 text-muted-foreground'>
											{messages.right.runStatsDuplicateLinks(
												run.skippedDuplicateLinks ?? 0,
											)}
										</p>
									)}
								</div>
							))}
						</div>
					)}
				</div>
			</ScrollArea>
		</aside>
	);
}

/** @deprecated AppShell 内の Panel でラップするため RightSidebarContent を使用 */
export const RightSidebar = RightSidebarContent;

function NodeResultPanel({
	node,
	formats,
	result,
}: {
	node: GraphNode;
	formats: ContentFormat[];
	result: CrawlResultPreview | null;
}) {
	const persistNodeSettings = useAppStore((s) => s.persistNodeSettings);
	const [tab, setTab] = useState<ContentFormat>(formats[0] ?? 'markdown');
	const [showNodeSettings, setShowNodeSettings] = useState(false);

	return (
		<Tabs
			value={tab}
			onValueChange={(v) => {
				setTab(v as ContentFormat);
				setShowNodeSettings(false);
			}}
			className='flex min-h-0 flex-1 flex-col px-3'
		>
			<div className='flex items-center gap-1'>
				<TabsList className='min-w-0 flex-1'>
					{formats.map((f) => (
						<TabsTrigger key={f} value={f}>
							{previewTabLabel(f)}
						</TabsTrigger>
					))}
				</TabsList>
				<ActionTooltip label={messages.right.nodeSettings}>
					<Button
						variant={showNodeSettings ? 'secondary' : 'ghost'}
						size='icon-xs'
						aria-label={messages.right.nodeSettings}
						onClick={() => setShowNodeSettings((v) => !v)}
					>
						<Settings className='size-3.5' />
					</Button>
				</ActionTooltip>
			</div>
			<ScrollArea className='flex-1 py-3'>
				{showNodeSettings ? (
					<ConfigEditor
						layer='node'
						settings={node.nodeSettings ?? {}}
						compact
						onSave={(settings) => persistNodeSettings(node.id, settings)}
					/>
				) : (
					formats.map((f) => (
						<TabsContent key={f} value={f}>
							<NodeFormatContent
								format={f}
								result={result ?? node.lastResult}
							/>
						</TabsContent>
					))
				)}
			</ScrollArea>
		</Tabs>
	);
}

function NodeFormatContent({
	format,
	result,
}: {
	format: string;
	result?: CrawlResultPreview;
}) {
	if (!result) {
		return (
			<p className='text-xs text-muted-foreground'>
				{messages.right.noResultApi}
			</p>
		);
	}
	if (format === 'markdown') {
		return (
			<pre className='whitespace-pre-wrap font-mono text-xs'>
				{result.markdown ?? '—'}
			</pre>
		);
	}
	if (format === 'html') {
		return (
			<pre className='whitespace-pre-wrap font-mono text-xs'>
				{result.html ?? '—'}
			</pre>
		);
	}
	if (format === 'raw_html') {
		return (
			<pre className='whitespace-pre-wrap font-mono text-xs'>
				{result.raw_html ?? '—'}
			</pre>
		);
	}
	if (format === 'json') {
		return <pre className='whitespace-pre-wrap font-mono text-xs'>—</pre>;
	}
	if (format === 'links') {
		return (
			<ul className='list-inside list-disc text-xs'>
				{(result.links ?? []).map((l) => (
					<li key={l} className='truncate'>
						{l}
					</li>
				))}
			</ul>
		);
	}
	if (format === 'metadata') {
		return (
			<dl className='space-y-1 text-xs'>
				{Object.entries(result.metadata ?? {}).map(([k, v]) => (
					<div key={k}>
						<dt className='text-muted-foreground'>{k}</dt>
						<dd>{v}</dd>
					</div>
				))}
			</dl>
		);
	}
	return <p className='text-xs text-muted-foreground'>—</p>;
}
