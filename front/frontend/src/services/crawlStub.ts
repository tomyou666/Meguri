import { getForwardReachableExisting, getOutgoingEdges } from '@/lib/graph';
import { configForMode2, mergeConfig } from '@/lib/mergeConfig';
import { hostFromUrl } from '@/lib/normalizeUrl';
import type { AppConfig, PartialConfig } from '@/types/config';
import type {
	CrawlEventHandlers,
	CrawlResultPreview,
	CrawlStubOptions,
	RunMode,
} from '@/types/crawl';
import type { GraphEdge, GraphNode } from '@/types/graph';
import type { Workspace } from '@/types/workspace';

function randomDelay(): Promise<void> {
	const ms = 300 + Math.floor(Math.random() * 401);
	return new Promise((r) => setTimeout(r, ms));
}

function mockResult(url: string, formats: string[]): CrawlResultPreview {
	const result: CrawlResultPreview = { url };
	if (formats.includes('markdown')) {
		result.markdown = `# Mock page\n\nContent for \`${url}\``;
	}
	if (formats.includes('links')) {
		result.links = [
			`${url.replace(/\/$/, '')}/child-a`,
			`${url.replace(/\/$/, '')}/child-b`,
		];
	}
	if (formats.includes('metadata')) {
		result.metadata = {
			title: 'Mock Title',
			description: 'Mock description',
		};
	}
	return result;
}

function isExcluded(url: string, excludeUrls: string[]): boolean {
	return excludeUrls.includes(url);
}

function buildVisitOrder(
	mode: RunMode,
	workspace: Workspace,
	startNodeId: string | undefined,
	seedUrl: string,
): { nodeIds: string[]; startId: string | null } {
	const nodes = workspace.nodes;
	const edges = workspace.edges;

	if (mode === 1) {
		const seedNode = nodes.find((n) => n.urlNormalized === seedUrl);
		if (seedNode) {
			return {
				nodeIds: bfsFrom(seedNode.id, nodes, edges, true),
				startId: seedNode.id,
			};
		}
		return { nodeIds: [], startId: null };
	}

	if (!startNodeId) {
		return { nodeIds: [], startId: null };
	}

	if (mode === 2) {
		return {
			nodeIds: bfsFrom(startNodeId, nodes, edges, true),
			startId: startNodeId,
		};
	}

	// mode 3: existing nodes only, BFS forward
	const reachable = getForwardReachableExisting(startNodeId, nodes, edges);
	return {
		nodeIds: [startNodeId, ...reachable],
		startId: startNodeId,
	};
}

function bfsFrom(
	startId: string,
	nodes: GraphNode[],
	edges: GraphEdge[],
	discoverNew: boolean,
): string[] {
	const order: string[] = [];
	const visited = new Set<string>();
	const queue = [startId];
	visited.add(startId);

	while (queue.length > 0) {
		const current = queue.shift()!;
		order.push(current);
		for (const edge of getOutgoingEdges(current, edges)) {
			if (!visited.has(edge.target)) {
				visited.add(edge.target);
				queue.push(edge.target);
			}
		}
	}

	if (!discoverNew) return order;
	return order.filter((id) => nodes.some((n) => n.id === id));
}

function resolveConfig(
	mode: RunMode,
	appDefaults: PartialConfig,
	workspace: Workspace,
	node: GraphNode,
): AppConfig {
	if (mode === 2) {
		return configForMode2(appDefaults);
	}
	const host = hostFromUrl(node.urlNormalized);
	const domain = workspace.domainSettings[host];
	return mergeConfig(
		appDefaults,
		workspace.settings,
		domain,
		node.nodeSettings,
	);
}

function mockDiscoverLinks(
	result: CrawlResultPreview,
	parentId: string,
	workspace: Workspace,
	mode: RunMode,
	handlers: CrawlEventHandlers,
): string[] {
	if (mode === 3 || !result.links?.length) return [];
	const toEnqueue: string[] = [];
	const parent = workspace.nodes.find((n) => n.id === parentId);
	if (!parent) return toEnqueue;

	result.links.forEach((link, i) => {
		try {
			const existing = workspace.nodes.find((n) => n.urlNormalized === link);
			if (existing) {
				const edgeId = `e-${parentId}-${existing.id}`;
				if (!workspace.edges.some((e) => e.id === edgeId)) {
					handlers.onEdgeDiscovered(parentId, existing.id, link);
				}
				toEnqueue.push(existing.id);
				return;
			}
			const id = `n-${Date.now()}-${i}`;
			handlers.onEdgeDiscovered(parentId, id, link);
			toEnqueue.push(id);
		} catch {
			// invalid link URL in mock
		}
	});
	return toEnqueue;
}

export async function runCrawlStub(
	workspace: Workspace,
	appDefaults: PartialConfig,
	seedUrl: string,
	handlers: CrawlEventHandlers,
	options: CrawlStubOptions,
): Promise<void> {
	const { mode, startNodeId, signal, waitWhilePaused, debugScenario } = options;

	if (debugScenario === 'global_fail') {
		handlers.onCrawlError('デバッグ: 全体エラー');
		return;
	}

	const { nodeIds, startId } = options.nodeIds?.length
		? { nodeIds: options.nodeIds, startId: options.nodeIds[0] ?? null }
		: buildVisitOrder(mode, workspace, startNodeId, seedUrl);

	if (mode !== 1 && !startId) {
		handlers.onCrawlError('開始ノードが選択されていません');
		return;
	}

	let enqueued = 0;
	let succeeded = 0;
	let failed = 0;
	let skipped = 0;
	const maxPages =
		mode === 2
			? (configForMode2(appDefaults).crawl?.max_pages ?? 100)
			: (mergeConfig(appDefaults, workspace.settings).crawl?.max_pages ?? 100);

	const queue = [...nodeIds];
	const visited = new Set<string>();
	let processed = 0;

	while (queue.length > 0) {
		if (signal.aborted) {
			handlers.onCrawlCompleted({
				mode,
				finishedAt: new Date().toISOString(),
				enqueued,
				succeeded,
				failed,
				skipped,
				stoppedReason: 'stopped',
			});
			return;
		}

		await waitWhilePaused();
		if (signal.aborted) {
			handlers.onCrawlCompleted({
				mode,
				finishedAt: new Date().toISOString(),
				enqueued,
				succeeded,
				failed,
				skipped,
				stoppedReason: 'stopped',
			});
			return;
		}

		const nodeId = queue.shift()!;
		if (visited.has(nodeId)) continue;
		visited.add(nodeId);

		const ws = options.getWorkspace();
		const node = ws.nodes.find((n) => n.id === nodeId);
		if (!node) continue;

		if (isExcluded(node.urlNormalized, ws.exclude_urls)) {
			skipped++;
			handlers.onNodeSkipped(nodeId, node.urlNormalized, 'exclude_urls');
			continue;
		}

		if (processed >= maxPages) break;
		processed++;
		enqueued++;

		handlers.onNodeStarted(nodeId, node.urlNormalized);
		await randomDelay();

		if (debugScenario === 'stop_mid' && processed === 2) {
			handlers.onCrawlCompleted({
				mode,
				finishedAt: new Date().toISOString(),
				enqueued,
				succeeded,
				failed,
				skipped,
				stoppedReason: 'stopped',
			});
			return;
		}

		const cfg = resolveConfig(mode, appDefaults, ws, node);
		const formats = cfg.content?.formats ?? ['markdown'];

		if (
			debugScenario === 'node_fail' &&
			node.urlNormalized === options.failNodeUrl
		) {
			failed++;
			handlers.onNodeFailed(
				nodeId,
				node.urlNormalized,
				'デバッグ: ノードエラー',
			);
			continue;
		}

		// mock robots skip
		const host = hostFromUrl(node.urlNormalized);
		const domainCfg = mergeConfig(
			appDefaults,
			ws.settings,
			ws.domainSettings[host],
		);
		if (
			domainCfg.crawl?.respect_robots_txt &&
			node.urlNormalized.includes('robots-block')
		) {
			skipped++;
			handlers.onNodeSkipped(nodeId, node.urlNormalized, 'robots');
			continue;
		}

		const result = mockResult(node.urlNormalized, formats);
		succeeded++;
		handlers.onNodeSucceeded(nodeId, result);

		if (mode !== 3) {
			const freshWs = options.getWorkspace();
			const toEnqueue = mockDiscoverLinks(
				result,
				nodeId,
				freshWs,
				mode,
				handlers,
			);
			for (const id of toEnqueue) {
				if (!visited.has(id) && queue.length + processed < maxPages) {
					queue.push(id);
				}
			}
		}
	}

	handlers.onCrawlCompleted({
		mode,
		finishedAt: new Date().toISOString(),
		enqueued,
		succeeded,
		failed,
		skipped,
		stoppedReason: 'completed',
	});
}
