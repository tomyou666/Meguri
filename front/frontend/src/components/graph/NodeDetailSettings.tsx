import { type RefObject, useEffect, useRef } from 'react';
import { ConfigEditor } from '@/components/settings/ConfigEditor';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { messages } from '@/i18n/messages';
import { useAppStore } from '@/stores/appStore';
import type { GraphNode } from '@/types/graph';

type NodeDetailSettingsProps = {
	node: GraphNode;
};

/** React Flow のズーム操作に wheel を渡さない */
function useStopWheelPropagation(ref: RefObject<HTMLElement | null>) {
	useEffect(() => {
		const el = ref.current;
		if (!el) return;
		const onWheel = (e: WheelEvent) => {
			e.stopPropagation();
		};
		el.addEventListener('wheel', onWheel, { passive: true, capture: true });
		return () => el.removeEventListener('wheel', onWheel, { capture: true });
	}, [ref]);
}

/** ノード詳細展開時にノード内へ表示する設定 */
export function NodeDetailSettings({ node }: NodeDetailSettingsProps) {
	const persistNodeSettings = useAppStore((s) => s.persistNodeSettings);
	const setNodeCrawlExclude = useAppStore((s) => s.setNodeCrawlExclude);
	const scrollRef = useRef<HTMLDivElement>(null);
	useStopWheelPropagation(scrollRef);

	return (
		<div
			ref={scrollRef}
			className='nodrag nopan nowheel max-h-[280px] overflow-y-auto overscroll-contain border-t border-border pt-2'
		>
			<Label className='mb-2 flex items-center gap-2 text-[10px]'>
				<Checkbox
					checked={node.crawlExclude}
					onCheckedChange={(c) => setNodeCrawlExclude(node.id, !!c)}
				/>
				{messages.right.crawlExclude}
			</Label>
			<ConfigEditor
				layer='node'
				settings={node.nodeSettings ?? {}}
				compact
				onSave={(settings) => persistNodeSettings(node.id, settings)}
			/>
		</div>
	);
}
