import { type ReactNode, useEffect, useMemo, useState } from 'react';
import { ConfigField } from '@/components/settings/ConfigField';
import {
	type FieldErrors,
	fieldInvalid,
	inputClassName,
	selectClassName,
} from '@/components/settings/configFormUtils';
import { DurationInput } from '@/components/settings/DurationInput';
import { FieldLabel } from '@/components/settings/FieldLabel';
import { LocalePresetSelect } from '@/components/settings/LocalePresetSelect';
import { OptionalNumberInput } from '@/components/settings/OptionalNumberInput';
import { TagListInput } from '@/components/settings/TagListInput';
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
	compact,
}: {
	path: string;
	label: string;
	help: string;
	values: string[];
	onChange: (v: string[]) => void;
	fieldErrors: FieldErrors;
	compact?: boolean;
}) {
	const invalid = fieldInvalid(fieldErrors, path);
	return (
		<ConfigField path={path} errors={fieldErrors} label={label} help={help}>
			<TagListInput
				values={values}
				onChange={onChange}
				invalid={invalid}
				compact={compact}
				placeholder={messages.settings.tagList.placeholder}
				removeLabel={messages.settings.tagList.remove}
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
				<DurationInput
					invalid={fieldInvalid(fieldErrors, 'request.timeout')}
					value={v.timeout}
					onChange={(timeout) => onChange({ ...v, timeout })}
				/>
			</ConfigField>
			<ConfigField
				path='request.retry_count'
				errors={fieldErrors}
				label='retry_count'
				help={h.retry_count}
			>
				<OptionalNumberInput
					className={inputClassName(
						fieldInvalid(fieldErrors, 'request.retry_count'),
					)}
					value={v.retry_count}
					onChange={(retry_count) => onChange({ ...v, retry_count })}
				/>
			</ConfigField>
			<ConfigField
				path='request.retry_interval'
				errors={fieldErrors}
				label='retry_interval'
				help={h.retry_interval}
			>
				<DurationInput
					invalid={fieldInvalid(fieldErrors, 'request.retry_interval')}
					value={v.retry_interval}
					onChange={(retry_interval) => onChange({ ...v, retry_interval })}
				/>
			</ConfigField>
		</div>
	);
}

export function ContentConfigFields({
	value,
	onChange,
	fieldErrors,
	compact,
}: FieldsProps<PartialConfig['content']> & { compact?: boolean }) {
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
				compact={compact}
			/>
			<StringListEditor
				path='content.exclude_tags'
				label='exclude_tags'
				help={h.exclude_tags}
				values={v.exclude_tags ?? []}
				onChange={(exclude_tags) => onChange({ ...v, exclude_tags })}
				fieldErrors={fieldErrors}
				compact={compact}
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
				<OptionalNumberInput
					className={inputClassName(fieldInvalid(fieldErrors, 'pdf.max_pages'))}
					value={v.max_pages}
					onChange={(max_pages) => onChange({ ...v, max_pages })}
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
	compact,
}: FieldsProps<PartialConfig['crawl']> & { compact?: boolean }) {
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
				<OptionalNumberInput
					className={inputClassName(
						fieldInvalid(fieldErrors, 'crawl.max_depth'),
					)}
					value={v.max_depth}
					onChange={(max_depth) => onChange({ ...v, max_depth })}
				/>
			</ConfigField>
			<ConfigField
				path='crawl.max_pages'
				errors={fieldErrors}
				label='max_pages'
				help={h.max_pages}
			>
				<OptionalNumberInput
					className={inputClassName(
						fieldInvalid(fieldErrors, 'crawl.max_pages'),
					)}
					value={v.max_pages}
					onChange={(max_pages) => onChange({ ...v, max_pages })}
				/>
			</ConfigField>
			<StringListEditor
				path='crawl.include_paths'
				label='include_paths'
				help={h.include_paths}
				values={v.include_paths ?? []}
				onChange={(include_paths) => onChange({ ...v, include_paths })}
				fieldErrors={fieldErrors}
				compact={compact}
			/>
			<StringListEditor
				path='crawl.exclude_paths'
				label='exclude_paths'
				help={h.exclude_paths}
				values={v.exclude_paths ?? []}
				onChange={(exclude_paths) => onChange({ ...v, exclude_paths })}
				fieldErrors={fieldErrors}
				compact={compact}
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
				<DurationInput
					invalid={fieldInvalid(fieldErrors, 'crawl.request_delay')}
					value={v.request_delay}
					onChange={(request_delay) => onChange({ ...v, request_delay })}
				/>
			</ConfigField>
			<ConfigField
				path='crawl.max_concurrency'
				errors={fieldErrors}
				label='max_concurrency'
				help={h.max_concurrency}
			>
				<OptionalNumberInput
					className={inputClassName(
						fieldInvalid(fieldErrors, 'crawl.max_concurrency'),
					)}
					value={v.max_concurrency}
					onChange={(max_concurrency) => onChange({ ...v, max_concurrency })}
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
					<OptionalNumberInput
						className={inputClassName(
							fieldInvalid(fieldErrors, 'crawl.fetch_limits.http_max_inflight'),
						)}
						value={v.fetch_limits?.http_max_inflight}
						onChange={(http_max_inflight) =>
							onChange({
								...v,
								fetch_limits: {
									...v.fetch_limits,
									http_max_inflight,
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
					<OptionalNumberInput
						className={inputClassName(
							fieldInvalid(
								fieldErrors,
								'crawl.fetch_limits.chromium_max_inflight',
							),
						)}
						value={v.fetch_limits?.chromium_max_inflight}
						onChange={(chromium_max_inflight) =>
							onChange({
								...v,
								fetch_limits: {
									...v.fetch_limits,
									chromium_max_inflight,
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
	const stealth = v.stealth ?? {};
	const httpStealth = stealth.http ?? {};
	const chromiumStealth = stealth.chromium ?? {};
	const waitUntil = (fc.wait_until as string | undefined) ?? 'load';
	const isChromium = (v.fetcher ?? 'http') === 'chromium';

	const patchFetcherConfig = (patch: Record<string, unknown>) =>
		onChange({
			...v,
			fetcher_config: { ...fc, ...patch },
		});

	const patchHTTPStealth = (patch: Record<string, unknown>) =>
		onChange({
			...v,
			stealth: {
				...stealth,
				http: { ...httpStealth, ...patch },
			},
		});

	const patchChromiumStealth = (patch: Record<string, unknown>) =>
		onChange({
			...v,
			stealth: {
				...stealth,
				chromium: { ...chromiumStealth, ...patch },
			},
		});

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
			{isChromium ? (
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
							patchFetcherConfig({ browser_path: e.target.value })
						}
					/>
				</ConfigField>
			) : null}
			{isChromium ? (
				<>
					<ConfigField
						path='plugins.fetcher_config.wait_until'
						errors={fieldErrors}
						label='wait_until'
						help={h.wait_until}
					>
						<select
							className={selectClassName(
								fieldInvalid(fieldErrors, 'plugins.fetcher_config.wait_until'),
								'mt-1 h-8 w-full rounded-lg border border-input bg-background px-2',
							)}
							value={waitUntil}
							onChange={(e) =>
								patchFetcherConfig({ wait_until: e.target.value })
							}
						>
							<option value='none'>none</option>
							<option value='load'>load</option>
							<option value='network_idle'>network_idle</option>
							<option value='selector'>selector</option>
						</select>
					</ConfigField>
					<ConfigField
						path='plugins.fetcher_config.wait_timeout'
						errors={fieldErrors}
						label='wait_timeout'
						help={h.wait_timeout}
					>
						<DurationInput
							invalid={fieldInvalid(
								fieldErrors,
								'plugins.fetcher_config.wait_timeout',
							)}
							value={fc.wait_timeout as string | undefined}
							onChange={(wait_timeout) => patchFetcherConfig({ wait_timeout })}
						/>
					</ConfigField>
					{waitUntil === 'selector' ? (
						<ConfigField
							path='plugins.fetcher_config.wait_visible_selector'
							errors={fieldErrors}
							label='wait_visible_selector'
							help={h.wait_visible_selector}
						>
							<Input
								className={inputClassName(
									fieldInvalid(
										fieldErrors,
										'plugins.fetcher_config.wait_visible_selector',
									),
									'mt-1 h-8',
								)}
								value={(fc.wait_visible_selector as string) ?? ''}
								onChange={(e) =>
									patchFetcherConfig({
										wait_visible_selector: e.target.value,
									})
								}
							/>
						</ConfigField>
					) : null}
					{waitUntil === 'network_idle' ? (
						<ConfigField
							path='plugins.fetcher_config.network_idle_duration'
							errors={fieldErrors}
							label='network_idle_duration'
							help={h.network_idle_duration}
						>
							<DurationInput
								invalid={fieldInvalid(
									fieldErrors,
									'plugins.fetcher_config.network_idle_duration',
								)}
								value={fc.network_idle_duration as string | undefined}
								onChange={(network_idle_duration) =>
									patchFetcherConfig({ network_idle_duration })
								}
							/>
						</ConfigField>
					) : null}
				</>
			) : null}
			<div className='space-y-3 rounded-lg border border-border/60 p-3'>
				<p className='text-xs font-medium text-muted-foreground'>
					ステルス対策
				</p>
				<p className='text-xs text-muted-foreground'>{h.stealth_group}</p>
				{isChromium ? (
					<>
						<ConfigField
							path='plugins.stealth.chromium.user_agent'
							errors={fieldErrors}
							label='user_agent'
							help={h.stealth_chromium_user_agent}
						>
							<Input
								className={inputClassName(
									fieldInvalid(
										fieldErrors,
										'plugins.stealth.chromium.user_agent',
									),
									'mt-1 h-8',
								)}
								value={chromiumStealth.user_agent ?? ''}
								onChange={(e) =>
									patchChromiumStealth({ user_agent: e.target.value })
								}
							/>
						</ConfigField>
						<ConfigCheckboxRow
							inputId={configCheckboxId('plugins.stealth.chromium.headless')}
							checked={chromiumStealth.headless ?? true}
							onCheckedChange={(c) => patchChromiumStealth({ headless: !!c })}
						>
							<FieldLabel label='headless' help={h.stealth_chromium_headless} />
						</ConfigCheckboxRow>
						<ConfigCheckboxRow
							inputId={configCheckboxId(
								'plugins.stealth.chromium.hide_automation',
							)}
							checked={chromiumStealth.hide_automation ?? true}
							onCheckedChange={(c) =>
								patchChromiumStealth({ hide_automation: !!c })
							}
						>
							<FieldLabel
								label='hide_automation'
								help={h.stealth_chromium_hide_automation}
							/>
						</ConfigCheckboxRow>
						<ConfigCheckboxRow
							inputId={configCheckboxId('plugins.stealth.chromium.disable_gpu')}
							checked={chromiumStealth.disable_gpu ?? true}
							onCheckedChange={(c) =>
								patchChromiumStealth({ disable_gpu: !!c })
							}
						>
							<FieldLabel
								label='disable_gpu'
								help={h.stealth_chromium_disable_gpu}
							/>
						</ConfigCheckboxRow>
						<ConfigField
							path='plugins.stealth.chromium.user_data_dir'
							errors={fieldErrors}
							label='user_data_dir'
							help={h.stealth_chromium_user_data_dir}
						>
							<Input
								className={inputClassName(
									fieldInvalid(
										fieldErrors,
										'plugins.stealth.chromium.user_data_dir',
									),
									'mt-1 h-8',
								)}
								value={chromiumStealth.user_data_dir ?? ''}
								onChange={(e) =>
									patchChromiumStealth({ user_data_dir: e.target.value })
								}
							/>
						</ConfigField>
						<ConfigField
							path='plugins.stealth.chromium.lang'
							errors={fieldErrors}
							label='lang'
							help={h.stealth_chromium_lang}
						>
							<LocalePresetSelect
								field='lang'
								value={chromiumStealth.lang}
								onChange={(lang) => patchChromiumStealth({ lang })}
								invalid={fieldInvalid(
									fieldErrors,
									'plugins.stealth.chromium.lang',
								)}
							/>
						</ConfigField>
						<ConfigField
							path='plugins.stealth.chromium.window_width'
							errors={fieldErrors}
							label='window_width'
							help={h.stealth_chromium_window_width}
						>
							<OptionalNumberInput
								className={inputClassName(
									fieldInvalid(
										fieldErrors,
										'plugins.stealth.chromium.window_width',
									),
								)}
								value={chromiumStealth.window_width}
								onChange={(window_width) =>
									patchChromiumStealth({ window_width })
								}
							/>
						</ConfigField>
						<ConfigField
							path='plugins.stealth.chromium.window_height'
							errors={fieldErrors}
							label='window_height'
							help={h.stealth_chromium_window_height}
						>
							<OptionalNumberInput
								className={inputClassName(
									fieldInvalid(
										fieldErrors,
										'plugins.stealth.chromium.window_height',
									),
								)}
								value={chromiumStealth.window_height}
								onChange={(window_height) =>
									patchChromiumStealth({ window_height })
								}
							/>
						</ConfigField>
						<ConfigField
							path='plugins.stealth.chromium.accept_language'
							errors={fieldErrors}
							label='accept_language'
							help={h.stealth_chromium_accept_language}
						>
							<LocalePresetSelect
								field='accept_language'
								value={chromiumStealth.accept_language}
								onChange={(accept_language) =>
									patchChromiumStealth({ accept_language })
								}
								invalid={fieldInvalid(
									fieldErrors,
									'plugins.stealth.chromium.accept_language',
								)}
							/>
						</ConfigField>
					</>
				) : (
					<>
						<ConfigField
							path='plugins.stealth.http.user_agent'
							errors={fieldErrors}
							label='user_agent'
							help={h.stealth_http_user_agent}
						>
							<Input
								className={inputClassName(
									fieldInvalid(fieldErrors, 'plugins.stealth.http.user_agent'),
									'mt-1 h-8',
								)}
								value={httpStealth.user_agent ?? ''}
								onChange={(e) =>
									patchHTTPStealth({ user_agent: e.target.value })
								}
							/>
						</ConfigField>
						<ConfigField
							path='plugins.stealth.http.accept_language'
							errors={fieldErrors}
							label='accept_language'
							help={h.stealth_http_accept_language}
						>
							<LocalePresetSelect
								field='accept_language'
								value={httpStealth.accept_language}
								onChange={(accept_language) =>
									patchHTTPStealth({ accept_language })
								}
								invalid={fieldInvalid(
									fieldErrors,
									'plugins.stealth.http.accept_language',
								)}
							/>
						</ConfigField>
						<ConfigField
							path='plugins.stealth.http.cookie'
							errors={fieldErrors}
							label='cookie'
							help={h.stealth_http_cookie}
						>
							<Input
								className={inputClassName(
									fieldInvalid(fieldErrors, 'plugins.stealth.http.cookie'),
									'mt-1 h-8',
								)}
								value={httpStealth.cookie ?? ''}
								onChange={(e) => patchHTTPStealth({ cookie: e.target.value })}
							/>
						</ConfigField>
					</>
				)}
			</div>
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
			<div className='min-w-0'>
				{tab === 'general' && showMeta && meta && onMetaChange && (
					<div className='space-y-2'>
						<div>
							<FieldLabel label='name' help={h.ws_name} />
							<Input
								className='mt-1 h-8 text-xs'
								value={meta.name}
								onChange={(e) =>
									onMetaChange({ ...meta, name: e.target.value })
								}
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
						compact={compact}
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
						compact={compact}
					/>
				)}
				{tab === 'plugins' && (
					<PluginsConfigFields
						value={settings.plugins}
						onChange={(plugins) => onChange({ ...settings, plugins })}
						fieldErrors={fieldErrors}
					/>
				)}
			</div>
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
