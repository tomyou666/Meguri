import type { ComponentProps } from 'react';
import {
	formatOptionalNumber,
	parseOptionalNumber,
} from '@/components/settings/configFormUtils';
import { Input } from '@/components/ui/input';

const SPINNER_WIDTH_PX = 20;

type OptionalNumberInputProps = Omit<
	ComponentProps<typeof Input>,
	'type' | 'value' | 'onChange'
> & {
	value: number | undefined;
	onChange: (value: number | undefined) => void;
};

/** 制御コンポーネントの number input でスピンボタンが onChange を発火しない問題への対処 */
export function OptionalNumberInput({
	value,
	onChange,
	onClick,
	...props
}: OptionalNumberInputProps) {
	const handleValueChange = (raw: string) => {
		onChange(parseOptionalNumber(raw));
	};

	const handleClick = (e: React.MouseEvent<HTMLInputElement>) => {
		onClick?.(e);
		const input = e.currentTarget;
		const rect = input.getBoundingClientRect();
		if (e.clientX < rect.right - SPINNER_WIDTH_PX) return;

		const isUp = e.clientY - rect.top < rect.height / 2;
		try {
			if (isUp) input.stepUp();
			else input.stepDown();
		} catch {
			return;
		}
		onChange(parseOptionalNumber(input.value));
	};

	return (
		<Input
			type='number'
			step={1}
			value={formatOptionalNumber(value)}
			onChange={(e) => handleValueChange(e.target.value)}
			onClick={handleClick}
			{...props}
		/>
	);
}
