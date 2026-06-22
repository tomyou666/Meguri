import { textareaClassName } from '@/components/settings/configFormUtils';
import { Button } from '@/components/ui/button';
import { messages } from '@/i18n/messages';

type EditableTextResultProps = {
	value: string;
	editing: boolean;
	saving?: boolean;
	onChange: (value: string) => void;
	onSave: () => void;
	onCancel: () => void;
};

export function EditableTextResult({
	value,
	editing,
	saving = false,
	onChange,
	onSave,
	onCancel,
}: EditableTextResultProps) {
	if (!editing) {
		return (
			<pre className='whitespace-pre-wrap font-mono text-xs'>
				{value || '—'}
			</pre>
		);
	}

	return (
		<div className='flex min-h-0 flex-1 flex-col gap-2'>
			<textarea
				className={textareaClassName(false, 'min-h-48 flex-1 w-full')}
				value={value}
				onChange={(e) => onChange(e.target.value)}
			/>
			<div className='flex gap-1'>
				<Button size='xs' onClick={onSave} disabled={saving}>
					{messages.right.save}
				</Button>
				<Button
					size='xs'
					variant='outline'
					onClick={onCancel}
					disabled={saving}
				>
					{messages.dialog.cancel}
				</Button>
			</div>
		</div>
	);
}
