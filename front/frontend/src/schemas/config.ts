import { z } from 'zod';

const durationSchema = z
	.string()
	.min(1, { message: '時間を入力してください（例: 30s）' });

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
	timeout: durationSchema.optional(),
	retry_count: optionalInt(0, 10),
	retry_interval: durationSchema.optional(),
});

export const contentConfigSchema = z.object({
	formats: z
		.array(contentFormatSchema)
		.min(1, { message: '保存形式を1つ以上選んでください' })
		.optional(),
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

export const crawlConfigSchema = z.object({
	enabled: z.boolean().optional(),
	max_depth: optionalInt(0, 10),
	max_pages: optionalInt(1, 100000),
	include_paths: z.array(z.string()).optional(),
	exclude_paths: z.array(z.string()).optional(),
	allow_external_links: z.boolean().optional(),
	allow_subdomains: z.boolean().optional(),
	request_delay: durationSchema.optional(),
	max_concurrency: optionalInt(1, 64),
	respect_robots_txt: z.boolean().optional(),
});

export const fetcherConfigSchema = z.object({
	browser_path: z.string().optional(),
	user_agent: z.string().optional(),
	headless: z.boolean().optional(),
	wait_visible_selector: z.string().optional(),
	wait_timeout: durationSchema.optional(),
});

export const pluginsConfigSchema = z.object({
	fetcher: z.enum(['http', 'chromium']).optional(),
	fetcher_config: fetcherConfigSchema.optional(),
	preprocessors: z.array(z.string()).optional(),
	parsers: z.array(z.string()).optional(),
	transformer: z.string().optional(),
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
