import ReactMarkdown from 'react-markdown';
import { EditableTextResult } from '@/components/layout/node-result/EditableTextResult';
import { Button } from '@/components/ui/button';
import { messages } from '@/i18n/messages';
import { PREVIEW_BASE_URL_ATTR } from '@/lib/externalLinkDelegation';

type MarkdownResultViewProps = {
	markdown: string;
	previewBaseUrl: string;
	view: 'source' | 'preview';
	editing: boolean;
	saving?: boolean;
	draft: string;
	onDraftChange: (value: string) => void;
	onSave: () => void;
	onCancel: () => void;
};

export function MarkdownResultView({
	markdown,
	previewBaseUrl,
	view,
	editing,
	saving = false,
	draft,
	onDraftChange,
	onSave,
	onCancel,
}: MarkdownResultViewProps) {
	if (editing || view === 'source') {
		return (
			<EditableTextResult
				value={editing ? draft : markdown}
				editing={editing}
				saving={saving}
				onChange={onDraftChange}
				onSave={onSave}
				onCancel={onCancel}
			/>
		);
	}

	if (!markdown) {
		return <p className='text-xs text-muted-foreground'>—</p>;
	}

	return (
		<div
			className='markdown-preview space-y-2 text-xs leading-relaxed [&_h1]:text-base [&_h1]:font-semibold [&_h2]:text-sm [&_h2]:font-semibold [&_h3]:font-medium [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_pre]:overflow-x-auto [&_pre]:rounded-md [&_pre]:bg-muted [&_pre]:p-2 [&_code]:font-mono [&_code]:text-[11px] [&_a]:text-primary [&_a]:underline [&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-3 [&_blockquote]:text-muted-foreground'
			{...{ [PREVIEW_BASE_URL_ATTR]: previewBaseUrl }}
		>
			<ReactMarkdown>{markdown}</ReactMarkdown>
		</div>
	);
}

type MarkdownViewToggleProps = {
	view: 'source' | 'preview';
	editing: boolean;
	onViewChange: (view: 'source' | 'preview') => void;
};

export function MarkdownViewToggle({
	view,
	editing,
	onViewChange,
}: MarkdownViewToggleProps) {
	if (editing) return null;

	return (
		<div className='mb-2 flex gap-1'>
			<Button
				size='xs'
				variant={view === 'source' ? 'secondary' : 'ghost'}
				onClick={() => onViewChange('source')}
			>
				{messages.right.source}
			</Button>
			<Button
				size='xs'
				variant={view === 'preview' ? 'secondary' : 'ghost'}
				onClick={() => onViewChange('preview')}
			>
				{messages.right.previewLabel}
			</Button>
		</div>
	);
}
