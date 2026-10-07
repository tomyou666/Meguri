import { Events } from '@wailsio/runtime';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Group, Panel, Separator } from 'react-resizable-panels';
import { scraperPort } from '@/adapters';
import { ExportOrderSidebar } from '@/components/layout/export/ExportOrderSidebar';
import {
	ExportPreviewPane,
	trimExportBodies,
} from '@/components/layout/export/ExportPreviewPane';
import { ExportSettingsSidebar } from '@/components/layout/export/ExportSettingsSidebar';
import { Button } from '@/components/ui/button';
import { Toaster } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import { messages } from '@/i18n/messages';
import {
	acceptBodyResponse,
	type ExportPreviewRowBody,
	type ExportPreviewRowMeta,
} from '@/lib/exportPreviewVirtual';
import {
	buildInitialFlatTree,
	DEFAULT_EXPORT_SETTINGS,
	type ExportFlatNode,
	type ExportMergeSettings,
	initialCheckedIds,
	preorderNodeIds,
} from '@/lib/exportTree';
import { notifyError, notifySuccess } from '@/lib/notify';
import type { ExportSessionSnapshot } from '@/types/adapter';
import type { GraphEdge, GraphNode } from '@/types/graph';

const TOPIC_EXPORT_OPEN = 'export:open';
const TOPIC_EXPORT_SAVE_PROGRESS = 'export:save-progress';

function graphNodesFromSession(
	nodes: ExportSessionSnapshot['nodes'],
): GraphNode[] {
	return nodes.map((n) => ({
		id: n.id,
		urlNormalized: n.urlNormalized,
		label: n.label,
		position: { x: 0, y: 0 },
		nodeSettings: {},
		crawlExclude: n.crawlExclude,
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
						crawlExclude: node.crawlExclude === true,
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
	const [previewLoading, setPreviewLoading] = useState(false);
	const [rows, setRows] = useState<ExportPreviewRowMeta[]>([]);
	const [bodies, setBodies] = useState<Map<number, ExportPreviewRowBody>>(
		() => new Map(),
	);
	const [heightResetKey, setHeightResetKey] = useState(0);
	const [scrollToIndexRequest, setScrollToIndexRequest] = useState<
		number | null
	>(null);
	const [previewActive, setPreviewActive] = useState(false);
	const [saving, setSaving] = useState(false);
	const [saveProgress, setSaveProgress] = useState({ done: 0, total: 0 });
	const [sessionKey, setSessionKey] = useState(0);

	const generationRef = useRef(0);
	const orderedIdsRef = useRef<string[]>([]);
	const formatRef = useRef(settings.format);
	const fetchInFlightRef = useRef<string | null>(null);
	const skipAutoRefetchRef = useRef(true);
	const scrollIndexRef = useRef(0);

	const bumpGeneration = useCallback(() => {
		const next = generationRef.current + 1;
		generationRef.current = next;
		return next;
	}, []);

	const clearPreview = useCallback(() => {
		setRows([]);
		setBodies(new Map());
		setPreviewActive(false);
		fetchInFlightRef.current = null;
	}, []);

	const loadSession = useCallback(
		(session: ExportSessionSnapshot) => {
			const next = applySession(session);
			setWorkspaceId(next.workspaceId);
			setFlatData(next.flatData);
			setCheckedIds(next.checkedIds);
			bumpGeneration();
			clearPreview();
			void scraperPort.clearExportPreviewCache();
		},
		[bumpGeneration, clearPreview],
	);

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

		const offProgress = Events.On(TOPIC_EXPORT_SAVE_PROGRESS, (ev) => {
			const data = ev.data as { done?: number; total?: number } | undefined;
			if (!data) return;
			setSaveProgress({
				done: Number(data.done ?? 0),
				total: Number(data.total ?? 0),
			});
		});

		return () => {
			cancelled = true;
			offOpen();
			offProgress();
		};
	}, [loadSession]);

	const loadMeta = useCallback(
		async (opts: {
			orderedIds: string[];
			format: ExportMergeSettings['format'];
			generation: number;
			preserveIndex?: number | null;
		}) => {
			if (!workspaceId || opts.orderedIds.length === 0) return;
			setPreviewLoading(true);
			try {
				const res = await scraperPort.getExportPreviewMeta({
					workspaceId,
					nodeIds: opts.orderedIds,
					format: opts.format,
					generation: opts.generation,
				});
				if (res.generation !== generationRef.current) return;
				orderedIdsRef.current = opts.orderedIds;
				formatRef.current = opts.format;
				setRows(res.rows);
				setBodies(new Map());
				setPreviewActive(true);
				setSessionKey(opts.generation);
				fetchInFlightRef.current = null;
				if (opts.preserveIndex != null) {
					setScrollToIndexRequest(opts.preserveIndex);
				} else {
					setScrollToIndexRequest(0);
				}
				if (res.skippedCount > 0) {
					notifySuccess(messages.export.skippedNoResult(res.skippedCount));
				}
			} catch (err) {
				notifyError(messages.export.previewFailed, {
					description: err instanceof Error ? err.message : String(err),
				});
			} finally {
				setPreviewLoading(false);
			}
		},
		[workspaceId],
	);

	const runPreview = async () => {
		if (!workspaceId || checkedIds.length === 0) return;
		const orderedIds = preorderNodeIds(flatData, checkedIds);
		const gen = bumpGeneration();
		await loadMeta({
			orderedIds,
			format: settings.format,
			generation: gen,
			preserveIndex: null,
		});
	};

	const onFetchBodies = useCallback(
		(startIndex: number, endIndex: number) => {
			if (!workspaceId || !previewActive) return;
			const gen = generationRef.current;
			const key = `${gen}:${startIndex}:${endIndex}`;
			if (fetchInFlightRef.current === key) return;
			fetchInFlightRef.current = key;
			void (async () => {
				try {
					const res = await scraperPort.getExportPreviewBodies({
						workspaceId,
						format: formatRef.current,
						generation: gen,
						startIndex,
						endIndex,
					});
					if (
						!acceptBodyResponse({
							responseGeneration: res.generation,
							currentGeneration: generationRef.current,
							responseStart: res.startIndex,
							responseEnd: res.endIndex,
							requestedStart: startIndex,
							requestedEnd: endIndex,
							rows,
							responseRows: res.rows,
						})
					) {
						return;
					}
					setBodies((prev) => {
						const next = trimExportBodies(prev, startIndex, endIndex);
						const idToIndex = new Map(rows.map((r, i) => [r.id, i] as const));
						for (const row of res.rows) {
							const index = idToIndex.get(row.id);
							if (index == null || index < startIndex || index >= endIndex) {
								continue;
							}
							next.set(index, row);
						}
						return next;
					});
				} catch (err) {
					notifyError(messages.export.previewFailed, {
						description: err instanceof Error ? err.message : String(err),
					});
				} finally {
					if (fetchInFlightRef.current === key) {
						fetchInFlightRef.current = null;
					}
				}
			})();
		},
		[workspaceId, previewActive, rows],
	);

	useEffect(() => {
		if (skipAutoRefetchRef.current) {
			skipAutoRefetchRef.current = false;
			return;
		}
		if (!previewActive || !workspaceId) return;
		const orderedIds = preorderNodeIds(flatData, checkedIds);
		if (orderedIds.length === 0) {
			bumpGeneration();
			clearPreview();
			return;
		}
		// プレビュー開始で previewActive が true になっただけでは世代を進めない。
		// 本文取得が、まだ保存されていない世代を要求してセッション切れになる。
		const loaded = orderedIdsRef.current;
		if (
			orderedIds.length === loaded.length &&
			orderedIds.every((id, i) => id === loaded[i])
		) {
			return;
		}
		const gen = bumpGeneration();
		void loadMeta({
			orderedIds,
			format: formatRef.current,
			generation: gen,
			preserveIndex: null,
		});
	}, [
		previewActive,
		workspaceId,
		flatData,
		checkedIds,
		bumpGeneration,
		clearPreview,
		loadMeta,
	]);

	const prevFormatRef = useRef(settings.format);
	useEffect(() => {
		if (prevFormatRef.current === settings.format) return;
		prevFormatRef.current = settings.format;
		if (!previewActive || !workspaceId) return;
		const orderedIds = orderedIdsRef.current;
		if (orderedIds.length === 0) return;
		const preserve = scrollIndexRef.current;
		const gen = bumpGeneration();
		setHeightResetKey((k) => k + 1);
		void loadMeta({
			orderedIds,
			format: settings.format,
			generation: gen,
			preserveIndex: preserve,
		});
	}, [settings.format, previewActive, workspaceId, bumpGeneration, loadMeta]);

	useEffect(() => {
		if (scrollToIndexRequest != null) {
			scrollIndexRef.current = scrollToIndexRequest;
		}
	}, [scrollToIndexRequest]);
	const headingSepKey = `${settings.includeHeading}:${settings.headingField}:${settings.separator}`;
	const prevHeadingSepRef = useRef(headingSepKey);
	useEffect(() => {
		if (prevHeadingSepRef.current === headingSepKey) return;
		prevHeadingSepRef.current = headingSepKey;
		setHeightResetKey((k) => k + 1);
	}, [headingSepKey]);

	const saveExport = async () => {
		if (!workspaceId || checkedIds.length === 0 || saving) return;
		const orderedIds = preorderNodeIds(flatData, checkedIds);
		setSaving(true);
		setSaveProgress({ done: 0, total: orderedIds.length });
		try {
			await scraperPort.saveExport({
				workspaceId,
				nodeIds: orderedIds,
				format: settings.format,
				separator: settings.separator,
				includeHeading: settings.includeHeading,
				headingField: settings.headingField,
				splitSave: settings.splitSave,
			});
			notifySuccess(
				settings.splitSave
					? messages.export.saveZipSuccess
					: messages.export.saveSuccess,
			);
		} catch (err) {
			const errMessage = err instanceof Error ? err.message : String(err);
			if (errMessage.includes('cancelled by user')) return;
			if (errMessage.includes('no exportable content')) {
				notifyError(messages.export.saveNoContent);
				return;
			}
			notifyError(messages.export.saveFailed, {
				description: errMessage,
			});
		} finally {
			setSaving(false);
			setSaveProgress({ done: 0, total: 0 });
		}
	};

	const cancelSave = async () => {
		try {
			await scraperPort.cancelExportSave();
		} catch {
			/* ignore */
		}
		setSaving(false);
	};

	useEffect(() => {
		const onUnload = () => {
			void scraperPort.cancelExportSave();
		};
		window.addEventListener('beforeunload', onUnload);
		return () => window.removeEventListener('beforeunload', onUnload);
	}, []);

	if (loading) {
		return (
			<div className='flex h-screen items-center justify-center bg-card text-muted-foreground text-sm'>
				{messages.bootstrapLoading}
			</div>
		);
	}

	return (
		<TooltipProvider>
			<div className='relative flex h-screen flex-col overflow-hidden bg-background text-foreground'>
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
							rows={rows}
							bodies={bodies}
							format={settings.format}
							settings={settings}
							flatData={flatData}
							loading={previewLoading}
							disabled={saving}
							sessionKey={sessionKey}
							onFetchBodies={onFetchBodies}
							onMarkdownViewChange={() => setHeightResetKey((k) => k + 1)}
							heightResetKey={heightResetKey}
							scrollToIndexRequest={scrollToIndexRequest}
						/>
					</Panel>
					<Separator className='w-1 shrink-0 bg-border hover:bg-primary/30' />
					<Panel defaultSize='20%' minSize='14rem' className='min-w-0'>
						<ExportSettingsSidebar
							settings={settings}
							onSettingsChange={setSettings}
							checkedCount={checkedIds.length}
							previewLoading={previewLoading}
							saving={saving}
							onPreviewStart={() => void runPreview()}
							onSave={() => void saveExport()}
						/>
					</Panel>
				</Group>
				{saving && (
					<div className='absolute inset-0 z-50 flex items-center justify-center bg-background/80'>
						<div className='flex w-80 flex-col gap-3 rounded-lg border border-border bg-card p-4 shadow-lg'>
							<p className='font-medium text-sm'>
								{messages.export.savingTitle}
							</p>
							<p className='text-muted-foreground text-xs'>
								{messages.export.savingProgress(
									saveProgress.done,
									saveProgress.total,
								)}
							</p>
							<Button
								size='sm'
								variant='outline'
								onClick={() => void cancelSave()}
							>
								{messages.export.saveCancel}
							</Button>
						</div>
					</div>
				)}
				<Toaster duration={5000} />
			</div>
		</TooltipProvider>
	);
}
