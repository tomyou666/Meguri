import { Events } from '@wailsio/runtime';
import { useCallback, useEffect, useState } from 'react';
import { Group, Panel, Separator } from 'react-resizable-panels';
import { scraperPort } from '@/adapters';
import { ExportOrderSidebar } from '@/components/layout/export/ExportOrderSidebar';
import { ExportPreviewPane } from '@/components/layout/export/ExportPreviewPane';
import { ExportSettingsSidebar } from '@/components/layout/export/ExportSettingsSidebar';
import { Toaster } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import { messages } from '@/i18n/messages';
import {
	buildExportPreview,
	buildExportPreviewSections,
	buildInitialFlatTree,
	buildSplitExportFiles,
	DEFAULT_EXPORT_SETTINGS,
	type ExportFlatNode,
	type ExportMergeSettings,
	type ExportPreviewSection,
	type ExportZipFileEntry,
	initialCheckedIds,
	preorderNodeIds,
} from '@/lib/exportTree';
import { notifyError, notifySuccess } from '@/lib/notify';
import type { ExportSessionSnapshot } from '@/types/adapter';
import type { CrawlResultPreview } from '@/types/crawl';
import type { GraphEdge, GraphNode } from '@/types/graph';

type PreviewSource = {
	orderedIds: string[];
	results: CrawlResultPreview[];
};

const TOPIC_EXPORT_OPEN = 'export:open';

function graphNodesFromSession(
	nodes: ExportSessionSnapshot['nodes'],
): GraphNode[] {
	return nodes.map((n) => ({
		id: n.id,
		urlNormalized: n.urlNormalized,
		label: n.label,
		position: { x: 0, y: 0 },
		nodeSettings: {},
		crawlExclude: false,
		status: n.status as GraphNode['status'],
	}));
}

function graphEdgesFromSession(
	edges: ExportSessionSnapshot['edges'],
): GraphEdge[] {
	return edges.map((e, i) => ({
		id: `e-${i}`,
		source: e.source,
		target: e.target,
	}));
}

function applySession(session: ExportSessionSnapshot) {
	const nodes = graphNodesFromSession(session.nodes);
	const edges = graphEdgesFromSession(session.edges);
	const flat = buildInitialFlatTree(
		nodes,
		edges,
		session.seedUrl,
		session.mode,
		session.selectedNodeIds ?? [],
	);
	return {
		workspaceId: session.workspaceId,
		flatData: flat,
		checkedIds: initialCheckedIds(flat),
	};
}

function snapshotFromEventData(data: unknown): ExportSessionSnapshot | null {
	if (!data || typeof data !== 'object') return null;
	const raw = data as Record<string, unknown>;
	if (typeof raw.workspaceId !== 'string') return null;
	return {
		title: String(raw.title ?? ''),
		workspaceId: raw.workspaceId,
		mode: raw.mode === 'selected' ? 'selected' : 'all',
		seedUrl: String(raw.seedUrl ?? ''),
		nodes: Array.isArray(raw.nodes)
			? raw.nodes.map((n) => {
					const node = n as Record<string, unknown>;
					return {
						id: String(node.id ?? ''),
						urlNormalized: String(node.urlNormalized ?? ''),
						label: String(node.label ?? ''),
						status: String(node.status ?? ''),
					};
				})
			: [],
		edges: Array.isArray(raw.edges)
			? raw.edges.map((e) => {
					const edge = e as Record<string, unknown>;
					return {
						source: String(edge.source ?? ''),
						target: String(edge.target ?? ''),
					};
				})
			: [],
		selectedNodeIds: Array.isArray(raw.selectedNodeIds)
			? raw.selectedNodeIds.map(String)
			: [],
	};
}

export function ExportApp() {
	const [loading, setLoading] = useState(true);
	const [workspaceId, setWorkspaceId] = useState('');
	const [flatData, setFlatData] = useState<ExportFlatNode[]>([]);
	const [checkedIds, setCheckedIds] = useState<string[]>([]);
	const [cascadeCheck, setCascadeCheck] = useState(true);
	const [settings, setSettings] = useState<ExportMergeSettings>(
		DEFAULT_EXPORT_SETTINGS,
	);
	const [previewContent, setPreviewContent] = useState<string | null>(null);
	const [previewSections, setPreviewSections] = useState<
		ExportPreviewSection[] | null
	>(null);
	const [splitFiles, setSplitFiles] = useState<ExportZipFileEntry[]>([]);
	const [previewLoading, setPreviewLoading] = useState(false);
	const [previewSource, setPreviewSource] = useState<PreviewSource | null>(
		null,
	);

	const applyPreview = useCallback(
		(
			orderedIds: string[],
			results: CrawlResultPreview[],
			mergeSettings: ExportMergeSettings,
		) => {
			const preview = buildExportPreview(
				orderedIds,
				flatData,
				results,
				mergeSettings,
			);
			setPreviewContent(preview.content);
			setPreviewSections(
				buildExportPreviewSections(
					orderedIds,
					flatData,
					results,
					mergeSettings,
				),
			);
			setSplitFiles(
				buildSplitExportFiles(orderedIds, flatData, results, mergeSettings),
			);
			return preview;
		},
		[flatData],
	);

	const loadSession = useCallback((session: ExportSessionSnapshot) => {
		const next = applySession(session);
		setWorkspaceId(next.workspaceId);
		setFlatData(next.flatData);
		setCheckedIds(next.checkedIds);
		setPreviewContent(null);
		setPreviewSections(null);
		setSplitFiles([]);
		setPreviewSource(null);
	}, []);

	useEffect(() => {
		if (!previewSource || previewLoading) return;
		applyPreview(previewSource.orderedIds, previewSource.results, settings);
	}, [settings, previewSource, previewLoading, applyPreview]);

	useEffect(() => {
		let cancelled = false;

		void scraperPort.getExportSession().then((initial) => {
			if (!cancelled && initial) loadSession(initial);
			if (!cancelled) setLoading(false);
		});

		const offOpen = Events.On(TOPIC_EXPORT_OPEN, (ev) => {
			const next = snapshotFromEventData(ev.data);
			if (next) loadSession(next);
		});

		return () => {
			cancelled = true;
			offOpen();
		};
	}, [loadSession]);

	const runPreview = async () => {
		if (!workspaceId || checkedIds.length === 0) return;
		setPreviewLoading(true);
		try {
			const orderedIds = preorderNodeIds(flatData, checkedIds);
			const results = await scraperPort.getNodeResults(workspaceId, orderedIds);
			setPreviewSource({ orderedIds, results });
			const preview = applyPreview(orderedIds, results, settings);
			if (preview.skippedCount > 0) {
				notifySuccess(messages.export.skippedNoResult(preview.skippedCount));
			}
		} catch (err) {
			notifyError(messages.export.previewStart, {
				description: err instanceof Error ? err.message : String(err),
			});
		} finally {
			setPreviewLoading(false);
		}
	};

	const copyPreview = async () => {
		if (!previewContent) return;
		try {
			await navigator.clipboard.writeText(previewContent);
			notifySuccess(messages.export.copied);
		} catch (err) {
			notifyError(messages.export.copyFailed, {
				description: err instanceof Error ? err.message : String(err),
			});
		}
	};

	const savePreview = async () => {
		if (!previewContent) return;
		const ext = settings.format === 'html' ? 'html' : 'md';
		try {
			if (settings.splitSave) {
				if (splitFiles.length === 0) return;
				await scraperPort.saveExportZip(splitFiles, ext);
				notifySuccess(messages.export.saveZipSuccess);
				return;
			}
			await scraperPort.saveExportFile(previewContent, ext);
			notifySuccess(messages.export.saveSuccess);
		} catch (err) {
			const errMessage = err instanceof Error ? err.message : String(err);
			if (errMessage.includes('cancelled by user')) return;
			notifyError(messages.export.saveFailed, {
				description: errMessage,
			});
		}
	};

	if (loading) {
		return (
			<div className='flex h-screen items-center justify-center bg-card text-sm text-muted-foreground'>
				{messages.bootstrapLoading}
			</div>
		);
	}

	const hasPreview = settings.splitSave
		? splitFiles.length > 0
		: previewContent !== null && previewContent.length > 0;

	return (
		<TooltipProvider>
			<div className='flex h-screen flex-col overflow-hidden bg-background text-foreground'>
				<Group orientation='horizontal' className='min-h-0 flex-1'>
					<Panel defaultSize='22%' minSize='14%' className='min-w-0'>
						<ExportOrderSidebar
							flatData={flatData}
							onFlatDataChange={setFlatData}
							checkedIds={checkedIds}
							onCheckedIdsChange={setCheckedIds}
							cascadeCheck={cascadeCheck}
							onCascadeCheckChange={setCascadeCheck}
						/>
					</Panel>
					<Separator className='w-1 shrink-0 bg-border hover:bg-primary/30' />
					<Panel minSize='30%' className='min-w-0'>
						<ExportPreviewPane
							content={previewContent}
							sections={previewSections}
							format={settings.format}
							loading={previewLoading}
						/>
					</Panel>
					<Separator className='w-1 shrink-0 bg-border hover:bg-primary/30' />
					<Panel defaultSize='20%' minSize='14rem' className='min-w-0'>
						<ExportSettingsSidebar
							settings={settings}
							onSettingsChange={setSettings}
							checkedCount={checkedIds.length}
							hasPreview={hasPreview}
							previewLoading={previewLoading}
							onPreviewStart={() => void runPreview()}
							onSave={() => void savePreview()}
							onCopy={() => void copyPreview()}
						/>
					</Panel>
				</Group>
				<Toaster duration={5000} />
			</div>
		</TooltipProvider>
	);
}
