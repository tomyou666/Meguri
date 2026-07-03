import { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { MarkdownViewToggle } from '@/components/layout/node-result/MarkdownResultView';
import { ScrollArea } from '@/components/ui/scroll-area';
import { messages } from '@/i18n/messages';
import type { ExportFormat, ExportPreviewSection } from '@/lib/exportTree';
import { PREVIEW_BASE_URL_ATTR } from '@/lib/externalLinkDelegation';

const markdownPreviewClassName =
	'markdown-preview space-y-2 text-xs leading-relaxed [&_h1]:text-base [&_h1]:font-semibold [&_h2]:text-sm [&_h2]:font-semibold [&_h3]:font-medium [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_pre]:overflow-x-auto [&_pre]:rounded-md [&_pre]:bg-muted [&_pre]:p-2 [&_code]:font-mono [&_code]:text-[11px] [&_a]:text-primary [&_a]:underline [&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-3 [&_blockquote]:text-muted-foreground';

type ExportPreviewPaneProps = {
	content: string | null;
	sections: ExportPreviewSection[] | null;
	format: ExportFormat;
	loading: boolean;
};

export function ExportPreviewPane({
	content,
	sections,
	format,
	loading,
}: ExportPreviewPaneProps) {
	const [markdownView, setMarkdownView] = useState<'source' | 'preview'>(
		'preview',
	);

	const hasRenderablePreview =
		(sections?.length ?? 0) > 0 || (content?.length ?? 0) > 0;

	return (
		<main className='flex h-full min-w-0 flex-col bg-background'>
			<div className='border-b border-border px-3 py-2 text-xs font-semibold'>
				{messages.export.previewTitle}
			</div>
			{format === 'markdown' && hasRenderablePreview && !loading && (
				<div className='border-b border-border px-3 py-2'>
					<MarkdownViewToggle
						view={markdownView}
						editing={false}
						onViewChange={setMarkdownView}
					/>
				</div>
			)}
			<ScrollArea className='min-h-0 flex-1'>
				{loading ? (
					<p className='p-4 text-sm text-muted-foreground'>
						{messages.export.previewLoading}
					</p>
				) : hasRenderablePreview ? (
					format === 'html' ? (
						<div className='space-y-4 p-4'>
							{(sections ?? []).map((section) => (
								<div
									key={section.id}
									className='prose prose-sm dark:prose-invert max-w-none'
									{...{ [PREVIEW_BASE_URL_ATTR]: section.baseUrl }}
									// biome-ignore lint/security/noDangerouslySetInnerHtml: export preview of scraped HTML
									dangerouslySetInnerHTML={{ __html: section.body }}
								/>
							))}
						</div>
					) : markdownView === 'source' ? (
						<pre className='whitespace-pre-wrap p-4 font-mono text-xs'>
							{content}
						</pre>
					) : (
						<div className='space-y-4 p-4'>
							{(sections ?? []).map((section) => (
								<div
									key={section.id}
									className={markdownPreviewClassName}
									{...{ [PREVIEW_BASE_URL_ATTR]: section.baseUrl }}
								>
									<ReactMarkdown>{section.body}</ReactMarkdown>
								</div>
							))}
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
