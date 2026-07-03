import { X } from 'lucide-react';
import { type KeyboardEvent, useRef, useState } from 'react';
import { tagListInputClassName } from '@/components/settings/configFormUtils';
import {
	addToken,
	normalizeToken,
	removeLastToken,
	removeTokenAt,
} from '@/components/settings/tagListInputUtils';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
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

	const commitDraft = () => {
		const token = normalizeToken(draft);
		if (!token) {
			setDraft('');
			return;
		}
		const next = addToken(values, token);
		if (next !== values) onChange(next);
		setDraft('');
	};

	const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
		if (e.key === 'Enter' || e.key === ',') {
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
					<span className='max-w-40 truncate'>{value}</span>
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
			/>
		</div>
	);
}
