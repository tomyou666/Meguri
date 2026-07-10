import {
	Bot,
	BotOff,
	ChevronDown,
	ChevronRight,
	CircleAlert,
	Loader2,
	RefreshCw,
} from 'lucide-react';
import {
	type ReactNode,
	useCallback,
	useEffect,
	useMemo,
	useRef,
	useState,
} from 'react';
import { ActionTooltip } from '@/components/ui/action-tooltip';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
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
import {
	loadRobotsCache,
	type RobotsInfo,
	type RobotsStatus,
	saveRobotsCacheEntry,
} from '@/lib/robotsCache';
import { cn } from '@/lib/utils';
import type { PartialConfig } from '@/types/config';
import type { GraphNode } from '@/types/graph';
import * as ScraperService from '../../../bindings/meguri-app/internal/usecase/wails_service/scraperservice';

/** host 単位のセッションキャッシュ。WS 切替・パネル unmount でも保持する。 */
let robotsSessionCache: Record<string, RobotsInfo> = loadRobotsCache();
const robotsFetchGenByHost: Record<string, number> = {};

function readRobotsSessionCache(): Record<string, RobotsInfo> {
	return robotsSessionCache;
}

function writeRobotsSessionCache(
	updater: (prev: Record<string, RobotsInfo>) => Record<string, RobotsInfo>,
): Record<string, RobotsInfo> {
	robotsSessionCache = updater(robotsSessionCache);
	return robotsSessionCache;
}

type DomainStatusPanelProps = {
	nodes: GraphNode[];
	appDefaults: PartialConfig;
	wsSettings: PartialConfig;
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

function isRobotsLoading(robots?: RobotsInfo): boolean {
	return !robots || robots.status === 'loading';
}

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

export function DomainStatusPanel({
	nodes,
	appDefaults,
	wsSettings,
}: DomainStatusPanelProps) {
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
	const [robotsByHost, setRobotsByHost] = useState(readRobotsSessionCache);
	const robotsByHostRef = useRef(robotsByHost);
	robotsByHostRef.current = robotsByHost;

	const commitRobotsCache = useCallback(
		(
			updater: (prev: Record<string, RobotsInfo>) => Record<string, RobotsInfo>,
		) => {
			const next = writeRobotsSessionCache(updater);
			robotsByHostRef.current = next;
			setRobotsByHost(next);
		},
		[],
	);

	const fetchTargetsRef = useRef(fetchTargets);
	fetchTargetsRef.current = fetchTargets;
	const configRef = useRef({ appDefaults, wsSettings });
	configRef.current = { appDefaults, wsSettings };

	const fetchRobotsForHosts = useCallback(
		(hostsToFetch: string[]) => {
			if (hostsToFetch.length === 0) return;

			const targets = fetchTargetsRef.current;
			const gens: Record<string, number> = {};
			for (const host of hostsToFetch) {
				const nextGen = (robotsFetchGenByHost[host] ?? 0) + 1;
				robotsFetchGenByHost[host] = nextGen;
				gens[host] = nextGen;
			}

			commitRobotsCache((prev) => {
				const next = { ...prev };
				for (const host of hostsToFetch) {
					next[host] = { status: 'loading' };
				}
				return next;
			});

			const { appDefaults: defaults, wsSettings: settings } = configRef.current;

			void Promise.allSettled(
				hostsToFetch.map(async (host) => {
					const baseURL = targets.get(host);
					if (!baseURL) {
						return { host, info: null as RobotsInfo | null };
					}
					try {
						const info = await ScraperService.FetchRobotsTxt(
							host,
							baseURL,
							defaults,
							settings,
						);
						return {
							host,
							info: {
								status: info.status as RobotsStatus,
								statusCode: info.statusCode,
								body: info.body,
								error: info.error,
							} satisfies RobotsInfo,
						};
					} catch (err) {
						return {
							host,
							info: {
								status: 'error' as const,
								error: String(err),
							} satisfies RobotsInfo,
						};
					}
				}),
			).then((results) => {
				commitRobotsCache((prev) => {
					const next = { ...prev };
					let changed = false;
					for (const result of results) {
						if (result.status !== 'fulfilled') continue;
						const { host, info } = result.value;
						if (gens[host] !== robotsFetchGenByHost[host]) continue;
						if (!info) continue;
						next[host] = info;
						saveRobotsCacheEntry(host, info);
						changed = true;
					}
					return changed ? next : prev;
				});
			});
		},
		[commitRobotsCache],
	);

	useEffect(() => {
		// セッションキャッシュを表示に同期（unmount 後の再 mount 含む）
		setRobotsByHost(readRobotsSessionCache());
		robotsByHostRef.current = readRobotsSessionCache();

		if (fetchTargets.size === 0) return;

		const hostsToFetch = [...fetchTargets.keys()].filter(
			(host) => !isRobotsCacheHit(readRobotsSessionCache()[host]),
		);
		fetchRobotsForHosts(hostsToFetch);
	}, [fetchTargets, fetchRobotsForHosts]);

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
				const loading = isRobotsLoading(robots);

				return (
					<div key={host} className='rounded-md'>
						<div
							className={cn(
								'flex w-full items-start gap-0.5 rounded-md px-1 py-0.5',
								expanded && 'bg-sidebar-accent',
							)}
						>
							<button
								type='button'
								className='flex min-w-0 flex-1 items-start gap-1 rounded-md px-1 py-1 text-left text-xs hover:bg-sidebar-accent'
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
								<RobotsStatusIcon robots={robots} />
							</button>
							<ActionTooltip label={messages.domainStatus.robotsRefresh}>
								<span className='mt-0.5 inline-flex shrink-0'>
									<Button
										variant='ghost'
										size='icon-xs'
										disabled={loading}
										aria-label={messages.domainStatus.robotsRefresh}
										onClick={() => fetchRobotsForHosts([host])}
									>
										<RefreshCw
											className={cn('size-3', loading && 'animate-spin')}
										/>
									</Button>
								</span>
							</ActionTooltip>
						</div>
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

function RobotsStatusIcon({ robots }: { robots?: RobotsInfo }) {
	let label: string;
	let icon: ReactNode;

	if (!robots || robots.status === 'loading') {
		label = messages.domainStatus.robotsLoading;
		icon = <Loader2 className='size-3.5 animate-spin text-muted-foreground' />;
	} else if (robots.status === 'found') {
		label = messages.domainStatus.robotsFound;
		icon = <Bot className='size-3.5 text-foreground' />;
	} else if (robots.status === 'not_found') {
		label = messages.domainStatus.robotsNotFound;
		icon = <BotOff className='size-3.5 text-muted-foreground' />;
	} else {
		label = messages.domainStatus.robotsError;
		icon = <CircleAlert className='size-3.5 text-destructive' />;
	}

	return (
		<ActionTooltip label={label}>
			<span
				className='mt-0.5 inline-flex shrink-0'
				aria-label={label}
				role='img'
			>
				{icon}
			</span>
		</ActionTooltip>
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
