import { PanelRightClose, PanelRightOpen } from 'lucide-react';
import { useMemo, useState } from 'react';
import { CollapsedSidebarRail } from '@/components/layout/CollapsedSidebarRail';
import { ConfigEditor } from '@/components/settings/ConfigEditor';
import { Alert } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { messages } from '@/i18n/messages';
import { useAppStore } from '@/stores/appStore';
import type { ContentFormat } from '@/types/config';
import type { CrawlResultPreview } from '@/types/crawl';
import { getActiveFormats } from '@/types/crawl';

export function RightSidebarContent() {
	const ws = useAppStore((s) => s.getActiveWorkspace());
	const node = useAppStore((s) => s.getSelectedNode());
	const selectedNodeIds = useAppStore((s) => s.selectedNodeIds);
	const selectedDomain = useAppStore((s) => s.selectedDomain);
	const rightCollapsed = useAppStore((s) => s.rightSidebarCollapsed);
	const runHistory = useAppStore((s) => s.runHistory);
	const crawlError = useAppStore((s) => s.crawlError);
	const loadedNodeResult = useAppStore((s) => s.loadedNodeResult);
	const resultPreview = useAppStore((s) => s.resultPreview);
	const clearCrawlError = useAppStore((s) => s.clearCrawlError);
	const appDefaults = useAppStore((s) => s.appDefaults);
	const persistDomainSettings = useAppStore((s) => s.persistDomainSettings);
	const previewSelectedResults = useAppStore((s) => s.previewSelectedResults);
	const saveSelectedResults = useAppStore((s) => s.saveSelectedResults);
	const deleteSelectedResults = useAppStore((s) => s.deleteSelectedResults);
	const toggleRightSidebar = useAppStore((s) => s.toggleRightSidebar);

	const formats = useMemo(
		() =>
			getActiveFormats(
				ws?.settings.content?.formats ?? appDefaults.content?.formats,
			),
		[ws, appDefaults],
	);

	const [tab, setTab] = useState<ContentFormat>(formats[0] ?? 'markdown');

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

	if (selectedDomain && ws) {
		const domainCfg = ws.domainSettings[selectedDomain] ?? {};
		return (
			<aside className={shellClass}>
				<div className='flex items-center justify-between border-b border-border px-3 py-2'>
					<span className='text-xs font-semibold'>
						{messages.right.domainSettings}: {selectedDomain}
					</span>
					<Button variant='ghost' size='icon-xs' onClick={toggleRightSidebar}>
						<PanelRightClose className='size-3.5' />
					</Button>
				</div>
				<ScrollArea className='flex-1 p-3'>
					<ConfigEditor
						layer='domain'
						settings={domainCfg}
						onSave={(settings) =>
							persistDomainSettings(selectedDomain, settings)
						}
					/>
				</ScrollArea>
			</aside>
		);
	}

	if (selectedNodeIds.length > 1) {
		return (
			<aside className={shellClass}>
				<div className='flex items-center justify-between border-b border-border px-3 py-2'>
					<p className='text-xs font-semibold'>
						{messages.right.multiSelectCount(selectedNodeIds.length)}
					</p>
					<Button variant='ghost' size='icon-xs' onClick={toggleRightSidebar}>
						<PanelRightClose className='size-3.5' />
					</Button>
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
								{r.markdown?.slice(0, 200) ?? '—'}
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
					</div>
					<Button variant='ghost' size='icon-xs' onClick={toggleRightSidebar}>
						<PanelRightClose className='size-3.5' />
					</Button>
				</div>
				{node.status === 'error' && node.lastError && (
					<Alert variant='destructive' className='m-2 text-xs'>
						{messages.error.nodeFailed}: {node.lastError}
					</Alert>
				)}
				<Tabs
					value={tab}
					onValueChange={(v) => setTab(v as ContentFormat)}
					className='flex min-h-0 flex-1 flex-col px-3'
				>
					<TabsList>
						{formats.map((f) => (
							<TabsTrigger key={f} value={f}>
								{f}
							</TabsTrigger>
						))}
					</TabsList>
					<ScrollArea className='flex-1 pb-3'>
						{formats.map((f) => (
							<TabsContent key={f} value={f}>
								<NodeFormatContent
									format={f}
									result={resultForDisplay ?? node.lastResult}
								/>
							</TabsContent>
						))}
					</ScrollArea>
				</Tabs>
			</aside>
		);
	}

	return (
		<aside className={shellClass}>
			<div className='flex items-center justify-between border-b border-border px-3 py-2 text-xs font-semibold'>
				{messages.right.runSummary}
				<Button variant='ghost' size='icon-xs' onClick={toggleRightSidebar}>
					<PanelRightClose className='size-3.5' />
				</Button>
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
				{runHistory.length === 0 ? (
					<p className='text-xs text-muted-foreground'>
						{messages.right.noSelection}
					</p>
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
							</div>
						))}
					</div>
				)}
			</ScrollArea>
		</aside>
	);
}

/** @deprecated AppShell 内の Panel でラップするため RightSidebarContent を使用 */
export const RightSidebar = RightSidebarContent;

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
			<pre className='whitespace-pre-wrap text-xs'>
				{result.markdown ?? '—'}
			</pre>
		);
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
	return (
		<p className='text-xs text-muted-foreground'>
			{messages.right.formatUnsupported(format)}
		</p>
	);
}
