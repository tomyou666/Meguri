import type { DiffKind, WorkspaceDiff } from '@/types/adapter';

export function diffNodeCount(diff: WorkspaceDiff | undefined): number {
	return diff?.nodes.length ?? 0;
}

export function summaryBadgeLabels(summary: WorkspaceDiff['summary']): {
	content: string;
	links: string;
	fetch: string;
} {
	return {
		content: `content ${summary.content}`,
		links: `links ${summary.links}`,
		fetch: `fetch ${summary.fetch}`,
	};
}

export function filterNodesByKind(
	nodes: WorkspaceDiff['nodes'],
	kind: DiffKind | 'all',
): WorkspaceDiff['nodes'] {
	if (kind === 'all') return nodes;
	return nodes.filter((n) => n.kinds.includes(kind));
}
