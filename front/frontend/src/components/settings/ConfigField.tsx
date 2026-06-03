import type { ReactNode } from 'react';
import type { FieldErrors } from '@/components/settings/configFormUtils';
import { fieldError } from '@/components/settings/configFormUtils';
import { FieldLabel } from '@/components/settings/FieldLabel';

type ConfigFieldProps = {
	path: string;
	errors: FieldErrors;
	label: string;
	help: string;
	children: ReactNode;
};

/** ラベル・入力・直下のバリデーションエラー */
export function ConfigField({
	path,
	errors,
	label,
	help,
	children,
}: ConfigFieldProps) {
	const message = fieldError(errors, path);
	return (
		<div className='space-y-1'>
			<FieldLabel label={label} help={help} />
			{children}
			{message && (
				<p className='text-[10px] leading-snug text-destructive' role='alert'>
					{message}
				</p>
			)}
		</div>
	);
}
