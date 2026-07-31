import { messages } from '@/i18n/messages';
import { useAppStore } from '@/stores/appStore';

export type NodeContextMenuState = {
	nodeId: string;
	hasChildren: boolean;
	isCollapsed: boolean;
	crawlExclude: boolean;
};

type NodeContextMenuItemsProps = {
	state: NodeContextMenuState;
	onClose: () => void;
};

/** ノード向けコンテキストメニュー項目（グラフ右クリック・ツリー Menu 共用）。 */
export function NodeContextMenuItems({
	state,
	onClose,
}: NodeContextMenuItemsProps) {
	const collapseNodes = useAppStore((s) => s.collapseNodes);
	const expandNodes = useAppStore((s) => s.expandNodes);
	const deleteSelectedNodes = useAppStore((s) => s.deleteSelectedNodes);
	const bulkScrapeSelected = useAppStore((s) => s.bulkScrapeSelected);
	const setNodeCrawlExclude = useAppStore((s) => s.setNodeCrawlExclude);

	return (
		<>
			{state.hasChildren && (
				<button
					type='button'
					className='block w-full px-2 py-1 text-left text-xs hover:bg-muted'
					onClick={() => {
						if (state.isCollapsed) {
							expandNodes([state.nodeId]);
						} else {
							collapseNodes([state.nodeId]);
						}
						onClose();
					}}
				>
					{state.isCollapsed
						? messages.graph.contextExpand
						: messages.graph.contextCollapse}
				</button>
			)}
			<button
				type='button'
				className='block w-full px-2 py-1 text-left text-xs hover:bg-muted'
				onClick={() => {
					setNodeCrawlExclude(state.nodeId, !state.crawlExclude);
					onClose();
				}}
			>
				{state.crawlExclude
					? messages.graph.contextIncludeCrawl
					: messages.graph.contextExcludeCrawl}
			</button>
			<button
				type='button'
				className='block w-full px-2 py-1 text-left text-xs hover:bg-muted'
				onClick={() => {
					void bulkScrapeSelected();
					onClose();
				}}
			>
				{messages.graph.contextScrape}
			</button>
			<button
				type='button'
				className='block w-full px-2 py-1 text-left text-destructive text-xs hover:bg-muted'
				onClick={() => {
					deleteSelectedNodes();
					onClose();
				}}
			>
				{messages.graph.contextDelete}
			</button>
		</>
	);
}
