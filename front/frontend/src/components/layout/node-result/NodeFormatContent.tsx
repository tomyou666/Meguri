import { EditableTextResult } from '@/components/layout/node-result/EditableTextResult';
import {
	MarkdownResultView,
	MarkdownViewToggle,
} from '@/components/layout/node-result/MarkdownResultView';
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
	if (!result) {
		return (
			<p className='text-xs text-muted-foreground'>
				{messages.right.noResultApi}
			</p>
		);
	}

	if (format === 'markdown') {
		return (
			<>
				<MarkdownViewToggle
					view={markdownView}
					editing={editing}
					onViewChange={onMarkdownViewChange}
				/>
				<MarkdownResultView
					markdown={result.markdown ?? ''}
					view={markdownView}
					editing={editing}
					saving={saving}
					draft={draft}
					onDraftChange={onDraftChange}
					onSave={onSave}
					onCancel={onCancel}
				/>
			</>
		);
	}

	if (format === 'html' || format === 'raw_html' || format === 'json') {
		const value =
			format === 'html'
				? (result.html ?? '')
				: format === 'raw_html'
					? (result.raw_html ?? '')
					: (result.json ?? '');
		return (
			<EditableTextResult
				value={editing ? draft : value}
				editing={editing}
				saving={saving}
				onChange={onDraftChange}
				onSave={onSave}
				onCancel={onCancel}
			/>
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

	return <p className='text-xs text-muted-foreground'>—</p>;
}
