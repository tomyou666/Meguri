import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ScrollArea } from '@/components/ui/scroll-area';
import { messages } from '@/i18n/messages';
import {
	DEFAULT_EXPORT_SEPARATOR,
	type ExportHeadingField,
	type ExportMergeSettings,
} from '@/lib/exportTree';

type ExportSettingsSidebarProps = {
	settings: ExportMergeSettings;
	onSettingsChange: (settings: ExportMergeSettings) => void;
	checkedCount: number;
	previewLoading: boolean;
	saving: boolean;
	onPreviewStart: () => void;
	onSave: () => void;
};

export function ExportSettingsSidebar({
	settings,
	onSettingsChange,
	checkedCount,
	previewLoading,
	saving,
	onPreviewStart,
	onSave,
}: ExportSettingsSidebarProps) {
	const actionsDisabled = checkedCount === 0 || saving;

	return (
		<aside className='flex h-full w-full min-w-[14rem] flex-col border-border border-l bg-card'>
			<div className='border-border border-b px-3 py-2 font-semibold text-xs'>
				{messages.export.settingsTitle}
			</div>
			<ScrollArea className='min-h-0 flex-1'>
				<div className='flex flex-col gap-4 p-3'>
					<div className='flex flex-col gap-1.5'>
						<Label className='text-xs'>{messages.export.format}</Label>
						<div className='flex gap-1'>
							<Button
								size='xs'
								variant={settings.format === 'markdown' ? 'default' : 'outline'}
								onClick={() =>
									onSettingsChange({ ...settings, format: 'markdown' })
								}
							>
								{messages.export.formatMarkdown}
							</Button>
							<Button
								size='xs'
								variant={settings.format === 'html' ? 'default' : 'outline'}
								onClick={() =>
									onSettingsChange({ ...settings, format: 'html' })
								}
							>
								{messages.export.formatHtml}
							</Button>
						</div>
					</div>

					<div className='flex items-center gap-2'>
						<Checkbox
							id='export-split-save'
							checked={settings.splitSave}
							onCheckedChange={(checked) =>
								onSettingsChange({
									...settings,
									splitSave: checked === true,
								})
							}
						/>
						<Label htmlFor='export-split-save' className='font-normal text-xs'>
							{messages.export.splitSave}
						</Label>
					</div>
					<p className='text-[10px] text-muted-foreground'>
						{messages.export.splitSaveHint}
					</p>

					<div className='flex flex-col gap-1.5'>
						<Label className='text-xs'>{messages.export.separator}</Label>
						<Input
							value={settings.separator}
							onChange={(e) =>
								onSettingsChange({ ...settings, separator: e.target.value })
							}
							placeholder={DEFAULT_EXPORT_SEPARATOR}
							className='h-8 font-mono text-xs'
						/>
						<p className='text-[10px] text-muted-foreground'>
							{messages.export.separatorHint}
						</p>
					</div>

					<div className='flex items-center gap-2'>
						<Checkbox
							id='export-include-heading'
							checked={settings.includeHeading}
							onCheckedChange={(checked) =>
								onSettingsChange({
									...settings,
									includeHeading: checked === true,
								})
							}
						/>
						<Label
							htmlFor='export-include-heading'
							className='font-normal text-xs'
						>
							{messages.export.includeHeading}
						</Label>
					</div>

					{settings.includeHeading && (
						<div className='flex flex-col gap-1.5'>
							<Label className='text-xs'>{messages.export.headingField}</Label>
							<div className='flex gap-1'>
								{(
									[
										['url', messages.export.headingUrl],
										['label', messages.export.headingLabel],
									] as const
								).map(([value, label]) => (
									<Button
										key={value}
										size='xs'
										variant={
											settings.headingField === value ? 'default' : 'outline'
										}
										onClick={() =>
											onSettingsChange({
												...settings,
												headingField: value as ExportHeadingField,
											})
										}
									>
										{label}
									</Button>
								))}
							</div>
						</div>
					)}
				</div>
			</ScrollArea>

			<div className='flex flex-col gap-2 border-border border-t p-3'>
				{checkedCount === 0 && (
					<p className='text-[10px] text-muted-foreground'>
						{messages.export.noNodesChecked}
					</p>
				)}
				<Button
					size='sm'
					className='w-full'
					disabled={actionsDisabled || previewLoading}
					onClick={onPreviewStart}
				>
					{messages.export.previewStart}
				</Button>
				<Button
					size='sm'
					variant='outline'
					className='w-full'
					disabled={actionsDisabled || previewLoading}
					onClick={onSave}
				>
					{messages.export.save}
				</Button>
			</div>
		</aside>
	);
}
