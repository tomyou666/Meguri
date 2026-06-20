import { useEffect } from 'react';
import { Group, Panel, Separator, usePanelRef } from 'react-resizable-panels';
import { CrawlGraph } from '@/components/graph/CrawlGraph';
import { TooltipProvider } from '@/components/ui/tooltip';
import { useAppStore } from '@/stores/appStore';
import { AppBootstrap } from './AppBootstrap';
import { AppDialogs } from './AppDialogs';
import { AppKeyboardShortcuts } from './AppKeyboardShortcuts';
import { ControlBar } from './ControlBar';
import { LeftSidebarContent } from './LeftSidebar';
import { MenuBar } from './MenuBar';
import { MergeSheet } from './MergeSheet';
import { RightSidebarContent } from './RightSidebar';

const LEFT_SIDEBAR_SIZE = '14rem';
const RIGHT_SIDEBAR_SIZE = '16rem';

function SidebarPanels() {
	const leftCollapsed = useAppStore((s) => s.leftSidebarCollapsed);
	const rightCollapsed = useAppStore((s) => s.rightSidebarCollapsed);
	const leftRef = usePanelRef();
	const rightRef = usePanelRef();

	useEffect(() => {
		const panel = leftRef.current;
		if (!panel) return;
		if (leftCollapsed) panel.collapse();
		else panel.expand();
	}, [leftCollapsed, leftRef]);

	useEffect(() => {
		const panel = rightRef.current;
		if (!panel) return;
		if (rightCollapsed) panel.collapse();
		else panel.expand();
	}, [rightCollapsed, rightRef]);

	return (
		<>
			<Panel
				id='left-sidebar'
				panelRef={leftRef}
				defaultSize={LEFT_SIDEBAR_SIZE}
				minSize={LEFT_SIDEBAR_SIZE}
				collapsible
				collapsedSize='2.75rem'
				className='min-w-0'
			>
				<LeftSidebarContent />
			</Panel>
			<Separator className='w-1 shrink-0 bg-border hover:bg-primary/30' />
			<Panel id='main-canvas' minSize='30%' className='min-w-0'>
				<main className='relative flex h-full min-w-0 flex-col'>
					<CrawlGraph />
				</main>
			</Panel>
			<Separator className='w-1 shrink-0 bg-border hover:bg-primary/30' />
			<Panel
				id='right-sidebar'
				panelRef={rightRef}
				defaultSize={RIGHT_SIDEBAR_SIZE}
				minSize={RIGHT_SIDEBAR_SIZE}
				// maxSize='32rem'
				collapsible
				collapsedSize='2.75rem'
				className='min-w-0'
			>
				<RightSidebarContent />
			</Panel>
		</>
	);
}

export function AppShell() {
	return (
		<AppBootstrap>
			<TooltipProvider>
				<AppKeyboardShortcuts />
				<div className='flex h-screen w-full flex-col overflow-hidden'>
					<MenuBar />
					<ControlBar />
					<Group orientation='horizontal' className='min-h-0 w-full flex-1'>
						<SidebarPanels />
					</Group>
					<AppDialogs />
					<MergeSheet />
				</div>
			</TooltipProvider>
		</AppBootstrap>
	);
}
