import ReactMarkdown from 'react-markdown';
import { ScrollArea } from '@/components/ui/scroll-area';
import { messages } from '@/i18n/messages';
import type { ExportFormat } from '@/lib/exportTree';

type ExportPreviewPaneProps = {
	content: string | null;
	format: ExportFormat;
	loading: boolean;
};

export function ExportPreviewPane({
	content,
	format,
	loading,
}: ExportPreviewPaneProps) {
	return (
		<main className='flex h-full min-w-0 flex-col bg-background'>
			<div className='border-b border-border px-3 py-2 text-xs font-semibold'>
				{messages.export.previewTitle}
			</div>
			<ScrollArea className='min-h-0 flex-1'>
				{loading ? (
					<p className='p-4 text-sm text-muted-foreground'>
						{messages.export.previewLoading}
					</p>
				) : content ? (
					format === 'html' ? (
						<div
							className='prose prose-sm dark:prose-invert max-w-none p-4'
							// biome-ignore lint/security/noDangerouslySetInnerHtml: export preview of scraped HTML
							dangerouslySetInnerHTML={{ __html: content }}
						/>
					) : (
						<div className='prose prose-sm dark:prose-invert max-w-none p-4'>
							<ReactMarkdown>{content}</ReactMarkdown>
						</div>
					)
				) : (
					<p className='p-4 text-sm text-muted-foreground'>
						{messages.export.previewEmpty}
					</p>
				)}
			</ScrollArea>
		</main>
	);
}
