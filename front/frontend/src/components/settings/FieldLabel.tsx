import { Label } from '@/components/ui/label';

type FieldLabelProps = {
	label: string;
	help: string;
};

/** 設定項目のラベルと説明（誰が見ても意図が分かる文言） */
export function FieldLabel({ label, help }: FieldLabelProps) {
	return (
		<div className='space-y-0.5'>
			<Label className='text-xs font-medium'>{label}</Label>
			<p className='text-[10px] leading-snug text-muted-foreground'>{help}</p>
		</div>
	);
}
