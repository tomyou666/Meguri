import { Copy, Maximize2, Pencil, Settings } from 'lucide-react';
import { useState } from 'react';
import { scraperPort } from '@/adapters';
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
import { cn } from '@/lib/utils';
import { useAppStore } from '@/stores/appStore';
import type { ContentFormat } from '@/types/config';
import type { CrawlResultPreview } from '@/types/crawl';
import type { GraphNode } from '@/types/graph';

type NodeResultPanelProps = {
	node?: GraphNode;
	formats: ContentFormat[];
	result: CrawlResultPreview | null;
	readonly?: boolean;
	panelMode?: 'sidebar' | 'maximized';
	workspaceId?: string;
	nodeId?: string;
	initialTab?: ContentFormat;
	initialMarkdownView?: 'source' | 'preview';
	onResultChange?: (result: CrawlResultPreview) => void;
	className?: string;
};

export function NodeResultPanel({
	node,
	formats,
	result,
	readonly = false,
	panelMode = 'sidebar',
	workspaceId,
	nodeId,
	initialTab,
	initialMarkdownView = 'preview',
	onResultChange,
	className,
}: NodeResultPanelProps) {
	const persistNodeSettings = useAppStore((s) => s.persistNodeSettings);
	const updateNodeResult = useAppStore((s) => s.updateNodeResult);
	const showMaximizedNodeResult = useAppStore((s) => s.showMaximizedNodeResult);
	const activeWorkspace = useAppStore((s) => s.getActiveWorkspace());

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

	const resolvedNodeId = nodeId ?? node?.id;
	const resolvedWorkspaceId = workspaceId ?? activeWorkspace?.id;
	const displayResult = result ?? node?.lastResult ?? null;
	const showPdfTab = isPdfResourceResult(displayResult);
	const isMaximized = panelMode === 'maximized';

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
		if (!resolvedNodeId || !resolvedWorkspaceId) return;
		const patch = updatePatchForFormat(tab, draft);
		if (!patch) return;
		setSaving(true);
		let saved = false;
		let updated: CrawlResultPreview | null = null;
		if (isMaximized) {
			try {
				updated = await scraperPort.updateNodeResult(
					resolvedWorkspaceId,
					resolvedNodeId,
					patch,
				);
				saved = !!updated;
				if (!saved) {
					notifyError(messages.right.updateFailed);
				}
			} catch (err) {
				notifyError(messages.right.updateFailed, {
					description: err instanceof Error ? err.message : String(err),
				});
			}
		} else {
			saved = await updateNodeResult(resolvedNodeId, patch);
		}
		setSaving(false);
		if (saved) {
			if (updated) onResultChange?.(updated);
			setEditing(false);
			setDraft('');
			if (isMaximized) {
				notifySuccess(messages.right.updateSaved);
			}
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
		if (readonly || isMaximized || !displayResult || !node) return;
		if (!resolvedWorkspaceId) return;
		void showMaximizedNodeResult({
			title: node.urlNormalized,
			workspaceId: resolvedWorkspaceId,
			nodeId: node.id,
			activeFormat: tab,
			markdownView,
			formats,
			result: displayResult,
		});
	};

	const isEditingContent = !readonly && editing;

	return (
		<Tabs
			value={tab}
			onValueChange={(v) => {
				setTab(v as ContentFormat);
				setShowNodeSettings(false);
				setEditing(false);
				setDraft('');
			}}
			className={cn('flex min-h-0 flex-1 flex-col px-3', className)}
		>
			<div className='shrink-0 space-y-1'>
				<TabsList className='w-full overflow-x-auto'>
					{formats.map((f) => (
						<TabsTrigger key={f} value={f}>
							{previewTabLabel(f)}
						</TabsTrigger>
					))}
				</TabsList>
				<div className='flex shrink-0 justify-end gap-1'>
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
							{!readonly && !isMaximized && (
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
					{!readonly && !isMaximized && node && (
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
			</div>
			{!readonly && showNodeSettings && node ? (
				<ScrollArea className='min-h-0 flex-1 py-3'>
					<ConfigEditor
						layer='node'
						settings={node.nodeSettings ?? {}}
						compact
						showPdfTab={showPdfTab}
						showRequestTab={false}
						showCrawlTab={false}
						onSave={(settings) => persistNodeSettings(node.id, settings)}
					/>
				</ScrollArea>
			) : isEditingContent ? (
				<div className='flex min-h-0 flex-1 flex-col overflow-hidden py-3'>
					{formats.map((f) => (
						<TabsContent
							key={f}
							value={f}
							className='flex min-h-0 flex-1 flex-col'
						>
							<NodeFormatContent
								format={f}
								result={displayResult}
								editing={editing && f === tab}
								saving={saving}
								markdownView={markdownView}
								draft={draft}
								onDraftChange={setDraft}
								onSave={() => void saveEdit()}
								onCancel={cancelEdit}
								onMarkdownViewChange={setMarkdownView}
							/>
						</TabsContent>
					))}
				</div>
			) : (
				<ScrollArea className='min-h-0 flex-1 py-3'>
					{formats.map((f) => (
						<TabsContent key={f} value={f}>
							<NodeFormatContent
								format={f}
								result={displayResult}
								editing={false}
								saving={saving}
								markdownView={markdownView}
								draft={draft}
								onDraftChange={setDraft}
								onSave={() => void saveEdit()}
								onCancel={cancelEdit}
								onMarkdownViewChange={setMarkdownView}
							/>
						</TabsContent>
					))}
				</ScrollArea>
			)}
		</Tabs>
	);
}
