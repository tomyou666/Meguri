import { ChevronDown, ChevronRight } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { messages } from '@/i18n/messages';
import {
	countNodesByStatus,
	domainStatusKey,
	groupNodesByHost,
	isRobotsCacheHit,
	type NodeStatusCounts,
	robotsTargetsFromNodes,
	robotsTargetsKey,
} from '@/lib/domainStats';
import { cn } from '@/lib/utils';
import type { GraphNode } from '@/types/graph';
import * as ScraperService from '../../../bindings/scraperbot-front/internal/usecase/wails_service/scraperservice';

type RobotsStatus = 'loading' | 'found' | 'not_found' | 'error';

type RobotsInfo = {
	status: RobotsStatus;
	statusCode?: number;
	body?: string;
	error?: string;
};

type DomainStatusPanelProps = {
	nodes: GraphNode[];
};

const STATUS_BADGE: Record<
	keyof NodeStatusCounts,
	'default' | 'secondary' | 'destructive' | 'outline'
> = {
	success: 'default',
	error: 'destructive',
	skipped: 'secondary',
	running: 'outline',
	idle: 'outline',
};

/** key が変わったときだけ value スナップショットを更新する。 */
function useSnapshotWhenKeyChanges<T>(value: T, key: string): T {
	const snapshotRef = useRef(value);
	const keyRef = useRef(key);
	if (keyRef.current !== key) {
		keyRef.current = key;
		snapshotRef.current = value;
	}
	return snapshotRef.current;
}

export function DomainStatusPanel({ nodes }: DomainStatusPanelProps) {
	const statusKey = useMemo(() => domainStatusKey(nodes), [nodes]);
	const statusNodes = useSnapshotWhenKeyChanges(nodes, statusKey);
	const hostNodes = useMemo(() => groupNodesByHost(statusNodes), [statusNodes]);
	const hosts = useMemo(() => [...hostNodes.keys()].sort(), [hostNodes]);

	const robotsTargets = useMemo(
		() => robotsTargetsFromNodes(statusNodes),
		[statusNodes],
	);
	const robotsTargetsKeyValue = useMemo(
		() => robotsTargetsKey(robotsTargets),
		[robotsTargets],
	);
	const fetchTargets = useSnapshotWhenKeyChanges(
		robotsTargets,
		robotsTargetsKeyValue,
	);

	const [expandedHost, setExpandedHost] = useState<string | null>(null);
	const [robotsByHost, setRobotsByHost] = useState<Record<string, RobotsInfo>>(
		{},
	);

	useEffect(() => {
		if (fetchTargets.size === 0) {
			setRobotsByHost({});
			return;
		}

		let cancelled = false;
		const hostsToFetch: string[] = [];

		setRobotsByHost((prev) => {
			const next: Record<string, RobotsInfo> = {};
			for (const host of fetchTargets.keys()) {
				if (isRobotsCacheHit(prev[host])) {
					next[host] = prev[host];
				} else {
					hostsToFetch.push(host);
					next[host] = { status: 'loading' };
				}
			}
			return next;
		});

		if (hostsToFetch.length === 0) {
			return;
		}

		void Promise.allSettled(
			hostsToFetch.map(async (host) => {
				const baseURL = fetchTargets.get(host)!;
				const info = await ScraperService.FetchRobotsTxt(host, baseURL);
				return { host, info };
			}),
		).then((results) => {
			if (cancelled) return;
			setRobotsByHost((prev) => {
				const next = { ...prev };
				for (const result of results) {
					if (result.status !== 'fulfilled') {
						continue;
					}
					const { host, info } = result.value;
					if (!fetchTargets.has(host)) continue;
					next[host] = {
						status: info.status as RobotsStatus,
						statusCode: info.statusCode,
						body: info.body,
						error: info.error,
					};
				}
				for (const host of Object.keys(next)) {
					if (!fetchTargets.has(host)) {
						delete next[host];
					}
				}
				return next;
			});
		});

		return () => {
			cancelled = true;
		};
	}, [fetchTargets]);

	if (hosts.length === 0) {
		return (
			<p className='px-2 py-2 text-xs text-muted-foreground'>
				{messages.sidebar.emptyDomains}
			</p>
		);
	}

	return (
		<div className='space-y-0.5 pb-2'>
			{hosts.map((host) => {
				const hostNodeList = hostNodes.get(host) ?? [];
				const counts = countNodesByStatus(hostNodeList);
				const robots = robotsByHost[host];
				const expanded = expandedHost === host;

				return (
					<div key={host} className='rounded-md'>
						<button
							type='button'
							className={cn(
								'flex w-full items-start gap-1 rounded-md px-2 py-1.5 text-left text-xs hover:bg-sidebar-accent',
								expanded && 'bg-sidebar-accent',
							)}
							onClick={() =>
								setExpandedHost((cur) => (cur === host ? null : host))
							}
							aria-expanded={expanded}
						>
							{expanded ? (
								<ChevronDown className='mt-0.5 size-3 shrink-0 text-muted-foreground' />
							) : (
								<ChevronRight className='mt-0.5 size-3 shrink-0 text-muted-foreground' />
							)}
							<span className='min-w-0 flex-1'>
								<span className='block truncate font-medium'>{host}</span>
								<span className='mt-0.5 block truncate text-[10px] text-muted-foreground'>
									{messages.domainStatus.statusSummary(counts)}
								</span>
							</span>
							<RobotsBadge robots={robots} />
						</button>
						{expanded && (
							<div className='space-y-2 px-2 pb-2 pt-1'>
								<div className='flex flex-wrap gap-1'>
									{(Object.keys(counts) as (keyof NodeStatusCounts)[]).map(
										(key) =>
											counts[key] > 0 ? (
												<Badge
													key={key}
													variant={STATUS_BADGE[key]}
													className='text-[10px]'
												>
													{messages.domainStatus.statusLabel(key, counts[key])}
												</Badge>
											) : null,
									)}
								</div>
								<RobotsDetail robots={robots} />
							</div>
						)}
					</div>
				);
			})}
		</div>
	);
}

function RobotsBadge({ robots }: { robots?: RobotsInfo }) {
	if (!robots || robots.status === 'loading') {
		return (
			<Badge variant='outline' className='shrink-0 text-[10px]'>
				{messages.domainStatus.robotsLoading}
			</Badge>
		);
	}
	if (robots.status === 'found') {
		return (
			<Badge variant='default' className='shrink-0 text-[10px]'>
				{messages.domainStatus.robotsFound}
			</Badge>
		);
	}
	if (robots.status === 'not_found') {
		return (
			<Badge variant='secondary' className='shrink-0 text-[10px]'>
				{messages.domainStatus.robotsNotFound}
			</Badge>
		);
	}
	return (
		<Badge variant='destructive' className='shrink-0 text-[10px]'>
			{messages.domainStatus.robotsError}
		</Badge>
	);
}

function RobotsDetail({ robots }: { robots?: RobotsInfo }) {
	if (!robots || robots.status === 'loading') {
		return (
			<p className='text-[10px] text-muted-foreground'>
				{messages.domainStatus.robotsLoading}
			</p>
		);
	}
	if (robots.status === 'found') {
		return (
			<pre className='max-h-40 overflow-auto rounded border border-border bg-muted/30 p-2 text-[10px] whitespace-pre-wrap break-all'>
				{robots.body || messages.domainStatus.robotsEmpty}
			</pre>
		);
	}
	if (robots.status === 'not_found') {
		return (
			<p className='text-[10px] text-muted-foreground'>
				{messages.domainStatus.robotsNotFoundDetail(robots.statusCode)}
			</p>
		);
	}
	return (
		<p className='text-[10px] text-destructive'>
			{messages.domainStatus.robotsErrorDetail(robots.error, robots.statusCode)}
		</p>
	);
}
