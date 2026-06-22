import { Copy, Maximize2, Pencil, Settings } from 'lucide-react';
import { useState } from 'react';
import { NodeFormatContent } from '@/components/layout/node-result/NodeFormatContent';
import { ConfigEditor } from '@/components/settings/ConfigEditor';
import { ActionTooltip } from '@/components/ui/action-tooltip';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { messages } from '@/i18n/messages';
import { isPdfResourceResult } from '@/lib/crawlResultUtils';
import { notifyError, notifySuccess } from '@/lib/notify';
import { previewTabLabel } from '@/lib/previewFormats';
import {
	editableValueForFormat,
	isEditableFormat,
	resultTextForFormat,
	updatePatchForFormat,
} from '@/lib/resultFormatText';
import { useAppStore } from '@/stores/appStore';
import type { ContentFormat } from '@/types/config';
import type { CrawlResultPreview } from '@/types/crawl';
import type { GraphNode } from '@/types/graph';

type NodeResultPanelProps = {
	node?: GraphNode;
	formats: ContentFormat[];
	result: CrawlResultPreview | null;
	readonly?: boolean;
	initialTab?: ContentFormat;
	initialMarkdownView?: 'source' | 'preview';
};

export function NodeResultPanel({
	node,
	formats,
	result,
	readonly = false,
	initialTab,
	initialMarkdownView = 'preview',
}: NodeResultPanelProps) {
	const persistNodeSettings = useAppStore((s) => s.persistNodeSettings);
	const updateNodeResult = useAppStore((s) => s.updateNodeResult);
	const showMaximizedNodeResult = useAppStore((s) => s.showMaximizedNodeResult);

	const [tab, setTab] = useState<ContentFormat>(
		initialTab ?? formats[0] ?? 'markdown',
	);
	const [showNodeSettings, setShowNodeSettings] = useState(false);
	const [editing, setEditing] = useState(false);
	const [saving, setSaving] = useState(false);
	const [draft, setDraft] = useState('');
	const [markdownView, setMarkdownView] = useState<'source' | 'preview'>(
		initialMarkdownView,
	);

	const displayResult = result ?? node?.lastResult ?? null;
	const showPdfTab = isPdfResourceResult(displayResult);

	const beginEdit = () => {
		if (readonly || !displayResult || !isEditableFormat(tab)) return;
		setDraft(editableValueForFormat(displayResult, tab));
		setEditing(true);
		if (tab === 'markdown') {
			setMarkdownView('source');
		}
	};

	const cancelEdit = () => {
		setEditing(false);
		setDraft('');
	};

	const saveEdit = async () => {
		if (!node) return;
		const patch = updatePatchForFormat(tab, draft);
		if (!patch) return;
		setSaving(true);
		const ok = await updateNodeResult(node.id, patch);
		setSaving(false);
		if (ok) {
			setEditing(false);
			setDraft('');
		}
	};

	const copyCurrentTab = async () => {
		if (!displayResult) return;
		const text = resultTextForFormat(displayResult, tab);
		try {
			await navigator.clipboard.writeText(text);
			notifySuccess(messages.right.copied);
		} catch (err) {
			notifyError(messages.right.copyFailed, {
				description: err instanceof Error ? err.message : String(err),
			});
		}
	};

	const maximizePanel = () => {
		if (readonly || !displayResult || !node) return;
		void showMaximizedNodeResult({
			title: node.urlNormalized,
			activeFormat: tab,
			markdownView,
			formats,
			result: displayResult,
		});
	};

	return (
		<Tabs
			value={tab}
			onValueChange={(v) => {
				setTab(v as ContentFormat);
				setShowNodeSettings(false);
				setEditing(false);
				setDraft('');
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
				{!showNodeSettings && displayResult && (
					<>
						<ActionTooltip label={messages.right.copy}>
							<Button
								variant='ghost'
								size='icon-xs'
								aria-label={messages.right.copy}
								onClick={() => void copyCurrentTab()}
							>
								<Copy className='size-3.5' />
							</Button>
						</ActionTooltip>
						{!readonly && (
							<ActionTooltip label={messages.right.maximize}>
								<Button
									variant='ghost'
									size='icon-xs'
									aria-label={messages.right.maximize}
									onClick={maximizePanel}
								>
									<Maximize2 className='size-3.5' />
								</Button>
							</ActionTooltip>
						)}
						{!readonly && isEditableFormat(tab) && !editing && (
							<ActionTooltip label={messages.right.edit}>
								<Button
									variant='ghost'
									size='icon-xs'
									aria-label={messages.right.edit}
									onClick={beginEdit}
								>
									<Pencil className='size-3.5' />
								</Button>
							</ActionTooltip>
						)}
					</>
				)}
				{!readonly && node && (
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
				)}
			</div>
			<ScrollArea className='flex-1 py-3'>
				{!readonly && showNodeSettings && node ? (
					<ConfigEditor
						layer='node'
						settings={node.nodeSettings ?? {}}
						compact
						showPdfTab={showPdfTab}
						showRequestTab={false}
						showCrawlTab={false}
						onSave={(settings) => persistNodeSettings(node.id, settings)}
					/>
				) : (
					formats.map((f) => (
						<TabsContent key={f} value={f}>
							<NodeFormatContent
								format={f}
								result={displayResult}
								editing={!readonly && editing && f === tab}
								saving={saving}
								markdownView={markdownView}
								draft={draft}
								onDraftChange={setDraft}
								onSave={() => void saveEdit()}
								onCancel={cancelEdit}
								onMarkdownViewChange={setMarkdownView}
							/>
						</TabsContent>
					))
				)}
			</ScrollArea>
		</Tabs>
	);
}
