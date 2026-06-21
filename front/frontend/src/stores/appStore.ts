import { create } from 'zustand';
import { scraperPort } from '@/adapters';
import { messages } from '@/i18n/messages';
import { validatePartialConfig } from '@/lib/configValidation';
import {
	computeDagrePositions,
	type DagreLayoutDirection,
	fallbackNearParent,
	positionForDiscoveredNode,
} from '@/lib/dagreLayout';
import {
	debouncedPatchNodePositions,
	debouncedSaveWorkspace,
	type NodePositionPatch,
} from '@/lib/debouncedWorkspaceSave';
import { DEFAULT_APP_CONFIG } from '@/lib/defaults';
import {
	collectDescendantUrls,
	getBfsNodeOrder,
	getDescendantNodeIds,
} from '@/lib/graph';
import { normalizeUrl } from '@/lib/normalizeUrl';
import { notifyError, notifySuccess } from '@/lib/notify';
import { withDerivedContentFormats } from '@/lib/previewFormats';
import {
	redoGraph,
	syncGraphHistory,
	undoGraph,
} from '@/stores/graphHistoryStore';
import type { WorkspaceDiff } from '@/types/adapter';
import type { PartialConfig } from '@/types/config';
import type {
	CrawlError,
	CrawlLogEntry,
	CrawlResultPreview,
	CrawlRunStatus,
	CrawlRunSummary,
	RunMode,
} from '@/types/crawl';
import type { GraphNode, NodeStatus } from '@/types/graph';
import type { Workspace } from '@/types/workspace';
import * as ScraperService from '../../bindings/scraperbot-front/internal/usecase/wails_service/scraperservice';

function uid(): string {
	return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

type PersistOptions =
	| { kind: 'workspace' }
	| { kind: 'positions'; workspaceId: string; update: NodePositionPatch };

function syncHistory(
	workspaces: Workspace[],
	activeWorkspaceId: string | null,
	persist: PersistOptions = { kind: 'workspace' },
) {
	syncGraphHistory(workspaces, activeWorkspaceId);
	if (persist.kind === 'positions') {
		debouncedPatchNodePositions(persist.workspaceId, persist.update);
		return;
	}
	const active = workspaces.find((w) => w.id === activeWorkspaceId);
	if (active) {
		debouncedSaveWorkspace(active);
	}
}

type PatchWorkspacesOptions = {
	recordHistory?: boolean;
	persist?: PersistOptions;
};

function patchWorkspaces(
	set: (fn: (s: AppState) => Partial<AppState>) => void,
	_get: () => AppState,
	updater: (workspaces: Workspace[]) => Workspace[],
	options: PatchWorkspacesOptions = {},
) {
	const recordHistory = options.recordHistory ?? true;
	const persist = options.persist ?? { kind: 'workspace' };
	set((s) => {
		const workspaces = updater(s.workspaces);
		if (recordHistory) {
			syncHistory(workspaces, s.activeWorkspaceId, persist);
		}
		return { workspaces };
	});
}

function emptyWorkspace(name: string, seedUrl: string): Workspace {
	const normalized = normalizeUrl(seedUrl);
	const rootId = uid();
	return {
		id: uid(),
		name,
		seedUrl: normalized,
		settings: { crawl: { enabled: true } },
		exclude_urls: [],
		nodes: [
			{
				id: rootId,
				urlNormalized: normalized,
				label: normalized,
				position: { x: 250, y: 200 },
				nodeSettings: {},
				crawlExclude: false,
				status: 'idle',
			},
		],
		edges: [],
		graphLayoutDirection: 'LR',
		collapsedNodeIds: [],
		expandedDetailNodeIds: [],
		createdAt: new Date().toISOString(),
	};
}

interface AppState {
	bootstrapped: boolean;
	appDefaults: PartialConfig;
	workspaces: Workspace[];
	activeWorkspaceId: string | null;
	selectedNodeId: string | null;
	selectedNodeIds: string[];
	selectionAnchorId: string | null;
	/** クリック選択時に useOnSelectionChange による上書きを防ぐ */
	_suppressSelectionSync: boolean;
	graphToolMode: 'pan' | 'select';
	leftSidebarCollapsed: boolean;
	rightSidebarCollapsed: boolean;
	minimapCollapsed: boolean;
	clipboard: {
		nodes: GraphNode[];
		edges: { source: string; target: string }[];
	} | null;
	loadedNodeResult: CrawlResultPreview | null;
	resultPreview: CrawlResultPreview[] | null;
	workspaceDiffCache: Record<string, WorkspaceDiff>;
	mergeSheetOpen: boolean;
	mergeSheetContent: string | null;
	runMode: RunMode;
	rescrapeExisting: boolean;
	crawlStatus: CrawlRunStatus;
	crawlLogs: CrawlLogEntry[];
	runHistory: CrawlRunSummary[];
	crawlError: CrawlError;
	showNewWorkspaceDialog: boolean;
	showAddNodeDialog: boolean;
	showDeleteNodeDialog: boolean;
	pendingDeleteWorkspaceId: string | null;
	pendingDuplicateWorkspaceId: string | null;
	addNodeContextPosition: { x: number; y: number } | null;

	_abortController: AbortController | null;
	_activeRunId: string | null;
	_paused: boolean;

	bootstrap: () => Promise<void>;
	setAppDefaults: (config: PartialConfig) => void;
	persistAppDefaults: (config: PartialConfig) => Promise<boolean>;
	persistWorkspaceSettings: (settings: PartialConfig) => Promise<boolean>;
	persistNodeSettings: (
		nodeId: string,
		settings: PartialConfig,
	) => Promise<boolean>;
	openNewWorkspaceDialog: () => void;
	closeNewWorkspaceDialog: () => void;
	createWorkspace: (name: string, seedUrl: string) => void;
	setActiveWorkspace: (id: string) => void;
	openDeleteWorkspaceDialog: (id: string) => void;
	closeDeleteWorkspaceDialog: () => void;
	openDuplicateWorkspaceDialog: (id: string) => void;
	closeDuplicateWorkspaceDialog: () => void;
	confirmDeleteWorkspace: () => Promise<void>;
	confirmDuplicateWorkspace: (name: string) => Promise<void>;
	loadWorkspaceFromServer: (id: string) => Promise<void>;
	selectNode: (
		id: string | null,
		opts?: { additive?: boolean; range?: boolean },
	) => void;
	selectNodes: (ids: string[]) => void;
	setGraphToolMode: (mode: 'pan' | 'select') => void;
	selectAllNodes: () => void;
	clearNodeSelection: () => void;
	toggleLeftSidebar: () => void;
	toggleRightSidebar: () => void;
	toggleMinimap: () => void;
	undo: () => void;
	redo: () => void;
	copySelectedNodes: () => void;
	pasteNodes: () => void;
	addEdge: (source: string, target: string) => boolean;
	collapseNodes: (nodeIds: string[]) => void;
	expandNodes: (nodeIds: string[]) => void;
	toggleNodeDetailExpand: (nodeId: string) => void;
	toggleNodeSubtreeCollapse: (nodeId: string) => void;
	expandAllNodes: () => void;
	collapseAllNodes: () => void;
	deleteSelectedNodes: () => void;
	fetchSelectedNodeResult: () => Promise<void>;
	previewSelectedResults: () => Promise<void>;
	mergeAllResults: () => Promise<void>;
	mergeSelectedResults: () => Promise<void>;
	saveSelectedResults: () => Promise<void>;
	deleteSelectedResults: () => Promise<void>;
	bulkScrapeSelected: () => Promise<void>;
	fetchWorkspaceDiff: (workspaceId: string) => Promise<WorkspaceDiff>;
	closeMergeSheet: () => void;
	setRunMode: (mode: RunMode) => void;
	setRescrapeExisting: (value: boolean) => void;
	updateNodePosition: (id: string, position: { x: number; y: number }) => void;
	layoutWorkspaceGraph: () => void;
	setGraphLayoutDirection: (direction: DagreLayoutDirection) => void;
	removeEdges: (edgeIds: string[]) => void;
	openAddNodeDialog: (screenPos?: { x: number; y: number }) => void;
	closeAddNodeDialog: () => void;
	addNode: (url: string) => void;
	openDeleteNodeDialog: () => void;
	closeDeleteNodeDialog: () => void;
	deleteSelectedSubtree: () => void;
	setNodeCrawlExclude: (nodeId: string, excluded: boolean) => void;
	updateWorkspaceSettings: (settings: PartialConfig) => void;
	updateNodeSettings: (nodeId: string, settings: PartialConfig) => void;
	clearCrawlError: () => void;
	startCrawl: (override?: {
		mode?: RunMode;
		nodeIds?: string[];
	}) => Promise<void>;
	pauseCrawl: () => void;
	resumeCrawl: () => void;
	stopCrawl: () => void;

	getActiveWorkspace: () => Workspace | null;
	getSelectedNode: () => GraphNode | null;
}

export const useAppStore = create<AppState>((set, get) => ({
	bootstrapped: false,
	appDefaults: DEFAULT_APP_CONFIG,
	workspaces: [],
	activeWorkspaceId: null,
	selectedNodeId: null,
	selectedNodeIds: [],
	selectionAnchorId: null,
	_suppressSelectionSync: false,
	graphToolMode: 'pan',
	leftSidebarCollapsed: false,
	rightSidebarCollapsed: false,
	minimapCollapsed: false,
	clipboard: null,
	loadedNodeResult: null,
	resultPreview: null,
	workspaceDiffCache: {},
	mergeSheetOpen: false,
	mergeSheetContent: null,
	runMode: 1,
	rescrapeExisting: false,
	crawlStatus: 'idle',
	crawlLogs: [],
	runHistory: [],
	crawlError: null,
	showNewWorkspaceDialog: true,
	showAddNodeDialog: false,
	showDeleteNodeDialog: false,
	pendingDeleteWorkspaceId: null,
	pendingDuplicateWorkspaceId: null,
	addNodeContextPosition: null,
	_abortController: null,
	_activeRunId: null as string | null,
	_paused: false,

	bootstrap: async () => {
		const defaults = await scraperPort.getAppDefaults();
		const list = await scraperPort.listWorkspaces();
		const workspaces: Workspace[] = [];
		for (const item of list) {
			const ws = await scraperPort.loadWorkspace(item.id);
			if (ws) workspaces.push(ws);
		}
		let activeWorkspaceId: string | null = null;
		if (list.length > 0) {
			const latest = list.reduce((a, b) =>
				a.updatedAt >= b.updatedAt ? a : b,
			);
			activeWorkspaceId = latest.id;
		}
		const active = workspaces.find((w) => w.id === activeWorkspaceId);
		set({
			bootstrapped: true,
			appDefaults: defaults,
			workspaces,
			activeWorkspaceId,
			showNewWorkspaceDialog: list.length === 0,
			selectedNodeId: active?.nodes[0]?.id ?? null,
			selectedNodeIds: active?.nodes[0]?.id ? [active.nodes[0].id] : [],
		});
		if (workspaces.length > 0) {
			syncGraphHistory(workspaces, activeWorkspaceId);
		}
	},

	setAppDefaults: (config) => {
		set({ appDefaults: config });
		void scraperPort.setAppDefaults(config);
	},

	persistAppDefaults: async (config) => {
		const withFormats = withDerivedContentFormats(config);
		const validated = validatePartialConfig(withFormats);
		if (!validated.ok) {
			notifyError(messages.settings.saveFailed, {
				description: messages.settings.validationFailed,
			});
			return false;
		}
		try {
			await scraperPort.saveAppDefaults(validated.data);
			get().setAppDefaults(validated.data);
			notifySuccess(messages.settings.saveSuccess);
			return true;
		} catch (err) {
			notifyError(messages.settings.saveFailed, {
				description: err instanceof Error ? err.message : undefined,
			});
			return false;
		}
	},

	persistWorkspaceSettings: async (settings) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return false;
		const withFormats = withDerivedContentFormats(settings, get().appDefaults);
		const validated = validatePartialConfig(withFormats);
		if (!validated.ok) {
			notifyError(messages.settings.saveFailed, {
				description: messages.settings.validationFailed,
			});
			return false;
		}
		try {
			await scraperPort.saveWorkspaceSettings(ws.id, validated.data);
			get().updateWorkspaceSettings(validated.data);
			notifySuccess(messages.settings.saveSuccess);
			return true;
		} catch (err) {
			notifyError(messages.settings.saveFailed, {
				description: err instanceof Error ? err.message : undefined,
			});
			return false;
		}
	},

	persistNodeSettings: async (nodeId, settings) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return false;
		const validated = validatePartialConfig(settings);
		if (!validated.ok) {
			notifyError(messages.settings.saveFailed, {
				description: messages.settings.validationFailed,
			});
			return false;
		}
		try {
			await scraperPort.saveNodeSettings(ws.id, nodeId, validated.data);
			get().updateNodeSettings(nodeId, validated.data);
			notifySuccess(messages.settings.saveSuccess);
			return true;
		} catch (err) {
			notifyError(messages.settings.saveFailed, {
				description: err instanceof Error ? err.message : undefined,
			});
			return false;
		}
	},

	openNewWorkspaceDialog: () => set({ showNewWorkspaceDialog: true }),
	closeNewWorkspaceDialog: () => set({ showNewWorkspaceDialog: false }),

	createWorkspace: (name, seedUrl) => {
		try {
			const ws = emptyWorkspace(name, seedUrl);
			set((s) => {
				const workspaces = [...s.workspaces, ws];
				syncHistory(workspaces, ws.id);
				void scraperPort.saveWorkspace(ws);
				return {
					workspaces,
					activeWorkspaceId: ws.id,
					selectedNodeId: ws.nodes[0]?.id ?? null,
					selectedNodeIds: ws.nodes[0]?.id ? [ws.nodes[0].id] : [],
					showNewWorkspaceDialog: false,
				};
			});
		} catch (e) {
			const message =
				e instanceof Error ? e.message : 'ワークスペース作成に失敗しました';
			notifyError(messages.error.globalBanner, { description: message });
		}
	},

	setActiveWorkspace: (id) =>
		set({
			activeWorkspaceId: id,
			selectedNodeId: null,
		}),

	openDeleteWorkspaceDialog: (id) => set({ pendingDeleteWorkspaceId: id }),
	closeDeleteWorkspaceDialog: () => set({ pendingDeleteWorkspaceId: null }),
	openDuplicateWorkspaceDialog: (id) =>
		set({ pendingDuplicateWorkspaceId: id }),
	closeDuplicateWorkspaceDialog: () =>
		set({ pendingDuplicateWorkspaceId: null }),

	confirmDeleteWorkspace: async () => {
		const id = get().pendingDeleteWorkspaceId;
		if (!id) return;
		try {
			await scraperPort.deleteWorkspace(id);
			set((s) => {
				const workspaces = s.workspaces.filter((w) => w.id !== id);
				const activeWorkspaceId =
					s.activeWorkspaceId === id
						? (workspaces[0]?.id ?? null)
						: s.activeWorkspaceId;
				const { [id]: _removed, ...workspaceDiffCache } = s.workspaceDiffCache;
				return {
					workspaces,
					activeWorkspaceId,
					workspaceDiffCache,
					selectedNodeId: null,
					selectedNodeIds: [],
					pendingDeleteWorkspaceId: null,
				};
			});
		} catch (err) {
			const message = err instanceof Error ? err.message : String(err);
			notifyError(messages.error.deleteWorkspaceFailed, {
				description: message,
			});
		}
	},

	confirmDuplicateWorkspace: async (name) => {
		const id = get().pendingDuplicateWorkspaceId;
		if (!id) return;
		const trimmed = name.trim();
		if (!trimmed) return;
		try {
			const copy = await scraperPort.duplicateWorkspace(id, trimmed);
			set((s) => {
				const workspaces = [...s.workspaces, copy];
				syncHistory(workspaces, copy.id);
				return {
					workspaces,
					activeWorkspaceId: copy.id,
					selectedNodeId: copy.nodes[0]?.id ?? null,
					selectedNodeIds: copy.nodes[0]?.id ? [copy.nodes[0].id] : [],
					pendingDuplicateWorkspaceId: null,
				};
			});
		} catch (err) {
			const message = err instanceof Error ? err.message : String(err);
			notifyError(messages.error.duplicateWorkspaceFailed, {
				description: message,
			});
		}
	},

	loadWorkspaceFromServer: async (id) => {
		const ws = await scraperPort.loadWorkspace(id);
		if (!ws) return;
		set((s) => {
			const exists = s.workspaces.some((w) => w.id === id);
			const workspaces = exists
				? s.workspaces.map((w) => (w.id === id ? ws : w))
				: [...s.workspaces, ws];
			syncGraphHistory(workspaces, id);
			return {
				workspaces,
				activeWorkspaceId: id,
				selectedNodeId: ws.nodes[0]?.id ?? null,
				selectedNodeIds: ws.nodes[0]?.id ? [ws.nodes[0].id] : [],
				showNewWorkspaceDialog: false,
			};
		});
	},

	selectNode: (id, opts) => {
		const ws = get().getActiveWorkspace();
		if (!id) {
			set({
				selectedNodeId: null,
				selectedNodeIds: [],
				selectionAnchorId: null,
				loadedNodeResult: null,
				resultPreview: null,
			});
			return;
		}
		if (!ws) return;

		const seedId = ws.nodes.find((n) => n.urlNormalized === ws.seedUrl)?.id;
		const order = getBfsNodeOrder(seedId, ws.nodes, ws.edges);
		let selectedNodeIds = [id];
		let selectionAnchorId = get().selectionAnchorId;

		if (opts?.range) {
			const anchor =
				selectionAnchorId ??
				get().selectedNodeId ??
				get().selectedNodeIds[0] ??
				null;
			if (anchor) {
				const a = order.indexOf(anchor);
				const b = order.indexOf(id);
				if (a >= 0 && b >= 0) {
					const [lo, hi] = a < b ? [a, b] : [b, a];
					selectedNodeIds = order.slice(lo, hi + 1);
				}
			}
			selectionAnchorId = anchor ?? id;
		} else if (opts?.additive) {
			const cur = get().selectedNodeIds;
			selectedNodeIds = cur.includes(id)
				? cur.filter((x) => x !== id)
				: [...cur, id];
			if (!selectionAnchorId) selectionAnchorId = id;
		} else {
			selectionAnchorId = id;
			selectedNodeIds = [id];
		}

		const primary = selectedNodeIds[selectedNodeIds.length - 1] ?? id;
		set({
			selectedNodeId: primary,
			selectedNodeIds,
			selectionAnchorId,
			loadedNodeResult: null,
			resultPreview: null,
		});
		if (selectedNodeIds.length === 1) {
			const node = ws.nodes.find((n) => n.id === primary);
			if (node?.status === 'success') {
				void get().fetchSelectedNodeResult();
			}
		}
	},

	selectNodes: (ids) => {
		const primary = ids[ids.length - 1] ?? null;
		set({
			selectedNodeIds: ids,
			selectedNodeId: primary,
			selectionAnchorId: primary,
			loadedNodeResult: null,
			resultPreview: null,
		});
		if (ids.length === 1) {
			const ws = get().getActiveWorkspace();
			const node = ws?.nodes.find((n) => n.id === ids[0]);
			if (node?.status === 'success') {
				void get().fetchSelectedNodeResult();
			}
		}
	},

	setGraphToolMode: (mode) => set({ graphToolMode: mode }),

	selectAllNodes: () => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		const ids = ws.nodes.map((n) => n.id);
		const primary = ids[ids.length - 1] ?? null;
		set({
			selectedNodeIds: ids,
			selectedNodeId: primary,
			selectionAnchorId: primary,
		});
	},

	clearNodeSelection: () =>
		set({
			selectedNodeId: null,
			selectedNodeIds: [],
			selectionAnchorId: null,
		}),

	toggleLeftSidebar: () =>
		set((s) => ({ leftSidebarCollapsed: !s.leftSidebarCollapsed })),
	toggleRightSidebar: () =>
		set((s) => ({ rightSidebarCollapsed: !s.rightSidebarCollapsed })),
	toggleMinimap: () => set((s) => ({ minimapCollapsed: !s.minimapCollapsed })),

	undo: () => {
		const { workspaces, activeWorkspaceId } = undoGraph();
		set({ workspaces, activeWorkspaceId });
	},

	redo: () => {
		const { workspaces, activeWorkspaceId } = redoGraph();
		set({ workspaces, activeWorkspaceId });
	},
	setRunMode: (mode) => set({ runMode: mode }),
	setRescrapeExisting: (value) => set({ rescrapeExisting: value }),

	updateNodePosition: (id, position) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		patchWorkspaces(
			set,
			get,
			(workspaces) =>
				workspaces.map((w) =>
					w.id !== ws.id
						? w
						: {
								...w,
								nodes: w.nodes.map((n) =>
									n.id === id ? { ...n, position, userPositioned: true } : n,
								),
							},
				),
			{
				persist: {
					kind: 'positions',
					workspaceId: ws.id,
					update: { nodeId: id, position, userPositioned: true },
				},
			},
		);
	},

	layoutWorkspaceGraph: () => {
		const ws = get().getActiveWorkspace();
		if (!ws || ws.nodes.length === 0) return;
		const direction = ws.graphLayoutDirection ?? 'LR';
		const positions = computeDagrePositions(ws.nodes, ws.edges, direction);
		set((s) => ({
			workspaces: s.workspaces.map((w) =>
				w.id !== ws.id
					? w
					: {
							...w,
							nodes: w.nodes.map((n) => {
								const pos = positions.get(n.id);
								return pos ? { ...n, position: pos, userPositioned: false } : n;
							}),
						},
			),
		}));
	},

	setGraphLayoutDirection: (direction) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		set((s) => ({
			workspaces: s.workspaces.map((w) =>
				w.id === ws.id ? { ...w, graphLayoutDirection: direction } : w,
			),
		}));
		get().layoutWorkspaceGraph();
	},

	removeEdges: (edgeIds) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		set((s) => ({
			workspaces: s.workspaces.map((w) =>
				w.id !== ws.id
					? w
					: {
							...w,
							edges: w.edges.filter((e) => !edgeIds.includes(e.id)),
						},
			),
		}));
	},

	openAddNodeDialog: (screenPos) =>
		set({ showAddNodeDialog: true, addNodeContextPosition: screenPos ?? null }),
	closeAddNodeDialog: () =>
		set({ showAddNodeDialog: false, addNodeContextPosition: null }),

	addNode: (url) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		try {
			const normalized = normalizeUrl(url);
			const existing = ws.nodes.find((n) => n.urlNormalized === normalized);
			if (existing) {
				set({ selectedNodeId: existing.id, showAddNodeDialog: false });
				return;
			}
			const id = uid();
			const pos = get().addNodeContextPosition ?? { x: 400, y: 300 };
			const node: GraphNode = {
				id,
				urlNormalized: normalized,
				label: normalized,
				position: pos,
				userPositioned: true,
				origin: 'manual',
				nodeSettings: {},
				crawlExclude: false,
				status: 'idle',
			};
			patchWorkspaces(set, get, (workspaces) =>
				workspaces.map((w) =>
					w.id === ws.id ? { ...w, nodes: [...w.nodes, node] } : w,
				),
			);
			set({
				selectedNodeId: id,
				selectedNodeIds: [id],
				showAddNodeDialog: false,
				addNodeContextPosition: null,
			});
		} catch (e) {
			const message = e instanceof Error ? e.message : 'URL が不正です';
			notifyError(messages.error.globalBanner, { description: message });
		}
	},

	openDeleteNodeDialog: () => set({ showDeleteNodeDialog: true }),
	closeDeleteNodeDialog: () => set({ showDeleteNodeDialog: false }),

	deleteSelectedSubtree: () => {
		const ws = get().getActiveWorkspace();
		const nodeId = get().selectedNodeId;
		if (!ws || !nodeId) return;
		const desc = getDescendantNodeIds(nodeId, ws.edges);
		const removeIds = new Set([nodeId, ...desc]);
		const removeUrls = new Set(
			ws.nodes.filter((n) => removeIds.has(n.id)).map((n) => n.urlNormalized),
		);

		set((s) => ({
			workspaces: s.workspaces.map((w) => {
				if (w.id !== ws.id) return w;
				const nodes = w.nodes.filter((n) => !removeIds.has(n.id));
				const edges = w.edges.filter(
					(e) => !removeIds.has(e.source) && !removeIds.has(e.target),
				);
				const exclude_urls = w.exclude_urls.filter((u) => !removeUrls.has(u));
				return { ...w, nodes, edges, exclude_urls };
			}),
			selectedNodeId: null,
			showDeleteNodeDialog: false,
		}));
	},

	setNodeCrawlExclude: (nodeId, excluded) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		const urls = collectDescendantUrls(nodeId, ws.nodes, ws.edges);
		set((s) => ({
			workspaces: s.workspaces.map((w) => {
				if (w.id !== ws.id) return w;
				let exclude_urls = [...w.exclude_urls];
				if (excluded) {
					for (const u of urls) {
						if (!exclude_urls.includes(u)) exclude_urls.push(u);
					}
				} else {
					exclude_urls = exclude_urls.filter((u) => !urls.includes(u));
				}
				return {
					...w,
					exclude_urls,
					nodes: w.nodes.map((n) =>
						n.id === nodeId ? { ...n, crawlExclude: excluded } : n,
					),
				};
			}),
		}));
	},

	updateWorkspaceSettings: (settings) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		set((s) => ({
			workspaces: s.workspaces.map((w) =>
				w.id === ws.id ? { ...w, settings: { ...w.settings, ...settings } } : w,
			),
		}));
	},

	updateNodeSettings: (nodeId, settings) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		set((s) => ({
			workspaces: s.workspaces.map((w) =>
				w.id === ws.id
					? {
							...w,
							nodes: w.nodes.map((n) =>
								n.id === nodeId
									? { ...n, nodeSettings: { ...n.nodeSettings, ...settings } }
									: n,
							),
						}
					: w,
			),
		}));
	},

	clearCrawlError: () => set({ crawlError: null }),

	pauseCrawl: () => {
		const runId = get()._activeRunId;
		if (runId) void ScraperService.PauseCrawl(runId);
		set({ _paused: true, crawlStatus: 'paused' });
	},
	resumeCrawl: () => {
		const runId = get()._activeRunId;
		if (runId) void ScraperService.ResumeCrawl(runId);
		set({ _paused: false, crawlStatus: 'running' });
	},

	stopCrawl: () => {
		const runId = get()._activeRunId;
		if (runId) void ScraperService.StopCrawl(runId);
		get()._abortController?.abort();
		set({ crawlStatus: 'idle', _paused: false, _activeRunId: null });
	},

	startCrawl: async (override) => {
		const state = get();
		const ws = state.getActiveWorkspace();
		if (!ws) return;

		const mode = override?.mode ?? state.runMode;

		if (mode === 4) {
			const nodeIds = override?.nodeIds ?? state.selectedNodeIds;
			if (nodeIds.length === 0) {
				set({
					crawlError: {
						type: 'crawl',
						message: 'モード 4 ではノードを 1 件以上選択してください',
						at: new Date().toISOString(),
					},
				});
				return;
			}
		} else if (mode !== 1 && !state.selectedNodeId) {
			set({
				crawlError: {
					type: 'crawl',
					message: 'モード 2/3 ではノードを選択してください',
					at: new Date().toISOString(),
				},
			});
			return;
		}

		state._abortController?.abort();
		const ac = new AbortController();
		set({
			_abortController: ac,
			_paused: false,
			crawlStatus: 'running',
			crawlError: null,
			crawlLogs: [],
		});

		let runId = '';
		const startedAt = new Date().toISOString();

		const patchNode = (nodeId: string, patch: Partial<GraphNode>) => {
			set((s) => ({
				workspaces: s.workspaces.map((w) =>
					w.id === ws.id
						? {
								...w,
								nodes: w.nodes.map((n) =>
									n.id === nodeId ? { ...n, ...patch } : n,
								),
							}
						: w,
				),
			}));
		};

		const getWs = () => get().getActiveWorkspace()!;

		const nodeIds =
			mode === 4
				? (override?.nodeIds ?? state.selectedNodeIds)
				: undefined;

		runId = await scraperPort.startCrawl({
			workspaceId: ws.id,
			mode,
			startNodeId: state.selectedNodeId ?? undefined,
			nodeIds,
			rescrapeExisting: state.rescrapeExisting,
			appDefaults: state.appDefaults,
			signal: ac.signal,
			isPaused: () => get()._paused,
			waitWhilePaused: async () => {
				while (get()._paused && !ac.signal.aborted) {
					await new Promise((r) => setTimeout(r, 100));
				}
			},
			onRunStarted: (id) => set({ _activeRunId: id }),
			getWorkspace: () => get().getActiveWorkspace()!,
			onNodeStarted: (nodeId, url) => {
				patchNode(nodeId, { status: 'running' as NodeStatus, label: url });
			},
			onNodeSucceeded: (nodeId, result: CrawlResultPreview) => {
				patchNode(nodeId, {
					status: 'success',
					lastResult: result,
					lastError: undefined,
				});
			},
			onNodeFailed: (nodeId, _url, error) => {
				patchNode(nodeId, { status: 'error', lastError: error });
			},
			onNodeSkipped: (nodeId) => {
				patchNode(nodeId, { status: 'skipped', lastError: undefined });
			},
			onLinkSkipped: (parentUrl, targetUrl, reason) => {
				set((s) => ({
					crawlLogs: [
						...s.crawlLogs,
						{
							at: new Date().toISOString(),
							parentUrl,
							targetUrl,
							reason,
						},
					],
				}));
			},
			onEdgeDiscovered: (sourceId, targetId, targetUrl) => {
				const current = getWs();
				const targetExists = current.nodes.some((n) => n.id === targetId);
				const nodes = [...current.nodes];
				const edges = [...current.edges];
				const edgeId = `e-${sourceId}-${targetId}`;
				const edgeIsNew = !edges.some((e) => e.id === edgeId);
				if (edgeIsNew) {
					edges.push({ id: edgeId, source: sourceId, target: targetId });
				}
				if (!targetExists) {
					const parent = nodes.find((n) => n.id === sourceId);
					const direction = current.graphLayoutDirection ?? 'LR';
					const fallback = parent
						? fallbackNearParent(parent, direction)
						: { x: 400, y: 300 };
					const position = positionForDiscoveredNode(
						[
							...nodes,
							{
								id: targetId,
								urlNormalized: targetUrl,
								label: targetUrl,
								position: fallback,
								nodeSettings: {},
								crawlExclude: false,
								status: 'idle' as const,
							},
						],
						edges,
						targetId,
						fallback,
						direction,
					);
					nodes.push({
						id: targetId,
						urlNormalized: targetUrl,
						label: targetUrl,
						position,
						userPositioned: false,
						origin: 'crawl',
						nodeSettings: {},
						crawlExclude: false,
						status: 'idle',
					});
				}
				set((s) => ({
					workspaces: s.workspaces.map((w) =>
						w.id === ws.id ? { ...w, nodes, edges } : w,
					),
				}));
			},
			onCrawlCompleted: (summary) => {
				const full: CrawlRunSummary = {
					id: runId,
					startedAt,
					...summary,
				};
				set((s) => ({
					crawlStatus: 'idle',
					_abortController: null,
					_activeRunId: null,
					runHistory: [full, ...s.runHistory].slice(0, 20),
				}));
			},
			onCrawlError: (message) => {
				set({
					crawlStatus: 'idle',
					_abortController: null,
					_activeRunId: null,
					crawlError: {
						type: 'crawl',
						message,
						runId,
						at: new Date().toISOString(),
					},
				});
			},
		});
	},

	copySelectedNodes: () => {
		const ws = get().getActiveWorkspace();
		const ids = new Set(get().selectedNodeIds);
		if (!ws || ids.size === 0) return;
		const nodes = ws.nodes.filter((n) => ids.has(n.id));
		const edges = ws.edges
			.filter((e) => ids.has(e.source) && ids.has(e.target))
			.map((e) => ({ source: e.source, target: e.target }));
		set({ clipboard: { nodes, edges } });
	},

	pasteNodes: () => {
		const ws = get().getActiveWorkspace();
		const clip = get().clipboard;
		if (!ws || !clip?.nodes.length) return;
		const idMap = new Map<string, string>();
		for (const n of clip.nodes) {
			idMap.set(n.id, uid());
		}
		const offset = 40;
		const newNodes = clip.nodes.map((n) => ({
			...n,
			id: idMap.get(n.id)!,
			position: { x: n.position.x + offset, y: n.position.y + offset },
			status: 'idle' as const,
			lastResult: undefined,
			lastError: undefined,
			userPositioned: true,
		}));
		const newEdges = clip.edges.map((e) => ({
			id: `e-${idMap.get(e.source)}-${idMap.get(e.target)}`,
			source: idMap.get(e.source)!,
			target: idMap.get(e.target)!,
		}));
		patchWorkspaces(set, get, (workspaces) =>
			workspaces.map((w) =>
				w.id === ws.id
					? {
							...w,
							nodes: [...w.nodes, ...newNodes],
							edges: [...w.edges, ...newEdges],
						}
					: w,
			),
		);
		set({
			selectedNodeIds: newNodes.map((n) => n.id),
			selectedNodeId: newNodes[0]?.id ?? null,
		});
	},

	addEdge: (source, target) => {
		const ws = get().getActiveWorkspace();
		if (!ws || source === target) return false;
		const edgeId = `e-${source}-${target}`;
		if (ws.edges.some((e) => e.id === edgeId)) return false;
		patchWorkspaces(set, get, (workspaces) =>
			workspaces.map((w) =>
				w.id === ws.id
					? { ...w, edges: [...w.edges, { id: edgeId, source, target }] }
					: w,
			),
		);
		return true;
	},

	collapseNodes: (nodeIds) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		const collapsed = new Set(ws.collapsedNodeIds ?? []);
		for (const id of nodeIds) collapsed.add(id);
		patchWorkspaces(set, get, (workspaces) =>
			workspaces.map((w) =>
				w.id === ws.id ? { ...w, collapsedNodeIds: [...collapsed] } : w,
			),
		);
	},

	expandNodes: (nodeIds) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		const remove = new Set(nodeIds);
		const collapsed = (ws.collapsedNodeIds ?? []).filter(
			(id) => !remove.has(id),
		);
		patchWorkspaces(set, get, (workspaces) =>
			workspaces.map((w) =>
				w.id === ws.id ? { ...w, collapsedNodeIds: collapsed } : w,
			),
		);
	},

	toggleNodeDetailExpand: (nodeId) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		const expanded = new Set(ws.expandedDetailNodeIds ?? []);
		if (expanded.has(nodeId)) expanded.delete(nodeId);
		else expanded.add(nodeId);
		patchWorkspaces(set, get, (workspaces) =>
			workspaces.map((w) =>
				w.id === ws.id ? { ...w, expandedDetailNodeIds: [...expanded] } : w,
			),
		);
	},

	toggleNodeSubtreeCollapse: (nodeId) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		const collapsed = new Set(ws.collapsedNodeIds ?? []);
		if (collapsed.has(nodeId)) collapsed.delete(nodeId);
		else collapsed.add(nodeId);
		patchWorkspaces(set, get, (workspaces) =>
			workspaces.map((w) =>
				w.id === ws.id ? { ...w, collapsedNodeIds: [...collapsed] } : w,
			),
		);
	},

	expandAllNodes: () => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		patchWorkspaces(set, get, (workspaces) =>
			workspaces.map((w) =>
				w.id === ws.id ? { ...w, collapsedNodeIds: [] } : w,
			),
		);
	},

	collapseAllNodes: () => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		const roots = ws.nodes
			.filter((n) => !ws.edges.some((e) => e.target === n.id))
			.map((n) => n.id);
		patchWorkspaces(set, get, (workspaces) =>
			workspaces.map((w) =>
				w.id === ws.id ? { ...w, collapsedNodeIds: roots } : w,
			),
		);
	},

	deleteSelectedNodes: () => {
		const ws = get().getActiveWorkspace();
		const ids = get().selectedNodeIds;
		if (!ws || ids.length === 0) return;
		const removeIds = new Set<string>();
		for (const id of ids) {
			removeIds.add(id);
			for (const d of getDescendantNodeIds(id, ws.edges)) {
				removeIds.add(d);
			}
		}
		const removeUrls = new Set(
			ws.nodes.filter((n) => removeIds.has(n.id)).map((n) => n.urlNormalized),
		);
		patchWorkspaces(set, get, (workspaces) =>
			workspaces.map((w) => {
				if (w.id !== ws.id) return w;
				return {
					...w,
					nodes: w.nodes.filter((n) => !removeIds.has(n.id)),
					edges: w.edges.filter(
						(e) => !removeIds.has(e.source) && !removeIds.has(e.target),
					),
					exclude_urls: w.exclude_urls.filter((u) => !removeUrls.has(u)),
				};
			}),
		);
		set({
			selectedNodeId: null,
			selectedNodeIds: [],
			showDeleteNodeDialog: false,
		});
	},

	fetchSelectedNodeResult: async () => {
		const ws = get().getActiveWorkspace();
		const nodeId = get().selectedNodeId;
		if (!ws || !nodeId) return;
		const result = await scraperPort.getNodeResult(ws.id, nodeId);
		set({ loadedNodeResult: result });
	},

	previewSelectedResults: async () => {
		const ws = get().getActiveWorkspace();
		const ids = get().selectedNodeIds;
		if (!ws || ids.length === 0) return;
		const previews = await scraperPort.getNodeResults(ws.id, ids);
		set({ resultPreview: previews });
	},

	mergeAllResults: async () => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		const res = await scraperPort.mergeResults(ws.id, null);
		set({ mergeSheetOpen: true, mergeSheetContent: res.merged });
	},

	mergeSelectedResults: async () => {
		const ws = get().getActiveWorkspace();
		const ids = get().selectedNodeIds;
		if (!ws || ids.length === 0) return;
		const res = await scraperPort.mergeResults(ws.id, ids);
		set({ mergeSheetOpen: true, mergeSheetContent: res.merged });
	},

	saveSelectedResults: async () => {
		const ws = get().getActiveWorkspace();
		const ids = get().selectedNodeIds;
		if (!ws || ids.length === 0) return;
		await scraperPort.saveResults(ws.id, ids);
	},

	deleteSelectedResults: async () => {
		const ws = get().getActiveWorkspace();
		const ids = get().selectedNodeIds;
		if (!ws || ids.length === 0) return;
		await scraperPort.deleteResults(ws.id, ids);
		set({ loadedNodeResult: null, resultPreview: null });
	},

	bulkScrapeSelected: async () => {
		const ids = get().selectedNodeIds;
		if (ids.length === 0) return;
		await get().startCrawl({ mode: 4, nodeIds: ids });
	},

	fetchWorkspaceDiff: async (workspaceId) => {
		const diff = await scraperPort.getWorkspaceDiff(workspaceId);
		set((s) => ({
			workspaceDiffCache: { ...s.workspaceDiffCache, [workspaceId]: diff },
		}));
		return diff;
	},

	closeMergeSheet: () =>
		set({ mergeSheetOpen: false, mergeSheetContent: null }),

	getActiveWorkspace: () => {
		const { workspaces, activeWorkspaceId } = get();
		return workspaces.find((w) => w.id === activeWorkspaceId) ?? null;
	},

	getSelectedNode: () => {
		const ws = get().getActiveWorkspace();
		const id = get().selectedNodeId;
		if (!ws || !id) return null;
		return ws.nodes.find((n) => n.id === id) ?? null;
	},
}));
