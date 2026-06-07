import { ChevronDown, Pause, Play, Square } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { messages } from '@/i18n/messages';
import { cn } from '@/lib/utils';
import { useAppStore } from '@/stores/appStore';
import type { ContentFormat } from '@/types/config';
import type { RunMode } from '@/types/crawl';

const ALL_FORMATS: ContentFormat[] = [
	'markdown',
	'html',
	'raw_html',
	'json',
	'links',
	'metadata',
];

const MODE_LABELS: Record<RunMode, string> = {
	1: messages.control.mode1,
	2: messages.control.mode2,
	3: messages.control.mode3,
};

export function ControlBar() {
	const runMode = useAppStore((s) => s.runMode);
	const setRunMode = useAppStore((s) => s.setRunMode);
	const rescrapeExisting = useAppStore((s) => s.rescrapeExisting);
	const setRescrapeExisting = useAppStore((s) => s.setRescrapeExisting);
	const crawlStatus = useAppStore((s) => s.crawlStatus);
	const startCrawl = useAppStore((s) => s.startCrawl);
	const pauseCrawl = useAppStore((s) => s.pauseCrawl);
	const resumeCrawl = useAppStore((s) => s.resumeCrawl);
	const stopCrawl = useAppStore((s) => s.stopCrawl);
	const ws = useAppStore((s) => s.getActiveWorkspace());
	const setWorkspaceFormats = useAppStore((s) => s.setWorkspaceFormats);
	const bulkScrapeSelected = useAppStore((s) => s.bulkScrapeSelected);
	const mergeAllResults = useAppStore((s) => s.mergeAllResults);
	const mergeSelectedResults = useAppStore((s) => s.mergeSelectedResults);
	const selectedNodeIds = useAppStore((s) => s.selectedNodeIds);
	const [modeMenuOpen, setModeMenuOpen] = useState(false);

	const formats = ws?.settings.content?.formats ?? ['markdown', 'links'];

	const toggleFormat = (f: ContentFormat) => {
		const next = formats.includes(f)
			? formats.filter((x) => x !== f)
			: [...formats, f];
		if (next.length === 0) return;
		setWorkspaceFormats({ formats: next });
	};

	const isRunning = crawlStatus === 'running';
	const isPaused = crawlStatus === 'paused';

	return (
		<div className='flex h-12 items-center justify-between gap-2 border-b border-border bg-card px-3'>
			<div className='flex shrink-0 items-center gap-2 text-sm font-semibold'>
				<span className='text-primary'>{messages.appName}</span>
				<span className='text-xs font-normal text-muted-foreground'>
					v{messages.version}
				</span>
			</div>

			<div className='flex min-w-0 flex-1 flex-wrap items-center justify-center gap-2'>
				<div className='relative'>
					<div className='flex'>
						<Button
							size='sm'
							disabled={!ws || isRunning}
							onClick={() => startCrawl()}
							className='rounded-r-none'
						>
							<Play className='size-3.5' />
							{messages.control.play}
						</Button>
						<Button
							size='sm'
							variant='outline'
							className='rounded-l-none border-l-0 px-1.5'
							onClick={() => setModeMenuOpen((o) => !o)}
						>
							<ChevronDown className='size-3.5' />
						</Button>
					</div>
					{modeMenuOpen && (
						<>
							<button
								type='button'
								aria-label={messages.control.closeModeMenu}
								className='fixed inset-0 z-40'
								onClick={() => setModeMenuOpen(false)}
							/>
							<div className='absolute left-0 top-full z-50 mt-1 min-w-56 rounded-lg border border-border bg-popover py-1 shadow-lg'>
								{([1, 2, 3] as RunMode[]).map((m) => (
									<button
										key={m}
										type='button'
										className={cn(
											'block w-full px-3 py-1.5 text-left text-xs hover:bg-muted',
											runMode === m && 'bg-muted font-medium',
										)}
										onClick={() => {
											setRunMode(m);
											setModeMenuOpen(false);
										}}
									>
										{MODE_LABELS[m]}
									</button>
								))}
							</div>
						</>
					)}
				</div>

				{isRunning && (
					<Button size='sm' variant='outline' onClick={pauseCrawl}>
						<Pause className='size-3.5' />
						{messages.control.pause}
					</Button>
				)}
				{isPaused && (
					<Button size='sm' variant='outline' onClick={resumeCrawl}>
						<Play className='size-3.5' />
						{messages.control.play}
					</Button>
				)}
				{(isRunning || isPaused) && (
					<Button size='sm' variant='destructive' onClick={stopCrawl}>
						<Square className='size-3.5' />
						{messages.control.stop}
					</Button>
				)}

				<div className='flex items-center gap-1.5'>
					<Checkbox
						id='rescrape-existing'
						checked={rescrapeExisting}
						disabled={isRunning || isPaused || !ws}
						onCheckedChange={(checked) => setRescrapeExisting(checked === true)}
					/>
					<Label
						htmlFor='rescrape-existing'
						className='cursor-pointer text-xs font-normal text-muted-foreground'
					>
						{messages.control.rescrapeExisting}
					</Label>
				</div>

				<span className='text-muted-foreground'>|</span>
				<span className='text-xs text-muted-foreground'>
					{messages.control.formats}
				</span>
				{ALL_FORMATS.map((f) => (
					<Button
						key={f}
						size='xs'
						variant={formats.includes(f) ? 'default' : 'outline'}
						onClick={() => toggleFormat(f)}
					>
						{f}
					</Button>
				))}

				{selectedNodeIds.length > 1 && (
					<>
						<span className='text-muted-foreground'>|</span>
						<Button
							size='xs'
							variant='outline'
							disabled={crawlStatus !== 'idle'}
							onClick={() => bulkScrapeSelected()}
						>
							{messages.control.bulkScrapeSelected}
						</Button>
					</>
				)}
			</div>

			<div className='flex shrink-0 items-center gap-2'>
				<Button
					variant='outline'
					size='xs'
					disabled={!ws}
					onClick={() => mergeAllResults()}
				>
					{messages.menu.mergeAll}
				</Button>
				<Button
					variant='outline'
					size='xs'
					disabled={selectedNodeIds.length === 0}
					onClick={() => mergeSelectedResults()}
				>
					{messages.menu.mergeSelected}
				</Button>
			</div>
		</div>
	);
}
