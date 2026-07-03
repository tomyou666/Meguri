import type { ReactNode } from 'react';
import { EditableTextResult } from '@/components/layout/node-result/EditableTextResult';
import {
	MarkdownResultView,
	MarkdownViewToggle,
} from '@/components/layout/node-result/MarkdownResultView';
import { ExternalLink } from '@/components/ui/external-link';
import { messages } from '@/i18n/messages';
import type { ContentFormat } from '@/types/config';
import type { CrawlResultPreview } from '@/types/crawl';

type NodeFormatContentProps = {
	format: ContentFormat;
	result?: CrawlResultPreview;
	editing: boolean;
	saving?: boolean;
	markdownView: 'source' | 'preview';
	draft: string;
	onDraftChange: (value: string) => void;
	onSave: () => void;
	onCancel: () => void;
	onMarkdownViewChange: (view: 'source' | 'preview') => void;
};

export function NodeFormatContent({
	format,
	result,
	editing,
	saving = false,
	markdownView,
	draft,
	onDraftChange,
	onSave,
	onCancel,
	onMarkdownViewChange,
}: NodeFormatContentProps) {
	const editLayout = (children: ReactNode) =>
		editing ? (
			<div className='flex min-h-0 flex-1 flex-col'>{children}</div>
		) : (
			children
		);

	if (!result) {
		return (
			<p className='text-xs text-muted-foreground'>
				{messages.right.noResultApi}
			</p>
		);
	}

	if (format === 'markdown') {
		return editLayout(
			<>
				<MarkdownViewToggle
					view={markdownView}
					editing={editing}
					onViewChange={onMarkdownViewChange}
				/>
				<MarkdownResultView
					markdown={result.markdown ?? ''}
					previewBaseUrl={result.url}
					view={markdownView}
					editing={editing}
					saving={saving}
					draft={draft}
					onDraftChange={onDraftChange}
					onSave={onSave}
					onCancel={onCancel}
				/>
			</>,
		);
	}

	if (format === 'html' || format === 'raw_html' || format === 'json') {
		const value =
			format === 'html'
				? (result.html ?? '')
				: format === 'raw_html'
					? (result.raw_html ?? '')
					: (result.json ?? '');
		return editLayout(
			<EditableTextResult
				value={editing ? draft : value}
				editing={editing}
				saving={saving}
				onChange={onDraftChange}
				onSave={onSave}
				onCancel={onCancel}
			/>,
		);
	}

	if (format === 'links') {
		return (
			<ul className='list-inside list-disc text-xs'>
				{(result.links ?? []).map((l) => (
					<li key={l} className='truncate'>
						<ExternalLink href={l} className='text-xs'>
							{l}
						</ExternalLink>
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
