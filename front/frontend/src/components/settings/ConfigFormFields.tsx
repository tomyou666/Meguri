import { useState } from 'react';
import { ConfigField } from '@/components/settings/ConfigField';
import {
	type FieldErrors,
	fieldInvalid,
	formatOptionalNumber,
	inputClassName,
	layerShowsContentFormats,
	parseOptionalNumber,
	selectClassName,
	textareaClassName,
} from '@/components/settings/configFormUtils';
import { FieldLabel } from '@/components/settings/FieldLabel';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { messages } from '@/i18n/messages';
import type { PartialConfig } from '@/types/config';

export type { ConfigLayer } from '@/components/settings/configFormUtils';

import type { ConfigLayer } from '@/components/settings/configFormUtils';

const h = messages.settings.help;
const t = messages.settings.tabs;

type FieldsProps<T> = {
	value: T;
	onChange: (v: T) => void;
	fieldErrors: FieldErrors;
};

function StringListEditor({
	path,
	label,
	help,
	values,
	onChange,
	fieldErrors,
}: {
	path: string;
	label: string;
	help: string;
	values: string[];
	onChange: (v: string[]) => void;
	fieldErrors: FieldErrors;
}) {
	const text = values.join('\n');
	const invalid = fieldInvalid(fieldErrors, path);
	return (
		<ConfigField path={path} errors={fieldErrors} label={label} help={help}>
			<textarea
				className={textareaClassName(invalid)}
				value={text}
				onChange={(e) =>
					onChange(
						e.target.value
							.split('\n')
							.map((s) => s.trim())
							.filter(Boolean),
					)
				}
			/>
		</ConfigField>
	);
}

export function RequestConfigFields({
	value,
	onChange,
	fieldErrors,
}: FieldsProps<PartialConfig['request']>) {
	const v = value ?? {};
	return (
		<div className='space-y-3'>
			<ConfigField
				path='request.timeout'
				errors={fieldErrors}
				label='timeout'
				help={h.timeout}
			>
				<Input
					className={inputClassName(
						fieldInvalid(fieldErrors, 'request.timeout'),
					)}
					value={v.timeout ?? ''}
					onChange={(e) => onChange({ ...v, timeout: e.target.value })}
				/>
			</ConfigField>
			<ConfigField
				path='request.retry_count'
				errors={fieldErrors}
				label='retry_count'
				help={h.retry_count}
			>
				<Input
					type='number'
					className={inputClassName(
						fieldInvalid(fieldErrors, 'request.retry_count'),
					)}
					value={formatOptionalNumber(v.retry_count)}
					onChange={(e) =>
						onChange({ ...v, retry_count: parseOptionalNumber(e.target.value) })
					}
				/>
			</ConfigField>
			<ConfigField
				path='request.retry_interval'
				errors={fieldErrors}
				label='retry_interval'
				help={h.retry_interval}
			>
				<Input
					className={inputClassName(
						fieldInvalid(fieldErrors, 'request.retry_interval'),
					)}
					value={v.retry_interval ?? ''}
					onChange={(e) => onChange({ ...v, retry_interval: e.target.value })}
				/>
			</ConfigField>
			<ConfigField
				path='request.headers'
				errors={fieldErrors}
				label='User-Agent'
				help={h.userAgent}
			>
				<Input
					className={inputClassName(
						fieldInvalid(fieldErrors, 'request.headers'),
					)}
					value={v.headers?.['User-Agent'] ?? ''}
					onChange={(e) =>
						onChange({
							...v,
							headers: { ...v.headers, 'User-Agent': e.target.value },
						})
					}
				/>
			</ConfigField>
		</div>
	);
}

const FORMATS = [
	'markdown',
	'html',
	'raw_html',
	'json',
	'links',
	'metadata',
] as const;

export function ContentConfigFields({
	value,
	onChange,
	fieldErrors,
	showFormats = true,
}: FieldsProps<PartialConfig['content']> & { showFormats?: boolean }) {
	const v = value ?? {};
	const formatsInvalid = fieldInvalid(fieldErrors, 'content.formats');
	return (
		<div className='space-y-3'>
			{showFormats && (
				<ConfigField
					path='content.formats'
					errors={fieldErrors}
					label='formats'
					help={h.formats}
				>
					<div
						className={`mt-1 flex flex-wrap gap-2 rounded-lg p-1 ${formatsInvalid ? 'border border-destructive bg-destructive/5' : ''}`}
					>
						{FORMATS.map((f) => (
							<label key={f} className='flex items-center gap-1 text-xs'>
								<Checkbox
									checked={v.formats?.includes(f) ?? false}
									onCheckedChange={(checked) => {
										const cur = v.formats ?? [];
										const next = checked
											? [...cur, f]
											: cur.filter((x) => x !== f);
										onChange({ ...v, formats: next });
									}}
								/>
								{f}
							</label>
						))}
					</div>
				</ConfigField>
			)}
			<label className='flex items-start gap-2 text-xs'>
				<Checkbox
					className='mt-0.5'
					checked={v.only_main_content ?? false}
					onCheckedChange={(c) => onChange({ ...v, only_main_content: !!c })}
				/>
				<FieldLabel label='only_main_content' help={h.only_main_content} />
			</label>
			<StringListEditor
				path='content.include_tags'
				label='include_tags'
				help={h.include_tags}
				values={v.include_tags ?? []}
				onChange={(include_tags) => onChange({ ...v, include_tags })}
				fieldErrors={fieldErrors}
			/>
			<StringListEditor
				path='content.exclude_tags'
				label='exclude_tags'
				help={h.exclude_tags}
				values={v.exclude_tags ?? []}
				onChange={(exclude_tags) => onChange({ ...v, exclude_tags })}
				fieldErrors={fieldErrors}
			/>
			<ConfigField
				path='content.selector'
				errors={fieldErrors}
				label='selector'
				help={h.selector}
			>
				<Input
					className={inputClassName(
						fieldInvalid(fieldErrors, 'content.selector'),
					)}
					value={v.selector ?? ''}
					onChange={(e) => onChange({ ...v, selector: e.target.value })}
				/>
			</ConfigField>
			<label className='flex items-start gap-2 text-xs'>
				<Checkbox
					className='mt-0.5'
					checked={v.extract_links ?? true}
					onCheckedChange={(c) => onChange({ ...v, extract_links: !!c })}
				/>
				<FieldLabel label='extract_links' help={h.extract_links} />
			</label>
			<label className='flex items-start gap-2 text-xs'>
				<Checkbox
					className='mt-0.5'
					checked={v.extract_metadata ?? true}
					onCheckedChange={(c) => onChange({ ...v, extract_metadata: !!c })}
				/>
				<FieldLabel label='extract_metadata' help={h.extract_metadata} />
			</label>
		</div>
	);
}

export function PdfConfigFields({
	value,
	onChange,
	fieldErrors,
}: FieldsProps<PartialConfig['pdf']>) {
	const v = value ?? {};
	return (
		<div className='space-y-3'>
			<label className='flex items-start gap-2 text-xs'>
				<Checkbox
					className='mt-0.5'
					checked={v.enabled ?? true}
					onCheckedChange={(c) => onChange({ ...v, enabled: !!c })}
				/>
				<FieldLabel label='enabled' help={h.pdf_enabled} />
			</label>
			<ConfigField
				path='pdf.mode'
				errors={fieldErrors}
				label='mode'
				help={h.pdf_mode}
			>
				<select
					className={selectClassName(fieldInvalid(fieldErrors, 'pdf.mode'))}
					value={v.mode ?? 'auto'}
					onChange={(e) =>
						onChange({
							...v,
							mode: e.target.value as 'fast' | 'auto' | 'ocr',
						})
					}
				>
					<option value='fast'>fast</option>
					<option value='auto'>auto</option>
					<option value='ocr'>ocr</option>
				</select>
			</ConfigField>
			<ConfigField
				path='pdf.max_pages'
				errors={fieldErrors}
				label='max_pages'
				help={h.pdf_max_pages}
			>
				<Input
					type='number'
					className={inputClassName(fieldInvalid(fieldErrors, 'pdf.max_pages'))}
					value={formatOptionalNumber(v.max_pages)}
					onChange={(e) =>
						onChange({ ...v, max_pages: parseOptionalNumber(e.target.value) })
					}
				/>
			</ConfigField>
			<ConfigField
				path='pdf.output'
				errors={fieldErrors}
				label='output'
				help={h.pdf_output}
			>
				<select
					className={selectClassName(fieldInvalid(fieldErrors, 'pdf.output'))}
					value={v.output ?? 'text'}
					onChange={(e) =>
						onChange({
							...v,
							output: e.target.value as 'text' | 'markdown' | 'raw',
						})
					}
				>
					<option value='text'>text</option>
					<option value='markdown'>markdown</option>
					<option value='raw'>raw</option>
				</select>
			</ConfigField>
		</div>
	);
}

export function CrawlConfigFields({
	value,
	onChange,
	fieldErrors,
}: FieldsProps<PartialConfig['crawl']>) {
	const v = value ?? {};
	return (
		<div className='space-y-3'>
			<label className='flex items-start gap-2 text-xs'>
				<Checkbox
					className='mt-0.5'
					checked={v.enabled ?? false}
					onCheckedChange={(c) => onChange({ ...v, enabled: !!c })}
				/>
				<FieldLabel label='enabled' help={h.crawl_enabled} />
			</label>
			<ConfigField
				path='crawl.max_depth'
				errors={fieldErrors}
				label='max_depth'
				help={h.max_depth}
			>
				<Input
					type='number'
					className={inputClassName(
						fieldInvalid(fieldErrors, 'crawl.max_depth'),
					)}
					value={formatOptionalNumber(v.max_depth)}
					onChange={(e) =>
						onChange({ ...v, max_depth: parseOptionalNumber(e.target.value) })
					}
				/>
			</ConfigField>
			<ConfigField
				path='crawl.max_pages'
				errors={fieldErrors}
				label='max_pages'
				help={h.max_pages}
			>
				<Input
					type='number'
					className={inputClassName(
						fieldInvalid(fieldErrors, 'crawl.max_pages'),
					)}
					value={formatOptionalNumber(v.max_pages)}
					onChange={(e) =>
						onChange({ ...v, max_pages: parseOptionalNumber(e.target.value) })
					}
				/>
			</ConfigField>
			<StringListEditor
				path='crawl.include_paths'
				label='include_paths'
				help={h.include_paths}
				values={v.include_paths ?? []}
				onChange={(include_paths) => onChange({ ...v, include_paths })}
				fieldErrors={fieldErrors}
			/>
			<StringListEditor
				path='crawl.exclude_paths'
				label='exclude_paths'
				help={h.exclude_paths}
				values={v.exclude_paths ?? []}
				onChange={(exclude_paths) => onChange({ ...v, exclude_paths })}
				fieldErrors={fieldErrors}
			/>
			<label className='flex items-start gap-2 text-xs'>
				<Checkbox
					className='mt-0.5'
					checked={v.allow_external_links ?? false}
					onCheckedChange={(c) => onChange({ ...v, allow_external_links: !!c })}
				/>
				<FieldLabel
					label='allow_external_links'
					help={h.allow_external_links}
				/>
			</label>
			<label className='flex items-start gap-2 text-xs'>
				<Checkbox
					className='mt-0.5'
					checked={v.allow_subdomains ?? false}
					onCheckedChange={(c) => onChange({ ...v, allow_subdomains: !!c })}
				/>
				<FieldLabel label='allow_subdomains' help={h.allow_subdomains} />
			</label>
			<ConfigField
				path='crawl.request_delay'
				errors={fieldErrors}
				label='request_delay'
				help={h.request_delay}
			>
				<Input
					className={inputClassName(
						fieldInvalid(fieldErrors, 'crawl.request_delay'),
					)}
					value={v.request_delay ?? ''}
					onChange={(e) => onChange({ ...v, request_delay: e.target.value })}
				/>
			</ConfigField>
			<ConfigField
				path='crawl.max_concurrency'
				errors={fieldErrors}
				label='max_concurrency'
				help={h.max_concurrency}
			>
				<Input
					type='number'
					className={inputClassName(
						fieldInvalid(fieldErrors, 'crawl.max_concurrency'),
					)}
					value={formatOptionalNumber(v.max_concurrency)}
					onChange={(e) =>
						onChange({
							...v,
							max_concurrency: parseOptionalNumber(e.target.value),
						})
					}
				/>
			</ConfigField>
			<label className='flex items-start gap-2 text-xs'>
				<Checkbox
					className='mt-0.5'
					checked={v.respect_robots_txt ?? true}
					onCheckedChange={(c) => onChange({ ...v, respect_robots_txt: !!c })}
				/>
				<FieldLabel label='respect_robots_txt' help={h.respect_robots_txt} />
			</label>
		</div>
	);
}

export function PluginsConfigFields({
	value,
	onChange,
	fieldErrors,
}: FieldsProps<PartialConfig['plugins']>) {
	const v = value ?? {};
	const fc = v.fetcher_config ?? {};
	return (
		<div className='space-y-3 text-xs'>
			<ConfigField
				path='plugins.fetcher'
				errors={fieldErrors}
				label='fetcher'
				help={h.fetcher}
			>
				<select
					className={selectClassName(
						fieldInvalid(fieldErrors, 'plugins.fetcher'),
						'mt-1 h-8 w-full rounded-lg border border-input bg-background px-2',
					)}
					value={v.fetcher ?? 'http'}
					onChange={(e) =>
						onChange({
							...v,
							fetcher: e.target.value as 'http' | 'chromium',
						})
					}
				>
					<option value='http'>http</option>
					<option value='chromium'>chromium</option>
				</select>
			</ConfigField>
			<ConfigField
				path='plugins.fetcher_config.browser_path'
				errors={fieldErrors}
				label='browser_path'
				help={h.browser_path}
			>
				<Input
					className={inputClassName(
						fieldInvalid(fieldErrors, 'plugins.fetcher_config.browser_path'),
						'mt-1 h-8',
					)}
					value={(fc.browser_path as string) ?? ''}
					onChange={(e) =>
						onChange({
							...v,
							fetcher_config: { ...fc, browser_path: e.target.value },
						})
					}
				/>
			</ConfigField>
			<label className='flex items-start gap-2'>
				<Checkbox
					className='mt-0.5'
					checked={(fc.headless as boolean) ?? true}
					onCheckedChange={(c) =>
						onChange({
							...v,
							fetcher_config: { ...fc, headless: !!c },
						})
					}
				/>
				<FieldLabel label='headless' help={h.headless} />
			</label>
		</div>
	);
}

export function OutputConfigFields({
	value,
	onChange,
	fieldErrors,
}: FieldsProps<PartialConfig['output']>) {
	const v = value ?? {};
	return (
		<div className='space-y-3'>
			<ConfigField
				path='output.dir'
				errors={fieldErrors}
				label='dir'
				help={h.output_dir}
			>
				<Input
					className={inputClassName(fieldInvalid(fieldErrors, 'output.dir'))}
					value={v.dir ?? ''}
					onChange={(e) => onChange({ ...v, dir: e.target.value })}
				/>
			</ConfigField>
			<ConfigField
				path='output.file_pattern'
				errors={fieldErrors}
				label='file_pattern'
				help={h.file_pattern}
			>
				<Input
					className={inputClassName(
						fieldInvalid(fieldErrors, 'output.file_pattern'),
					)}
					value={v.file_pattern ?? ''}
					onChange={(e) => onChange({ ...v, file_pattern: e.target.value })}
				/>
			</ConfigField>
		</div>
	);
}

const TAB_LABELS: Record<string, string> = {
	general: t.general,
	request: t.request,
	content: t.content,
	pdf: t.pdf,
	crawl: t.crawl,
	plugins: t.plugins,
	output: t.output,
};

function tabsForLayer(layer: ConfigLayer, showMeta: boolean): string[] {
	return [
		...(showMeta ? ['general'] : []),
		'request',
		'content',
		'pdf',
		'crawl',
		...(layer === 'app' || layer === 'workspace' ? ['plugins'] : []),
		...(layer === 'app' ? ['output'] : []),
	];
}

export function TabsConfig({
	settings,
	onChange,
	fieldErrors = {},
	showMeta,
	meta,
	onMetaChange,
	layer = 'app',
	compact,
}: {
	settings: PartialConfig;
	onChange: (s: PartialConfig) => void;
	fieldErrors?: FieldErrors;
	showMeta?: boolean;
	meta?: { name: string; seedUrl: string };
	onMetaChange?: (m: { name: string; seedUrl: string }) => void;
	layer?: ConfigLayer;
	compact?: boolean;
}) {
	const [tab, setTab] = useState('request');
	const tabs = tabsForLayer(layer, !!showMeta);
	const tabBtn = compact ? 'px-1.5 py-0.5 text-[9px]' : 'px-2 py-1 text-xs';

	return (
		<div className='flex flex-col gap-2'>
			<div className='flex flex-wrap gap-1'>
				{tabs.map((key) => (
					<button
						key={key}
						type='button'
						className={`rounded ${tabBtn} ${tab === key ? 'bg-primary text-primary-foreground' : 'bg-muted'}`}
						onClick={() => setTab(key)}
					>
						{TAB_LABELS[key] ?? key}
					</button>
				))}
			</div>
			{tab === 'general' && showMeta && meta && onMetaChange && (
				<div className='space-y-2'>
					<div>
						<FieldLabel label='name' help={h.ws_name} />
						<Input
							className='mt-1 h-8 text-xs'
							value={meta.name}
							onChange={(e) => onMetaChange({ ...meta, name: e.target.value })}
						/>
					</div>
					<div>
						<FieldLabel label='seed_url' help={h.seed_url} />
						<Input
							className='mt-1 h-8 text-xs'
							value={meta.seedUrl}
							onChange={(e) =>
								onMetaChange({ ...meta, seedUrl: e.target.value })
							}
						/>
					</div>
				</div>
			)}
			{tab === 'request' && (
				<RequestConfigFields
					value={settings.request}
					onChange={(request) => onChange({ ...settings, request })}
					fieldErrors={fieldErrors}
				/>
			)}
			{tab === 'content' && (
				<ContentConfigFields
					value={settings.content}
					showFormats={layerShowsContentFormats(layer)}
					onChange={(content) => {
						if (!content) {
							onChange({ ...settings, content: undefined });
							return;
						}
						if (!layerShowsContentFormats(layer)) {
							const { formats: _f, ...rest } = content;
							onChange({
								...settings,
								content: Object.keys(rest).length > 0 ? rest : undefined,
							});
							return;
						}
						onChange({ ...settings, content });
					}}
					fieldErrors={fieldErrors}
				/>
			)}
			{tab === 'pdf' && (
				<PdfConfigFields
					value={settings.pdf}
					onChange={(pdf) => onChange({ ...settings, pdf })}
					fieldErrors={fieldErrors}
				/>
			)}
			{tab === 'crawl' && (
				<CrawlConfigFields
					value={settings.crawl}
					onChange={(crawl) => onChange({ ...settings, crawl })}
					fieldErrors={fieldErrors}
				/>
			)}
			{tab === 'plugins' && (
				<PluginsConfigFields
					value={settings.plugins}
					onChange={(plugins) => onChange({ ...settings, plugins })}
					fieldErrors={fieldErrors}
				/>
			)}
			{tab === 'output' && (
				<OutputConfigFields
					value={settings.output}
					onChange={(output) => onChange({ ...settings, output })}
					fieldErrors={fieldErrors}
				/>
			)}
		</div>
	);
}

export function WorkspaceConfigTabs(props: {
	settings: PartialConfig;
	onChange: (s: PartialConfig) => void;
	fieldErrors?: FieldErrors;
	showMeta?: boolean;
	meta?: { name: string; seedUrl: string };
	onMetaChange?: (m: { name: string; seedUrl: string }) => void;
}) {
	return <TabsConfig {...props} layer='workspace' />;
}

export { TabsConfig as AppConfigTabs };
