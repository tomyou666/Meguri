import { type ReactNode, useEffect, useMemo, useState } from 'react';
import { ConfigField } from '@/components/settings/ConfigField';
import {
	type FieldErrors,
	fieldInvalid,
	formatOptionalNumber,
	inputClassName,
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

function configCheckboxId(path: string): string {
	return `cfg-${path.replace(/\./g, '-')}`;
}

function ConfigCheckboxRow({
	inputId,
	checked,
	onCheckedChange,
	children,
	className = 'flex items-start gap-2 text-xs',
}: {
	inputId: string;
	checked: boolean;
	onCheckedChange: (checked: boolean) => void;
	children: ReactNode;
	className?: string;
}) {
	return (
		<label htmlFor={inputId} className={className}>
			<Checkbox
				id={inputId}
				className='mt-0.5'
				checked={checked}
				onCheckedChange={onCheckedChange}
			/>
			{children}
		</label>
	);
}

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

export function ContentConfigFields({
	value,
	onChange,
	fieldErrors,
}: FieldsProps<PartialConfig['content']>) {
	const v = value ?? {};
	return (
		<div className='space-y-3'>
			<ConfigCheckboxRow
				inputId={configCheckboxId('content.only_main_content')}
				checked={v.only_main_content ?? false}
				onCheckedChange={(c) => onChange({ ...v, only_main_content: !!c })}
			>
				<FieldLabel label='only_main_content' help={h.only_main_content} />
			</ConfigCheckboxRow>
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
			<ConfigCheckboxRow
				inputId={configCheckboxId('content.extract_links')}
				checked={v.extract_links ?? true}
				onCheckedChange={(c) => onChange({ ...v, extract_links: !!c })}
			>
				<FieldLabel label='extract_links' help={h.extract_links} />
			</ConfigCheckboxRow>
			<ConfigCheckboxRow
				inputId={configCheckboxId('content.extract_metadata')}
				checked={v.extract_metadata ?? true}
				onCheckedChange={(c) => onChange({ ...v, extract_metadata: !!c })}
			>
				<FieldLabel label='extract_metadata' help={h.extract_metadata} />
			</ConfigCheckboxRow>
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
			<ConfigCheckboxRow
				inputId={configCheckboxId('pdf.enabled')}
				checked={v.enabled ?? true}
				onCheckedChange={(c) => onChange({ ...v, enabled: !!c })}
			>
				<FieldLabel label='enabled' help={h.pdf_enabled} />
			</ConfigCheckboxRow>
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
			<ConfigCheckboxRow
				inputId={configCheckboxId('crawl.enabled')}
				checked={v.enabled ?? false}
				onCheckedChange={(c) => onChange({ ...v, enabled: !!c })}
			>
				<FieldLabel label='enabled' help={h.crawl_enabled} />
			</ConfigCheckboxRow>
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
			<ConfigCheckboxRow
				inputId={configCheckboxId('crawl.allow_external_links')}
				checked={v.allow_external_links ?? false}
				onCheckedChange={(c) => onChange({ ...v, allow_external_links: !!c })}
			>
				<FieldLabel
					label='allow_external_links'
					help={h.allow_external_links}
				/>
			</ConfigCheckboxRow>
			<ConfigCheckboxRow
				inputId={configCheckboxId('crawl.allow_subdomains')}
				checked={v.allow_subdomains ?? false}
				onCheckedChange={(c) => onChange({ ...v, allow_subdomains: !!c })}
			>
				<FieldLabel label='allow_subdomains' help={h.allow_subdomains} />
			</ConfigCheckboxRow>
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
			<div className='space-y-3 rounded-lg border border-border/60 p-3'>
				<p className='text-xs font-medium text-muted-foreground'>
					取得並列 (fetch_limits)
				</p>
				<p className='text-xs text-muted-foreground'>
					{h.fetch_limits_overview}
				</p>
				<ConfigField
					path='crawl.fetch_limits.http_max_inflight'
					errors={fieldErrors}
					label='http_max_inflight'
					help={h.http_max_inflight}
				>
					<Input
						type='number'
						className={inputClassName(
							fieldInvalid(fieldErrors, 'crawl.fetch_limits.http_max_inflight'),
						)}
						value={formatOptionalNumber(v.fetch_limits?.http_max_inflight)}
						onChange={(e) =>
							onChange({
								...v,
								fetch_limits: {
									...v.fetch_limits,
									http_max_inflight: parseOptionalNumber(e.target.value),
								},
							})
						}
					/>
				</ConfigField>
				<ConfigField
					path='crawl.fetch_limits.chromium_max_inflight'
					errors={fieldErrors}
					label='chromium_max_inflight'
					help={h.chromium_max_inflight}
				>
					<Input
						type='number'
						className={inputClassName(
							fieldInvalid(
								fieldErrors,
								'crawl.fetch_limits.chromium_max_inflight',
							),
						)}
						value={formatOptionalNumber(v.fetch_limits?.chromium_max_inflight)}
						onChange={(e) =>
							onChange({
								...v,
								fetch_limits: {
									...v.fetch_limits,
									chromium_max_inflight: parseOptionalNumber(e.target.value),
								},
							})
						}
					/>
				</ConfigField>
				<ConfigCheckboxRow
					inputId={configCheckboxId('crawl.fetch_limits.auto_calibrate')}
					checked={v.fetch_limits?.auto_calibrate ?? true}
					onCheckedChange={(c) =>
						onChange({
							...v,
							fetch_limits: { ...v.fetch_limits, auto_calibrate: !!c },
						})
					}
				>
					<FieldLabel label='auto_calibrate' help={h.fetch_auto_calibrate} />
				</ConfigCheckboxRow>
				<ConfigCheckboxRow
					inputId={configCheckboxId('crawl.fetch_limits.dynamic_chromium')}
					checked={v.fetch_limits?.dynamic_chromium ?? true}
					onCheckedChange={(c) =>
						onChange({
							...v,
							fetch_limits: { ...v.fetch_limits, dynamic_chromium: !!c },
						})
					}
				>
					<FieldLabel
						label='dynamic_chromium'
						help={h.fetch_dynamic_chromium}
					/>
				</ConfigCheckboxRow>
			</div>
			<ConfigCheckboxRow
				inputId={configCheckboxId('crawl.respect_robots_txt')}
				checked={v.respect_robots_txt ?? true}
				onCheckedChange={(c) => onChange({ ...v, respect_robots_txt: !!c })}
			>
				<FieldLabel label='respect_robots_txt' help={h.respect_robots_txt} />
			</ConfigCheckboxRow>
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
				path='plugins.transformer'
				errors={fieldErrors}
				label='transformer'
				help={h.transformer}
			>
				<select
					className={selectClassName(
						fieldInvalid(fieldErrors, 'plugins.transformer'),
						'mt-1 h-8 w-full rounded-lg border border-input bg-background px-2',
					)}
					value={v.transformer ?? 'markdown'}
					onChange={(e) =>
						onChange({
							...v,
							transformer: e.target.value,
						})
					}
				>
					<option value='markdown'>markdown</option>
					<option value='html'>html</option>
					<option value='raw_html'>raw_html</option>
					<option value='json'>json</option>
				</select>
			</ConfigField>
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
			<ConfigCheckboxRow
				inputId={configCheckboxId('plugins.fetcher_config.headless')}
				checked={(fc.headless as boolean) ?? true}
				onCheckedChange={(c) =>
					onChange({
						...v,
						fetcher_config: { ...fc, headless: !!c },
					})
				}
			>
				<FieldLabel label='headless' help={h.headless} />
			</ConfigCheckboxRow>
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

type TabVisibilityOptions = {
	showPdfTab?: boolean;
	showRequestTab?: boolean;
	showCrawlTab?: boolean;
};

function tabsForLayer(
	layer: ConfigLayer,
	showMeta: boolean,
	options?: TabVisibilityOptions,
): string[] {
	const tabs = [
		...(showMeta ? ['general'] : []),
		'request',
		'content',
		'pdf',
		'crawl',
		...(layer === 'app' || layer === 'workspace' ? ['plugins'] : []),
		...(layer === 'app' ? ['output'] : []),
	];
	return tabs.filter((key) => {
		if (key === 'pdf' && options?.showPdfTab === false) return false;
		if (key === 'request' && options?.showRequestTab === false) return false;
		if (key === 'crawl' && options?.showCrawlTab === false) return false;
		return true;
	});
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
	showPdfTab = true,
	showRequestTab = true,
	showCrawlTab = true,
}: {
	settings: PartialConfig;
	onChange: (s: PartialConfig) => void;
	fieldErrors?: FieldErrors;
	showMeta?: boolean;
	meta?: { name: string; seedUrl: string };
	onMetaChange?: (m: { name: string; seedUrl: string }) => void;
	layer?: ConfigLayer;
	compact?: boolean;
	showPdfTab?: boolean;
	showRequestTab?: boolean;
	showCrawlTab?: boolean;
}) {
	const [tab, setTab] = useState('request');
	const tabs = useMemo(
		() =>
			tabsForLayer(layer, !!showMeta, {
				showPdfTab,
				showRequestTab,
				showCrawlTab,
			}),
		[layer, showMeta, showPdfTab, showRequestTab, showCrawlTab],
	);
	const tabBtn = compact ? 'px-1.5 py-0.5 text-[9px]' : 'px-2 py-1 text-xs';

	useEffect(() => {
		if (!tabs.includes(tab)) {
			setTab(tabs.includes('content') ? 'content' : (tabs[0] ?? 'request'));
		}
	}, [tab, tabs]);

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
					onChange={(content) => {
						if (!content) {
							onChange({ ...settings, content: undefined });
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
