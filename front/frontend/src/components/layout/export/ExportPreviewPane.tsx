import { useVirtualizer } from '@tanstack/react-virtual';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { MarkdownViewToggle } from '@/components/layout/node-result/MarkdownResultView';
import { messages } from '@/i18n/messages';
import {
	computeBodyWindow,
	type ExportPreviewRowBody,
	type ExportPreviewRowMeta,
	estimateExportRowHeight,
	pruneBodyMap,
	shouldFetchBodies,
} from '@/lib/exportPreviewVirtual';
import {
	type ExportFlatNode,
	type ExportFormat,
	type ExportMergeSettings,
	resolveExportSeparator,
} from '@/lib/exportTree';
import { PREVIEW_BASE_URL_ATTR } from '@/lib/externalLinkDelegation';
import { cn } from '@/lib/utils';

const markdownPreviewClassName =
	'markdown-preview space-y-2 text-xs leading-relaxed [&_h1]:text-base [&_h1]:font-semibold [&_h2]:text-sm [&_h2]:font-semibold [&_h3]:font-medium [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_pre]:overflow-x-auto [&_pre]:rounded-md [&_pre]:bg-muted [&_pre]:p-2 [&_code]:font-mono [&_code]:text-[11px] [&_a]:text-primary [&_a]:underline [&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-3 [&_blockquote]:text-muted-foreground';

type ExportPreviewPaneProps = {
	rows: ExportPreviewRowMeta[];
	bodies: Map<number, ExportPreviewRowBody>;
	format: ExportFormat;
	settings: ExportMergeSettings;
	flatData: ExportFlatNode[];
	loading: boolean;
	disabled: boolean;
	sessionKey: number;
	onFetchBodies: (startIndex: number, endIndex: number) => void;
	onMarkdownViewChange?: (view: 'source' | 'preview') => void;
	/** 高さキャッシュ破棄のトリガ（値の変化で measure し直す）。 */
	heightResetKey?: number;
	scrollToIndexRequest?: number | null;
};

export function ExportPreviewPane({
	rows,
	bodies,
	format,
	settings,
	flatData,
	loading,
	disabled,
	sessionKey,
	onFetchBodies,
	onMarkdownViewChange,
	heightResetKey = 0,
	scrollToIndexRequest = null,
}: ExportPreviewPaneProps) {
	const [markdownView, setMarkdownView] = useState<'source' | 'preview'>(
		'preview',
	);
	const parentRef = useRef<HTMLDivElement>(null);
	const heldWindowRef = useRef<{ start: number; end: number } | null>(null);
	const metaById = useMemo(
		() => new Map(flatData.map((n) => [n.id, n])),
		[flatData],
	);

	const estimateSize = useCallback(
		(index: number) => {
			const row = rows[index];
			if (!row) return 64;
			return estimateExportRowHeight(row.byteLength, row.isFirst);
		},
		[rows],
	);

	const virtualizer = useVirtualizer({
		count: rows.length,
		getScrollElement: () => parentRef.current,
		estimateSize,
		overscan: 6,
		getItemKey: (index) => rows[index]?.id ?? index,
	});

	useEffect(() => {
		virtualizer.measure();
		// heightResetKey の変化で測り直す。
		void heightResetKey;
	}, [heightResetKey, virtualizer]);

	useEffect(() => {
		if (scrollToIndexRequest == null) return;
		virtualizer.scrollToIndex(scrollToIndexRequest, { align: 'start' });
	}, [scrollToIndexRequest, virtualizer]);

	const virtualItems = virtualizer.getVirtualItems();

	useEffect(() => {
		if (rows.length === 0 || loading) return;
		const first = virtualItems[0];
		const last = virtualItems[virtualItems.length - 1];
		if (!first || !last) {
			const initialEnd = Math.min(20, rows.length);
			const initial = computeBodyWindow(0, initialEnd, rows.length, 20);
			if (shouldFetchBodies(initial, heldWindowRef.current)) {
				heldWindowRef.current = initial;
				onFetchBodies(initial.start, initial.end);
			}
			return;
		}
		const viewportSpan = Math.max(1, last.index - first.index + 1);
		const needed = computeBodyWindow(
			first.index,
			last.index + 1,
			rows.length,
			viewportSpan,
		);
		if (!shouldFetchBodies(needed, heldWindowRef.current)) return;
		heldWindowRef.current = needed;
		onFetchBodies(needed.start, needed.end);
	}, [virtualItems, rows.length, loading, onFetchBodies]);

	useEffect(() => {
		void sessionKey;
		heldWindowRef.current = null;
	}, [sessionKey]);

	const handleViewChange = (view: 'source' | 'preview') => {
		setMarkdownView(view);
		onMarkdownViewChange?.(view);
	};

	const hasRows = rows.length > 0;

	return (
		<main className='flex h-full min-w-0 flex-col bg-background'>
			<div className='border-border border-b px-3 py-2 font-semibold text-xs'>
				{messages.export.previewTitle}
			</div>
			{format === 'markdown' && hasRows && !loading && (
				<div className='border-border border-b px-3 py-2'>
					<MarkdownViewToggle
						view={markdownView}
						editing={false}
						onViewChange={handleViewChange}
					/>
				</div>
			)}
			<div
				ref={parentRef}
				className={cn(
					'min-h-0 flex-1 overflow-auto',
					disabled && 'pointer-events-none opacity-60',
				)}
			>
				{loading ? (
					<p className='p-4 text-muted-foreground text-sm'>
						{messages.export.previewLoading}
					</p>
				) : hasRows ? (
					<div
						className='relative w-full'
						style={{ height: virtualizer.getTotalSize() }}
					>
						{virtualItems.map((item) => {
							const row = rows[item.index];
							if (!row) return null;
							const body = bodies.get(item.index);
							const meta = metaById.get(row.nodeId);
							const headingText =
								settings.headingField === 'label'
									? (meta?.label ?? '')
									: (meta?.urlNormalized ?? '');
							const showHeading =
								row.isFirst && settings.includeHeading && headingText !== '';
							const showSeparator = row.isFirst && item.index > 0;
							const separator = showSeparator
								? resolveExportSeparator(settings.separator, format)
								: '';
							const baseUrl = body?.url ?? meta?.urlNormalized ?? '';
							const headingMarkdown = showHeading
								? `## ${headingText}\n\n`
								: '';

							return (
								<div
									key={item.key}
									data-index={item.index}
									ref={virtualizer.measureElement}
									className='absolute top-0 left-0 w-full px-4 py-2'
									style={{ transform: `translateY(${item.start}px)` }}
								>
									{showSeparator && format === 'markdown' && (
										<pre className='whitespace-pre-wrap py-2 font-mono text-muted-foreground text-xs'>
											{separator}
										</pre>
									)}
									{showSeparator && format === 'html' && (
										<div
											className='py-2 text-muted-foreground text-xs'
											// biome-ignore lint/security/noDangerouslySetInnerHtml: escaped separator for HTML export
											dangerouslySetInnerHTML={{ __html: separator }}
										/>
									)}
									{!body ? (
										<div
											className='rounded-md bg-muted/40'
											style={{
												height: estimateExportRowHeight(
													row.byteLength,
													row.isFirst,
												),
											}}
										/>
									) : format === 'html' ? (
										<div className='flex flex-col gap-2'>
											{showHeading && (
												<h2 className='font-semibold text-sm'>{headingText}</h2>
											)}
											<div
												className='prose prose-sm dark:prose-invert max-w-none'
												{...{ [PREVIEW_BASE_URL_ATTR]: baseUrl }}
												// biome-ignore lint/security/noDangerouslySetInnerHtml: export preview of scraped HTML
												dangerouslySetInnerHTML={{ __html: body.body }}
											/>
										</div>
									) : markdownView === 'source' ? (
										<pre className='whitespace-pre-wrap font-mono text-xs'>
											{headingMarkdown}
											{body.body}
										</pre>
									) : (
										<div
											className={markdownPreviewClassName}
											{...{ [PREVIEW_BASE_URL_ATTR]: baseUrl }}
										>
											<ReactMarkdown>
												{`${headingMarkdown}${body.body}`}
											</ReactMarkdown>
										</div>
									)}
								</div>
							);
						})}
					</div>
				) : (
					<p className='p-4 text-muted-foreground text-sm'>
						{messages.export.previewEmpty}
					</p>
				)}
			</div>
		</main>
	);
}

/** 本文 Map を窓で刈り込むヘルパー（親から利用）。 */
export function trimExportBodies(
	bodies: Map<number, ExportPreviewRowBody>,
	start: number,
	end: number,
): Map<number, ExportPreviewRowBody> {
	return pruneBodyMap(bodies, { start, end });
}
