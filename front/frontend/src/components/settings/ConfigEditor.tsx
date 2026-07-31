import { useEffect, useMemo, useState } from 'react';
import { AppConfigTabs } from '@/components/settings/ConfigFormFields';
import type { ConfigLayer } from '@/components/settings/configFormUtils';
import {
	defaultsForLayer,
	sanitizeConfigForLayer,
} from '@/components/settings/configFormUtils';
import { Button } from '@/components/ui/button';
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import { messages } from '@/i18n/messages';
import {
	getConfigFieldErrors,
	validatePartialConfig,
} from '@/lib/configValidation';
import { DEFAULT_APP_CONFIG } from '@/lib/defaults';
import type { PartialConfig } from '@/types/config';

type ConfigEditorProps = {
	layer: ConfigLayer;
	settings: PartialConfig;
	onSave: (settings: PartialConfig) => Promise<boolean>;
	/** ノード設定削除。渡したときのみ削除ボタンを表示する。 */
	onDelete?: () => Promise<boolean>;
	/** リセット先。省略時は DEFAULT_APP_CONFIG。 */
	defaults?: PartialConfig;
	compact?: boolean;
	/** 省略時 true。false のとき PDF 設定タブを非表示にする。 */
	showPdfTab?: boolean;
	/** 省略時 true。false のとき HTTP 設定タブを非表示にする。 */
	showRequestTab?: boolean;
	/** 省略時 true。false のときクロール設定タブを非表示にする。 */
	showCrawlTab?: boolean;
};

/** ドラフト編集 + 入力時バリデーション + 保存ボタン */
export function ConfigEditor({
	layer,
	settings,
	onSave,
	onDelete,
	defaults,
	compact,
	showPdfTab = true,
	showRequestTab = true,
	showCrawlTab = true,
}: ConfigEditorProps) {
	const [draft, setDraft] = useState(settings);
	const [saveErrors, setSaveErrors] = useState<string[]>([]);
	const [saving, setSaving] = useState(false);
	const [deleting, setDeleting] = useState(false);
	const [confirmDelete, setConfirmDelete] = useState(false);

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

	const handleReset = () => {
		setDraft(defaultsForLayer(defaults ?? DEFAULT_APP_CONFIG, layer));
		setSaveErrors([]);
	};

	const handleDelete = async () => {
		if (!onDelete) return;
		setDeleting(true);
		try {
			const ok = await onDelete();
			if (ok) {
				setConfirmDelete(false);
			} else {
				setSaveErrors([messages.settings.deleteFailed]);
			}
		} finally {
			setDeleting(false);
		}
	};

	const busy = saving || deleting;

	return (
		<div className='flex h-full min-h-0 flex-col'>
			<div
				className={`min-h-0 flex-1 overflow-y-auto ${compact ? 'space-y-2 pr-1' : 'space-y-3 pr-2'}`}
			>
				<AppConfigTabs
					layer={layer}
					settings={draft}
					onChange={setDraft}
					fieldErrors={fieldErrors}
					compact={compact}
					showPdfTab={showPdfTab}
					showRequestTab={showRequestTab}
					showCrawlTab={showCrawlTab}
				/>
			</div>
			<div
				className={`shrink-0 border-border border-t ${compact ? 'space-y-2' : 'space-y-3'}`}
			>
				{saveErrors.length > 0 && (
					<ul className='rounded border border-destructive/40 bg-destructive/10 px-2 py-1 text-[10px] text-destructive'>
						{saveErrors.map((e) => (
							<li key={e}>{e}</li>
						))}
					</ul>
				)}
				<div className='flex gap-2'>
					<Button
						type='button'
						variant='outline'
						size={compact ? 'xs' : 'sm'}
						className='nodrag nopan nowheel flex-1'
						disabled={busy}
						onClick={(e) => {
							e.stopPropagation();
							handleReset();
						}}
					>
						{messages.settings.reset}
					</Button>
					<Button
						type='button'
						size={compact ? 'xs' : 'sm'}
						className='nodrag nopan nowheel flex-1'
						disabled={busy || hasFieldErrors}
						onClick={(e) => {
							e.stopPropagation();
							void handleSave();
						}}
					>
						{saving ? messages.settings.saving : messages.settings.save}
					</Button>
				</div>
				{onDelete && (
					<Button
						type='button'
						variant='destructive'
						size={compact ? 'xs' : 'sm'}
						className='nodrag nopan nowheel w-full'
						disabled={busy}
						onClick={(e) => {
							e.stopPropagation();
							setConfirmDelete(true);
						}}
					>
						{deleting ? messages.settings.deleting : messages.settings.delete}
					</Button>
				)}
			</div>
			<Dialog open={confirmDelete} onOpenChange={setConfirmDelete}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{messages.dialog.deleteNodeSettingsTitle}</DialogTitle>
					</DialogHeader>
					<p className='text-sm'>{messages.dialog.deleteNodeSettingsConfirm}</p>
					<DialogFooter>
						<Button
							variant='outline'
							size='sm'
							disabled={deleting}
							onClick={() => setConfirmDelete(false)}
						>
							{messages.dialog.cancel}
						</Button>
						<Button
							variant='destructive'
							size='sm'
							disabled={deleting}
							onClick={() => void handleDelete()}
						>
							{deleting ? messages.settings.deleting : messages.dialog.delete}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
