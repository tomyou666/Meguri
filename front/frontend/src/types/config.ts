export type ContentFormat =
	| 'markdown'
	| 'html'
	| 'raw_html'
	| 'json'
	| 'links'
	| 'metadata';

export interface RequestConfig {
	headers?: Record<string, string>;
	timeout?: string;
	retry_count?: number;
	retry_interval?: string;
}

export interface ContentConfig {
	formats?: ContentFormat[];
	only_main_content?: boolean;
	include_tags?: string[];
	exclude_tags?: string[];
	selector?: string;
	extract_links?: boolean;
	extract_metadata?: boolean;
}

export interface PdfConfig {
	enabled?: boolean;
	mode?: 'fast' | 'auto' | 'ocr';
	max_pages?: number;
	output?: 'text' | 'markdown' | 'raw';
}

export interface FetchLimitsConfig {
	http_max_inflight?: number;
	chromium_max_inflight?: number;
	auto_calibrate?: boolean;
	dynamic_chromium?: boolean;
	memory_high_watermark?: number;
	memory_low_watermark?: number;
}

export interface CrawlConfig {
	enabled?: boolean;
	max_depth?: number;
	max_pages?: number;
	include_paths?: string[];
	exclude_paths?: string[];
	allow_external_links?: boolean;
	allow_subdomains?: boolean;
	request_delay?: string;
	max_concurrency?: number;
	respect_robots_txt?: boolean;
	fetch_limits?: FetchLimitsConfig;
}

export interface PluginsConfig {
	fetcher?: string;
	fetcher_config?: Record<string, unknown>;
	preprocessors?: string[];
	parsers?: string[];
	transformer?: string;
	filters?: string[];
	link_extractor?: string;
}

export interface OutputConfig {
	dir?: string;
	file_pattern?: string;
}

export interface AppConfig {
	request?: RequestConfig;
	content?: ContentConfig;
	pdf?: PdfConfig;
	crawl?: CrawlConfig;
	plugins?: PluginsConfig;
	output?: OutputConfig;
}

export type PartialConfig = {
	[K in keyof AppConfig]?: AppConfig[K] extends object
		? Partial<AppConfig[K]>
		: AppConfig[K];
};
