import { sortFlatData } from 'he-tree-react';
import type { CrawlResultPreview } from '@/types/crawl';
import type { GraphEdge, GraphNode } from '@/types/graph';

export type ExportMode = 'all' | 'selected';

export type ExportFlatNode = {
	id: string;
	parent_id: string | null;
	urlNormalized: string;
	label: string;
	status: string;
};

export type ExportHeadingField = 'url' | 'label';

export type ExportFormat = 'markdown' | 'html';

export type ExportMergeSettings = {
	format: ExportFormat;
	separator: string;
	includeHeading: boolean;
	headingField: ExportHeadingField;
};

export const DEFAULT_EXPORT_SEPARATOR = '\n\n---\n\n';

export const DEFAULT_EXPORT_SETTINGS: ExportMergeSettings = {
	format: 'markdown',
	separator: DEFAULT_EXPORT_SEPARATOR,
	includeHeading: true,
	headingField: 'url',
};

const FLAT_KEYS = { idKey: 'id' as const, parentIdKey: 'parent_id' as const };

/** 表示対象ノードを mode に応じて絞り込む。 */
export function filterExportNodes(
	nodes: GraphNode[],
	mode: ExportMode,
	selectedIds: string[],
): GraphNode[] {
	if (mode === 'selected') {
		const set = new Set(selectedIds);
		return nodes.filter((n) => set.has(n.id));
	}
	return nodes.filter((n) => n.status === 'success');
}

/** シードから BFS で親を決めたフラットツリーを構築する。 */
export function buildInitialFlatTree(
	nodes: GraphNode[],
	edges: GraphEdge[],
	seedUrl: string,
	mode: ExportMode,
	selectedIds: string[],
): ExportFlatNode[] {
	const visible = filterExportNodes(nodes, mode, selectedIds);
	if (visible.length === 0) return [];

	const visibleIds = new Set(visible.map((n) => n.id));
	const adj = new Map<string, string[]>();
	for (const e of edges) {
		if (!visibleIds.has(e.source) || !visibleIds.has(e.target)) continue;
		const list = adj.get(e.source) ?? [];
		list.push(e.target);
		adj.set(e.source, list);
	}

	const seedNode =
		visible.find((n) => n.urlNormalized === seedUrl) ??
		visible.find(
			(n) => !edges.some((e) => e.target === n.id && visibleIds.has(e.source)),
		);

	const parentById = new Map<string, string | null>();
	const queue: string[] = [];

	if (seedNode) {
		parentById.set(seedNode.id, null);
		queue.push(seedNode.id);
	}

	while (queue.length > 0) {
		const current = queue.shift();
		if (!current) continue;
		for (const child of adj.get(current) ?? []) {
			if (parentById.has(child)) continue;
			parentById.set(child, current);
			queue.push(child);
		}
	}

	for (const n of visible) {
		if (!parentById.has(n.id)) {
			parentById.set(n.id, null);
		}
	}

	const flat: ExportFlatNode[] = visible.map((n) => ({
		id: n.id,
		parent_id: parentById.get(n.id) ?? null,
		urlNormalized: n.urlNormalized,
		label: n.label,
		status: n.status,
	}));

	return sortFlatData(flat, FLAT_KEYS) as ExportFlatNode[];
}

/** フラットデータの全ノード ID を返す（初期チェック ON 用）。 */
export function initialCheckedIds(flatData: ExportFlatNode[]): string[] {
	return flatData.map((n) => n.id);
}

/** チェック ON のノードのみ深さ優先・前順で ID を返す。 */
export function preorderNodeIds(
	flatData: ExportFlatNode[],
	checkedIds: string[],
): string[] {
	const checked = new Set(checkedIds);
	const byParent = new Map<string | null, ExportFlatNode[]>();

	for (const node of flatData) {
		if (!checked.has(node.id)) continue;
		// 親が未チェックの子はルート扱いにして走査対象に含める
		const key =
			node.parent_id !== null && checked.has(node.parent_id)
				? node.parent_id
				: null;
		const list = byParent.get(key) ?? [];
		list.push(node);
		byParent.set(key, list);
	}

	const order: string[] = [];
	const walk = (parentId: string | null) => {
		for (const node of byParent.get(parentId) ?? []) {
			order.push(node.id);
			walk(node.id);
		}
	};
	walk(null);
	return order;
}

export type ExportNodeMeta = {
	id: string;
	urlNormalized: string;
	label: string;
};

export type MergeExportInput = {
	results: CrawlResultPreview[];
	nodeMeta: ExportNodeMeta[];
	settings: ExportMergeSettings;
};

function headingForNode(
	meta: ExportNodeMeta,
	settings: ExportMergeSettings,
): string {
	const text =
		settings.headingField === 'label' ? meta.label : meta.urlNormalized;
	return `## ${text}`;
}

function bodyForResult(
	result: CrawlResultPreview,
	format: ExportFormat,
): string {
	if (format === 'html') {
		return result.html ?? '';
	}
	return result.markdown ?? '';
}

/** マージ設定に従いエクスポート本文を連結する。 */
export function mergeExportContent(input: MergeExportInput): string {
	const { results, nodeMeta, settings } = input;
	const metaByUrl = new Map(nodeMeta.map((m) => [m.urlNormalized, m]));
	const parts: string[] = [];

	for (const result of results) {
		const body = bodyForResult(result, settings.format);
		if (!body) continue;
		const meta = metaByUrl.get(result.url);
		if (settings.includeHeading && meta) {
			parts.push(`${headingForNode(meta, settings)}\n\n${body}`);
		} else {
			parts.push(body);
		}
	}

	return parts.join(
		resolveExportSeparator(settings.separator, settings.format),
	);
}

/** 区切り文字のエスケープシーケンスを実際の制御文字に変換する。 */
export function parseExportSeparator(raw: string): string {
	return raw
		.replace(/\\r\\n/g, '\r\n')
		.replace(/\\n/g, '\n')
		.replace(/\\r/g, '\r')
		.replace(/\\t/g, '\t')
		.replace(/\\\\/g, '\\');
}

function escapeHtmlText(text: string): string {
	return text
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#39;');
}

/** 形式に応じた区切り文字を返す（HTML では XSS 防止のためエスケープ）。 */
export function resolveExportSeparator(
	raw: string,
	format: ExportFormat,
): string {
	const parsed = parseExportSeparator(raw);
	return format === 'html' ? escapeHtmlText(parsed) : parsed;
}

/** nodeId 配下のノード ID を深さ優先で返す（自身を含む）。 */
export function collectDescendantIds(
	flatData: ExportFlatNode[],
	nodeId: string,
): string[] {
	const byParent = new Map<string | null, ExportFlatNode[]>();
	for (const n of flatData) {
		const list = byParent.get(n.parent_id) ?? [];
		list.push(n);
		byParent.set(n.parent_id, list);
	}
	const out: string[] = [];
	const walk = (id: string) => {
		out.push(id);
		for (const child of byParent.get(id) ?? []) {
			walk(child.id);
		}
	};
	walk(nodeId);
	return out;
}

/** チェック状態を切り替える。ON は配下へ連動、OFF は配下のみ（親は維持）。 */
export function toggleExportNodeCheck(
	flatData: ExportFlatNode[],
	checkedIds: string[],
	nodeId: string,
	checked: boolean,
): string[] {
	const subtree = collectDescendantIds(flatData, nodeId);
	const next = new Set(checkedIds);
	if (checked) {
		for (const id of subtree) next.add(id);
	} else {
		for (const id of subtree) next.delete(id);
	}
	return [...next];
}

export type FetchExportPreviewResult = {
	content: string;
	skippedCount: number;
	includedCount: number;
};

/** DB 取得結果を順序どおりにマージし、結果なし件数を返す。 */
export function buildExportPreview(
	orderedIds: string[],
	flatData: ExportFlatNode[],
	results: CrawlResultPreview[],
	settings: ExportMergeSettings,
): FetchExportPreviewResult {
	const metaById = new Map(flatData.map((n) => [n.id, n]));
	const resultByUrl = new Map(results.map((r) => [r.url, r]));
	const orderedResults: CrawlResultPreview[] = [];
	const nodeMeta: ExportNodeMeta[] = [];
	let skippedCount = 0;

	for (const id of orderedIds) {
		const meta = metaById.get(id);
		if (!meta) continue;
		const result = resultByUrl.get(meta.urlNormalized);
		if (!result || !bodyForResult(result, settings.format)) {
			skippedCount += 1;
			continue;
		}
		orderedResults.push(result);
		nodeMeta.push({
			id: meta.id,
			urlNormalized: meta.urlNormalized,
			label: meta.label,
		});
	}

	return {
		content: mergeExportContent({
			results: orderedResults,
			nodeMeta,
			settings,
		}),
		skippedCount,
		includedCount: orderedResults.length,
	};
}
