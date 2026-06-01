import { create } from 'zustand';
import {
	computeDagrePositions,
	type DagreLayoutDirection,
	fallbackNearParent,
	positionForDiscoveredNode,
} from '@/lib/dagreLayout';
import { DEFAULT_APP_CONFIG } from '@/lib/defaults';
import { collectDescendantUrls, getDescendantNodeIds } from '@/lib/graph';
import { hostFromUrl, normalizeUrl } from '@/lib/normalizeUrl';
import { runCrawlStub } from '@/services/crawlStub';
import type { PartialConfig } from '@/types/config';
import type {
	CrawlError,
	CrawlResultPreview,
	CrawlRunStatus,
	CrawlRunSummary,
	GlobalError,
	RunMode,
} from '@/types/crawl';
import type { GraphNode, NodeStatus } from '@/types/graph';
import type { Workspace } from '@/types/workspace';

function uid(): string {
	return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
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
		domainSettings: {},
	};
}

interface AppState {
	bootstrapped: boolean;
	appDefaults: PartialConfig;
	workspaces: Workspace[];
	activeWorkspaceId: string | null;
	selectedNodeId: string | null;
	selectedDomain: string | null;
	runMode: RunMode;
	crawlStatus: CrawlRunStatus;
	runHistory: CrawlRunSummary[];
	globalError: GlobalError;
	crawlError: CrawlError;
	showNewWorkspaceDialog: boolean;
	showAddNodeDialog: boolean;
	showDeleteNodeDialog: boolean;
	addNodeContextPosition: { x: number; y: number } | null;

	_abortController: AbortController | null;
	_paused: boolean;

	bootstrap: () => Promise<void>;
	setAppDefaults: (config: PartialConfig) => void;
	openNewWorkspaceDialog: () => void;
	closeNewWorkspaceDialog: () => void;
	createWorkspace: (name: string, seedUrl: string) => void;
	setActiveWorkspace: (id: string) => void;
	deleteWorkspace: (id: string) => void;
	selectNode: (id: string | null) => void;
	selectDomain: (host: string | null) => void;
	setRunMode: (mode: RunMode) => void;
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
	updateDomainSettings: (host: string, settings: PartialConfig) => void;
	setWorkspaceFormats: (formats: PartialConfig['content']) => void;
	clearGlobalError: () => void;
	clearCrawlError: () => void;
	startCrawl: () => Promise<void>;
	pauseCrawl: () => void;
	resumeCrawl: () => void;
	stopCrawl: () => void;

	getActiveWorkspace: () => Workspace | null;
	getSelectedNode: () => GraphNode | null;
	getDomains: () => string[];
}

export const useAppStore = create<AppState>((set, get) => ({
	bootstrapped: false,
	appDefaults: DEFAULT_APP_CONFIG,
	workspaces: [],
	activeWorkspaceId: null,
	selectedNodeId: null,
	selectedDomain: null,
	runMode: 1,
	crawlStatus: 'idle',
	runHistory: [],
	globalError: null,
	crawlError: null,
	showNewWorkspaceDialog: true,
	showAddNodeDialog: false,
	showDeleteNodeDialog: false,
	addNodeContextPosition: null,
	_abortController: null,
	_paused: false,

	bootstrap: async () => {
		const start = Date.now();
		await new Promise((r) => setTimeout(r, 150));
		const elapsed = Date.now() - start;
		if (elapsed < 400) {
			await new Promise((r) => setTimeout(r, 400 - elapsed));
		}
		set({ bootstrapped: true });
	},

	setAppDefaults: (config) => set({ appDefaults: config }),

	openNewWorkspaceDialog: () => set({ showNewWorkspaceDialog: true }),
	closeNewWorkspaceDialog: () => set({ showNewWorkspaceDialog: false }),

	createWorkspace: (name, seedUrl) => {
		try {
			const ws = emptyWorkspace(name, seedUrl);
			set((s) => ({
				workspaces: [...s.workspaces, ws],
				activeWorkspaceId: ws.id,
				selectedNodeId: ws.nodes[0]?.id ?? null,
				showNewWorkspaceDialog: false,
			}));
		} catch (e) {
			set({
				globalError: {
					type: 'global',
					message:
						e instanceof Error ? e.message : 'ワークスペース作成に失敗しました',
					at: new Date().toISOString(),
				},
			});
		}
	},

	setActiveWorkspace: (id) =>
		set({
			activeWorkspaceId: id,
			selectedNodeId: null,
			selectedDomain: null,
		}),

	deleteWorkspace: (id) =>
		set((s) => {
			const workspaces = s.workspaces.filter((w) => w.id !== id);
			const activeWorkspaceId =
				s.activeWorkspaceId === id
					? (workspaces[0]?.id ?? null)
					: s.activeWorkspaceId;
			return { workspaces, activeWorkspaceId };
		}),

	selectNode: (id) => set({ selectedNodeId: id, selectedDomain: null }),
	selectDomain: (host) => set({ selectedDomain: host, selectedNodeId: null }),
	setRunMode: (mode) => set({ runMode: mode }),

	updateNodePosition: (id, position) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		set((s) => ({
			workspaces: s.workspaces.map((w) =>
				w.id !== ws.id
					? w
					: {
							...w,
							nodes: w.nodes.map((n) =>
								n.id === id ? { ...n, position, userPositioned: true } : n,
							),
						},
			),
		}));
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
				nodeSettings: {},
				crawlExclude: false,
				status: 'idle',
			};
			set((s) => ({
				workspaces: s.workspaces.map((w) =>
					w.id === ws.id ? { ...w, nodes: [...w.nodes, node] } : w,
				),
				selectedNodeId: id,
				showAddNodeDialog: false,
				addNodeContextPosition: null,
			}));
		} catch (e) {
			set({
				globalError: {
					type: 'global',
					message: e instanceof Error ? e.message : 'URL が不正です',
					at: new Date().toISOString(),
				},
			});
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
		const hostsToCheck = new Set([...removeUrls].map((u) => hostFromUrl(u)));

		set((s) => ({
			workspaces: s.workspaces.map((w) => {
				if (w.id !== ws.id) return w;
				const nodes = w.nodes.filter((n) => !removeIds.has(n.id));
				const edges = w.edges.filter(
					(e) => !removeIds.has(e.source) && !removeIds.has(e.target),
				);
				const exclude_urls = w.exclude_urls.filter((u) => !removeUrls.has(u));
				const domainSettings = { ...w.domainSettings };
				for (const host of hostsToCheck) {
					if (!nodes.some((n) => hostFromUrl(n.urlNormalized) === host)) {
						delete domainSettings[host];
					}
				}
				return { ...w, nodes, edges, exclude_urls, domainSettings };
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

	updateDomainSettings: (host, settings) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		set((s) => ({
			workspaces: s.workspaces.map((w) =>
				w.id === ws.id
					? {
							...w,
							domainSettings: {
								...w.domainSettings,
								[host]: { ...w.domainSettings[host], ...settings },
							},
						}
					: w,
			),
		}));
	},

	setWorkspaceFormats: (content) => {
		const ws = get().getActiveWorkspace();
		if (!ws) return;
		set((s) => ({
			workspaces: s.workspaces.map((w) =>
				w.id === ws.id
					? {
							...w,
							settings: {
								...w.settings,
								content: { ...w.settings.content, ...content },
							},
						}
					: w,
			),
		}));
	},

	clearGlobalError: () => set({ globalError: null }),
	clearCrawlError: () => set({ crawlError: null }),

	pauseCrawl: () => set({ _paused: true, crawlStatus: 'paused' }),
	resumeCrawl: () => set({ _paused: false, crawlStatus: 'running' }),

	stopCrawl: () => {
		get()._abortController?.abort();
		set({ crawlStatus: 'idle', _paused: false });
	},

	startCrawl: async () => {
		const state = get();
		const ws = state.getActiveWorkspace();
		if (!ws) return;

		if (state.runMode !== 1 && !state.selectedNodeId) {
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
		});

		const runId = uid();
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

		await runCrawlStub(
			getWs(),
			state.appDefaults,
			ws.seedUrl,
			{
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
						runHistory: [full, ...s.runHistory].slice(0, 20),
					}));
				},
				onCrawlError: (message) => {
					set({
						crawlStatus: 'idle',
						_abortController: null,
						crawlError: {
							type: 'crawl',
							message,
							runId,
							at: new Date().toISOString(),
						},
					});
				},
			},
			{
				mode: state.runMode,
				startNodeId: state.selectedNodeId ?? undefined,
				workspaceId: ws.id,
				getWorkspace: () => get().getActiveWorkspace()!,
				signal: ac.signal,
				isPaused: () => get()._paused,
				waitWhilePaused: async () => {
					while (get()._paused && !ac.signal.aborted) {
						await new Promise((r) => setTimeout(r, 100));
					}
				},
			},
		);
	},

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

	getDomains: () => {
		const ws = get().getActiveWorkspace();
		if (!ws) return [];
		const hosts = new Set(ws.nodes.map((n) => hostFromUrl(n.urlNormalized)));
		return [...hosts].sort();
	},
}));
