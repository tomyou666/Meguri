import { Copy, X } from 'lucide-react';
import {
	type ClipboardEvent,
	type KeyboardEvent,
	useRef,
	useState,
} from 'react';
import { tagListInputClassName } from '@/components/settings/configFormUtils';
import {
	addTokens,
	removeLastToken,
	removeTokenAt,
	tokensToCopyText,
} from '@/components/settings/tagListInputUtils';
import { ActionTooltip } from '@/components/ui/action-tooltip';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { messages } from '@/i18n/messages';
import { notifyError, notifySuccess } from '@/lib/notify';
import { cn } from '@/lib/utils';

type TagListInputProps = {
	values: string[];
	onChange: (values: string[]) => void;
	invalid?: boolean;
	compact?: boolean;
	placeholder?: string;
	removeLabel: (value: string) => string;
};

export function TagListInput({
	values,
	onChange,
	invalid = false,
	compact = false,
	placeholder,
	removeLabel,
}: TagListInputProps) {
	const [draft, setDraft] = useState('');
	const inputRef = useRef<HTMLInputElement>(null);
	const m = messages.settings.tagList;

	const commitDraft = () => {
		if (!draft.trim()) {
			setDraft('');
			return;
		}
		const next = addTokens(values, draft);
		if (next !== values) onChange(next);
		setDraft('');
	};

	const handlePaste = (e: ClipboardEvent<HTMLInputElement>) => {
		const text = e.clipboardData.getData('text');
		if (!/[\n\r]/.test(text)) return;

		e.preventDefault();
		const next = addTokens(values, draft + text);
		if (next !== values) onChange(next);
		setDraft('');
	};

	const handleCopy = async () => {
		try {
			await navigator.clipboard.writeText(tokensToCopyText(values));
			notifySuccess(m.copied);
		} catch (err) {
			notifyError(m.copyFailed, {
				description: err instanceof Error ? err.message : String(err),
			});
		}
	};

	const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
		if (e.key === 'Enter') {
			e.preventDefault();
			commitDraft();
			return;
		}
		if (e.key === 'Backspace' && draft === '' && values.length > 0) {
			e.preventDefault();
			onChange(removeLastToken(values));
		}
	};

	return (
		// biome-ignore lint/a11y/useKeyWithClickEvents: コンテナクリックで入力へフォーカス
		// biome-ignore lint/a11y/noStaticElementInteractions: 同上
		<div
			className={tagListInputClassName(invalid, compact)}
			onClick={() => inputRef.current?.focus()}
		>
			{values.map((value, index) => (
				<Badge
					key={value}
					variant='secondary'
					className={cn(
						'gap-0.5 py-0 font-normal',
						compact ? 'px-1 text-[10px]' : 'px-1.5 text-xs',
					)}
				>
					<ActionTooltip label={value}>
						<span className='max-w-40 truncate'>{value}</span>
					</ActionTooltip>
					<button
						type='button'
						className='inline-flex shrink-0 rounded-sm opacity-70 hover:opacity-100'
						aria-label={removeLabel(value)}
						onClick={(e) => {
							e.stopPropagation();
							onChange(removeTokenAt(values, index));
						}}
					>
						<X className={compact ? 'size-2.5' : 'size-3'} />
					</button>
				</Badge>
			))}
			<Input
				ref={inputRef}
				value={draft}
				placeholder={values.length === 0 ? placeholder : undefined}
				className={cn(
					'min-w-16 flex-1 border-0 bg-transparent px-1 shadow-none focus-visible:border-transparent focus-visible:ring-0',
					compact ? 'h-6 text-[10px]' : 'h-7 text-xs',
				)}
				onChange={(e) => setDraft(e.target.value)}
				onKeyDown={handleKeyDown}
				onPaste={handlePaste}
			/>
			{values.length > 0 && (
				<ActionTooltip label={m.copy}>
					<Button
						type='button'
						variant='ghost'
						size='icon-xs'
						className='shrink-0'
						aria-label={m.copy}
						onClick={(e) => {
							e.stopPropagation();
							void handleCopy();
						}}
					>
						<Copy className={compact ? 'size-2.5' : 'size-3.5'} />
					</Button>
				</ActionTooltip>
			)}
		</div>
	);
}
