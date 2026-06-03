import { Button } from '@/components/ui/button';
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useAppStore } from '@/stores/appStore';

export function MergeSheet() {
	const open = useAppStore((s) => s.mergeSheetOpen);
	const content = useAppStore((s) => s.mergeSheetContent);
	const close = useAppStore((s) => s.closeMergeSheet);

	return (
		<Dialog open={open} onOpenChange={(o) => !o && close()}>
			<DialogContent className='max-h-[85vh] max-w-2xl'>
				<DialogHeader>
					<DialogTitle>マージ結果</DialogTitle>
				</DialogHeader>
				<ScrollArea className='max-h-[60vh]'>
					<pre className='whitespace-pre-wrap p-2 font-mono text-xs'>
						{content}
					</pre>
				</ScrollArea>
				<Button size='sm' onClick={close}>
					閉じる
				</Button>
			</DialogContent>
		</Dialog>
	);
}
