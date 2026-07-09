import { useState } from 'react';
import {
	inputClassName,
	selectClassName,
} from '@/components/settings/configFormUtils';
import { Input } from '@/components/ui/input';
import { messages } from '@/i18n/messages';
import {
	LOCALE_PRESET_CUSTOM,
	LOCALE_PRESETS,
	type LocalePresetField,
	localePresetOptionValue,
	resolveLocaleSelectValue,
} from '@/lib/localePresets';

type LocalePresetSelectProps = {
	field: LocalePresetField;
	value: string | undefined;
	onChange: (value: string) => void;
	invalid?: boolean;
};

/** ロケールプリセットのセレクト。カスタム時のみ下に自由入力を出す。 */
export function LocalePresetSelect({
	field,
	value,
	onChange,
	invalid,
}: LocalePresetSelectProps) {
	// プリセット値のまま「カスタム」を選んだとき、厳密一致でセレクトが戻らないようにする
	const [forceCustom, setForceCustom] = useState(false);
	const matched = resolveLocaleSelectValue(value, field);
	const isCustom = forceCustom || matched === LOCALE_PRESET_CUSTOM;
	const selectValue = isCustom ? LOCALE_PRESET_CUSTOM : matched;
	const labels = messages.settings.localePresets;

	return (
		<div className='space-y-1'>
			<select
				className={selectClassName(!!invalid)}
				value={selectValue}
				onChange={(e) => {
					const next = e.target.value;
					if (next === LOCALE_PRESET_CUSTOM) {
						setForceCustom(true);
						// 現在値を残してカスタム編集へ（空なら空のまま）
						onChange(value ?? '');
						return;
					}
					setForceCustom(false);
					onChange(next);
				}}
			>
				<option value=''>{labels.unset}</option>
				{LOCALE_PRESETS.map((p) => (
					<option key={p.id} value={localePresetOptionValue(p, field)}>
						{labels.countries[p.id as keyof typeof labels.countries]}
					</option>
				))}
				<option value={LOCALE_PRESET_CUSTOM}>{labels.custom}</option>
			</select>
			{isCustom ? (
				<Input
					className={inputClassName(!!invalid, 'mt-1 h-8')}
					value={value ?? ''}
					onChange={(e) => onChange(e.target.value)}
				/>
			) : null}
		</div>
	);
}
