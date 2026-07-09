import { z } from 'zod';
import {
	createDurationRangeRefine,
	isValidGoDuration,
} from '@/components/settings/durationFormUtils';
import { messages } from '@/i18n/messages';

const durationStringSchema = z
	.string()
	.min(1, { message: messages.settings.validation.durationRequired })
	.refine(isValidGoDuration, {
		message: messages.settings.validation.durationInvalid,
	});

const timeoutSchema = durationStringSchema
	.optional()
	.superRefine(createDurationRangeRefine('timeout'));

const retryIntervalSchema = durationStringSchema
	.optional()
	.superRefine(createDurationRangeRefine('retry_interval'));

const requestDelaySchema = durationStringSchema
	.optional()
	.superRefine(createDurationRangeRefine('request_delay'));

const waitTimeoutSchema = durationStringSchema
	.optional()
	.superRefine(createDurationRangeRefine('wait_timeout'));

const networkIdleDurationSchema = durationStringSchema
	.optional()
	.superRefine(createDurationRangeRefine('network_idle_duration'));

const contentFormatSchema = z.enum([
	'markdown',
	'html',
	'raw_html',
	'json',
	'links',
	'metadata',
]);

const optionalInt = (min: number, max: number) =>
	z
		.number({ message: '数値を入力してください' })
		.int({ message: '整数で入力してください' })
		.min(min, { message: `${min}以上で入力してください` })
		.max(max, { message: `${max}以下で入力してください` })
		.optional();

export const requestConfigSchema = z.object({
	headers: z.record(z.string(), z.string()).optional(),
	timeout: timeoutSchema,
	retry_count: optionalInt(0, 10),
	retry_interval: retryIntervalSchema,
});

export const contentConfigSchema = z.object({
	formats: z.array(contentFormatSchema).optional(),
	only_main_content: z.boolean().optional(),
	include_tags: z.array(z.string()).optional(),
	exclude_tags: z.array(z.string()).optional(),
	selector: z.string().optional(),
	extract_links: z.boolean().optional(),
	extract_metadata: z.boolean().optional(),
});

export const pdfConfigSchema = z.object({
	enabled: z.boolean().optional(),
	mode: z.enum(['fast', 'auto', 'ocr']).optional(),
	max_pages: optionalInt(0, 10000),
	output: z.enum(['text', 'markdown', 'raw']).optional(),
});

export const fetchLimitsConfigSchema = z.object({
	http_max_inflight: optionalInt(1, 64),
	chromium_max_inflight: optionalInt(1, 8),
	auto_calibrate: z.boolean().optional(),
	dynamic_chromium: z.boolean().optional(),
	memory_high_watermark: z.number().min(0.5).max(0.95).optional(),
	memory_low_watermark: z.number().min(0.5).max(0.95).optional(),
});

export const crawlConfigSchema = z.object({
	enabled: z.boolean().optional(),
	max_depth: optionalInt(0, 10),
	max_pages: optionalInt(1, 100000),
	include_paths: z.array(z.string()).optional(),
	exclude_paths: z.array(z.string()).optional(),
	allow_external_links: z.boolean().optional(),
	allow_subdomains: z.boolean().optional(),
	request_delay: requestDelaySchema,
	max_concurrency: optionalInt(1, 64),
	respect_robots_txt: z.boolean().optional(),
	fetch_limits: fetchLimitsConfigSchema.optional(),
});

export const fetcherConfigSchema = z
	.object({
		browser_path: z.string().optional(),
		wait_until: z.enum(['none', 'load', 'network_idle', 'selector']).optional(),
		wait_visible_selector: z.string().optional(),
		wait_timeout: waitTimeoutSchema,
		network_idle_duration: networkIdleDurationSchema,
	})
	.superRefine((val, ctx) => {
		if (val.wait_until === 'selector' && !val.wait_visible_selector?.trim()) {
			ctx.addIssue({
				code: 'custom',
				message: messages.settings.validation.waitVisibleSelectorRequired,
				path: ['wait_visible_selector'],
			});
		}
	});

export const httpStealthConfigSchema = z.object({
	user_agent: z.string().optional(),
	accept_language: z.string().optional(),
	cookie: z.string().optional(),
});

export const chromiumStealthConfigSchema = z.object({
	user_agent: z.string().optional(),
	headless: z.boolean().optional(),
	hide_automation: z.boolean().optional(),
	disable_gpu: z.boolean().optional(),
	user_data_dir: z.string().optional(),
	lang: z.string().optional(),
	window_width: optionalInt(0, 7680),
	window_height: optionalInt(0, 7680),
	accept_language: z.string().optional(),
});

export const stealthConfigSchema = z.object({
	http: httpStealthConfigSchema.optional(),
	chromium: chromiumStealthConfigSchema.optional(),
});

export const pluginsConfigSchema = z.object({
	fetcher: z.enum(['http', 'chromium']).optional(),
	fetcher_config: fetcherConfigSchema.optional(),
	stealth: stealthConfigSchema.optional(),
	preprocessors: z.array(z.string()).optional(),
	parsers: z.array(z.string()).optional(),
	transformer: z.enum(['markdown', 'html', 'raw_html', 'json']).optional(),
	filters: z.array(z.string()).optional(),
	link_extractor: z.string().optional(),
});

export const outputConfigSchema = z.object({
	dir: z.string().optional(),
	file_pattern: z.string().optional(),
});

export const partialConfigSchema = z
	.object({
		request: requestConfigSchema.optional(),
		content: contentConfigSchema.optional(),
		pdf: pdfConfigSchema.optional(),
		crawl: crawlConfigSchema.optional(),
		plugins: pluginsConfigSchema.optional(),
		output: outputConfigSchema.optional(),
	})
	.passthrough();

export const appConfigSchema = partialConfigSchema;

export function parsePartialConfig(json: unknown) {
	return partialConfigSchema.safeParse(json);
}
