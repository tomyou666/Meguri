import { useEffect, useState } from 'react';
import {
	inputClassName,
	selectClassName,
} from '@/components/settings/configFormUtils';
import {
	type DurationUnit,
	formatDurationForSave,
	parseDurationForForm,
} from '@/components/settings/durationFormUtils';
import { OptionalNumberInput } from '@/components/settings/OptionalNumberInput';
import { messages } from '@/i18n/messages';
import { cn } from '@/lib/utils';

type DurationInputProps = {
	value?: string;
	onChange: (value: string | undefined) => void;
	invalid?: boolean;
	className?: string;
};

/** 整数入力 + s/ms セレクトで Go duration 文字列を編集する */
export function DurationInput({
	value,
	onChange,
	invalid = false,
	className,
}: DurationInputProps) {
	const parsed = parseDurationForForm(value ?? '');
	const [amount, setAmount] = useState(parsed.amount);
	const [unit, setUnit] = useState<DurationUnit>(parsed.unit);

	useEffect(() => {
		const next = parseDurationForForm(value ?? '');
		setAmount(next.amount);
		setUnit(next.unit);
	}, [value]);

	const emitChange = (
		nextAmount: number | undefined,
		nextUnit: DurationUnit,
	) => {
		onChange(formatDurationForSave(nextAmount, nextUnit));
	};

	return (
		<div className={cn('flex gap-1', className)}>
			<OptionalNumberInput
				className={inputClassName(invalid, 'mt-1 h-8 min-w-0 flex-1')}
				value={amount}
				onChange={(nextAmount) => {
					setAmount(nextAmount);
					emitChange(nextAmount, unit);
				}}
			/>
			<select
				className={selectClassName(invalid, 'mt-1 h-8 w-16 shrink-0')}
				value={unit}
				onChange={(e) => {
					const nextUnit = e.target.value as DurationUnit;
					setUnit(nextUnit);
					emitChange(amount, nextUnit);
				}}
				aria-label={messages.settings.units.label}
			>
				<option value='s'>{messages.settings.units.seconds}</option>
				<option value='ms'>{messages.settings.units.milliseconds}</option>
			</select>
		</div>
	);
}
