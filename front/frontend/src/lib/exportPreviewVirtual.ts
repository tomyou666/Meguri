/** プレビュー仮想行の高さ見積もりと本文窓。 */

export const EXPORT_PREVIEW_TARGET_BYTES = 8192;
export const EXPORT_PREVIEW_BASE_ESTIMATE_PX = 240;
export const EXPORT_PREVIEW_MIN_ESTIMATE_PX = 64;
export const EXPORT_PREVIEW_FIRST_ROW_EXTRA_PX = 48;

export type ExportPreviewRowMeta = {
	id: string;
	nodeId: string;
	byteLength: number;
	isFirst: boolean;
};

export type ExportPreviewRowBody = {
	id: string;
	nodeId: string;
	url: string;
	body: string;
};

export type BodyWindow = {
	start: number;
	end: number;
};

/** 未計測行の高さ見積もり。 */
export function estimateExportRowHeight(
	byteLength: number,
	isFirst: boolean,
): number {
	const base = Math.max(
		EXPORT_PREVIEW_MIN_ESTIMATE_PX,
		(EXPORT_PREVIEW_BASE_ESTIMATE_PX * byteLength) /
			EXPORT_PREVIEW_TARGET_BYTES,
	);
	return isFirst ? base + EXPORT_PREVIEW_FIRST_ROW_EXTRA_PX : base;
}

/** 行 index のサイズ（測り済み優先）。 */
export function resolveExportRowSize(
	index: number,
	rows: ExportPreviewRowMeta[],
	measured: Map<number, number>,
): number {
	const measuredSize = measured.get(index);
	if (measuredSize != null) return measuredSize;
	const row = rows[index];
	if (!row) return EXPORT_PREVIEW_MIN_ESTIMATE_PX;
	return estimateExportRowHeight(row.byteLength, row.isFirst);
}

/**
 * 見えている範囲 + overscan の前後 1 画面分の本文窓。
 *
 * overscan 行も含めて本文を持つ。
 */
export function computeBodyWindow(
	rangeStart: number,
	rangeEnd: number,
	rowCount: number,
	viewportRowSpan: number,
): BodyWindow {
	const span = Math.max(1, viewportRowSpan);
	const start = Math.max(0, rangeStart - span);
	const end = Math.min(rowCount, rangeEnd + span);
	return { start, end };
}

/**
 * 現在の本文窓から外れたときだけ再取得が必要。
 *
 * 窓の端に触れただけでは取らない。
 */
export function shouldFetchBodies(
	needed: BodyWindow,
	held: BodyWindow | null,
): boolean {
	if (!held) return needed.end > needed.start;
	if (needed.start >= held.start && needed.end <= held.end) return false;
	return needed.end > needed.start;
}

/** 範囲外の本文を Map から削除する。 */
export function pruneBodyMap(
	bodies: Map<number, ExportPreviewRowBody>,
	window: BodyWindow,
): Map<number, ExportPreviewRowBody> {
	const next = new Map<number, ExportPreviewRowBody>();
	for (const [index, body] of bodies) {
		if (index >= window.start && index < window.end) {
			next.set(index, body);
		}
	}
	return next;
}

/**
 * 本文応答を採用するか判定する。
 *
 * 世代一致、範囲一致、各行 id が要求範囲のメタに含まれるときだけ採用。
 */
export function acceptBodyResponse(args: {
	responseGeneration: number;
	currentGeneration: number;
	responseStart: number;
	responseEnd: number;
	requestedStart: number;
	requestedEnd: number;
	rows: ExportPreviewRowMeta[];
	responseRows: ExportPreviewRowBody[];
}): boolean {
	if (args.responseGeneration !== args.currentGeneration) return false;
	if (
		args.responseStart !== args.requestedStart ||
		args.responseEnd !== args.requestedEnd
	) {
		return false;
	}
	const allowed = new Set<string>();
	for (let i = args.requestedStart; i < args.requestedEnd; i++) {
		const meta = args.rows[i];
		if (meta) allowed.add(meta.id);
	}
	for (const body of args.responseRows) {
		if (!allowed.has(body.id)) return false;
	}
	return true;
}
