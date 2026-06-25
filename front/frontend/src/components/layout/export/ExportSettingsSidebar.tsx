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
	hasPreview: boolean;
	previewLoading: boolean;
	onPreviewStart: () => void;
	onSave: () => void;
	onCopy: () => void;
};

export function ExportSettingsSidebar({
	settings,
	onSettingsChange,
	checkedCount,
	hasPreview,
	previewLoading,
	onPreviewStart,
	onSave,
	onCopy,
}: ExportSettingsSidebarProps) {
	const actionsDisabled = checkedCount === 0;

	return (
		<aside className='flex h-full w-full min-w-[14rem] flex-col border-l border-border bg-card'>
			<div className='border-b border-border px-3 py-2 text-xs font-semibold'>
				{messages.export.settingsTitle}
			</div>
			<ScrollArea className='min-h-0 flex-1'>
				<div className='space-y-4 p-3'>
					<div className='space-y-1.5'>
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

					<div className='space-y-1.5'>
						<Label htmlFor='export-separator' className='text-xs'>
							{messages.export.separator}
						</Label>
						<Input
							id='export-separator'
							className='h-8 font-mono text-xs'
							value={settings.separator}
							onChange={(e) =>
								onSettingsChange({ ...settings, separator: e.target.value })
							}
						/>
						<p className='text-[10px] text-muted-foreground'>
							{messages.export.separatorHint}
						</p>
						<Button
							size='xs'
							variant='ghost'
							className='h-6 px-1 text-[10px]'
							onClick={() =>
								onSettingsChange({
									...settings,
									separator: DEFAULT_EXPORT_SEPARATOR,
								})
							}
						>
							---
						</Button>
					</div>

					<div className='flex items-center gap-2'>
						<Checkbox
							id='export-heading'
							checked={settings.includeHeading}
							onCheckedChange={(checked) =>
								onSettingsChange({
									...settings,
									includeHeading: checked === true,
								})
							}
						/>
						<Label htmlFor='export-heading' className='text-xs font-normal'>
							{messages.export.includeHeading}
						</Label>
					</div>

					{settings.includeHeading && (
						<div className='space-y-1.5'>
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

			<div className='space-y-2 border-t border-border p-3'>
				{actionsDisabled && (
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
					disabled={actionsDisabled || !hasPreview || previewLoading}
					onClick={onSave}
				>
					{messages.export.save}
				</Button>
				<Button
					size='sm'
					variant='outline'
					className='w-full'
					disabled={actionsDisabled || !hasPreview || previewLoading}
					onClick={onCopy}
				>
					{messages.export.copy}
				</Button>
			</div>
		</aside>
	);
}
