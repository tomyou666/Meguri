import { useEffect, useMemo, useState } from 'react';
import { AppConfigTabs } from '@/components/settings/ConfigFormFields';
import type { ConfigLayer } from '@/components/settings/configFormUtils';
import { sanitizeConfigForLayer } from '@/components/settings/configFormUtils';
import { Button } from '@/components/ui/button';
import { messages } from '@/i18n/messages';
import {
	getConfigFieldErrors,
	validatePartialConfig,
} from '@/lib/configValidation';
import type { PartialConfig } from '@/types/config';

type ConfigEditorProps = {
	layer: ConfigLayer;
	settings: PartialConfig;
	onSave: (settings: PartialConfig) => Promise<boolean>;
	compact?: boolean;
};

/** ドラフト編集 + 入力時バリデーション + 保存ボタン */
export function ConfigEditor({
	layer,
	settings,
	onSave,
	compact,
}: ConfigEditorProps) {
	const [draft, setDraft] = useState(settings);
	const [saveErrors, setSaveErrors] = useState<string[]>([]);
	const [saving, setSaving] = useState(false);

	const validationDraft = useMemo(
		() => sanitizeConfigForLayer(draft, layer),
		[draft, layer],
	);
	const fieldErrors = useMemo(
		() => getConfigFieldErrors(validationDraft),
		[validationDraft],
	);
	const hasFieldErrors = Object.keys(fieldErrors).length > 0;

	useEffect(() => {
		setDraft(settings);
		setSaveErrors([]);
	}, [settings]);

	const handleSave = async () => {
		const validated = validatePartialConfig(validationDraft);
		if (validated.ok === false) {
			setSaveErrors(validated.errors);
			return;
		}
		setSaveErrors([]);
		setSaving(true);
		try {
			const ok = await onSave(sanitizeConfigForLayer(validated.data, layer));
			if (!ok) setSaveErrors([messages.settings.saveFailed]);
		} finally {
			setSaving(false);
		}
	};

	return (
		<div className={compact ? 'space-y-2' : 'space-y-3'}>
			<AppConfigTabs
				layer={layer}
				settings={draft}
				onChange={setDraft}
				fieldErrors={fieldErrors}
				compact={compact}
			/>
			{saveErrors.length > 0 && (
				<ul className='rounded border border-destructive/40 bg-destructive/10 px-2 py-1 text-[10px] text-destructive'>
					{saveErrors.map((e) => (
						<li key={e}>{e}</li>
					))}
				</ul>
			)}
			<Button
				type='button'
				size={compact ? 'xs' : 'sm'}
				className='w-full nodrag nopan nowheel'
				disabled={saving || hasFieldErrors}
				onClick={(e) => {
					e.stopPropagation();
					void handleSave();
				}}
			>
				{saving ? messages.settings.saving : messages.settings.save}
			</Button>
		</div>
	);
}
