import type { ZodError, ZodIssue } from 'zod';
import { messages } from '@/i18n/messages';
import { parsePartialConfig } from '@/schemas/config';
import type { PartialConfig } from '@/types/config';

const FIELD_LABELS: Record<string, string> = {
	'request.timeout': 'タイムアウト',
	'request.retry_count': '再試行回数',
	'request.retry_interval': '再試行間隔',
	'request.headers': 'HTTP ヘッダー',
	'content.formats': '保存形式',
	'content.selector': 'セレクタ',
	'pdf.max_pages': 'PDF 最大ページ数',
	'pdf.mode': 'PDF モード',
	'pdf.output': 'PDF 出力形式',
	'crawl.max_depth': '最大深度',
	'crawl.max_pages': '最大ページ数',
	'crawl.request_delay': 'リクエスト間隔',
	'crawl.max_concurrency': '同時取得数',
	'plugins.fetcher': '取得方式 (fetcher)',
	'plugins.fetcher_config.browser_path': 'ブラウザパス',
	'output.dir': '出力フォルダ',
	'output.file_pattern': 'ファイル名パターン',
};

function pathKey(path: PropertyKey[]): string {
	return path.map(String).join('.');
}

function labelForPath(path: PropertyKey[]): string {
	const key = pathKey(path);
	return FIELD_LABELS[key] ?? key;
}

/** 英語のデフォルトメッセージを日本語に補正（Zod 4 フォールバック） */
function translateIssueMessage(issue: ZodIssue): string {
	const msg = issue.message;
	if (issue.code === 'too_big' && 'maximum' in issue) {
		const origin = issue.origin;
		if (origin === 'number' || origin === 'int') {
			return `${issue.maximum}以下の数値を入力してください`;
		}
		if (origin === 'array') {
			return `${issue.maximum}件以下にしてください`;
		}
	}
	if (issue.code === 'too_small' && 'minimum' in issue) {
		const origin = issue.origin;
		if (origin === 'number' || origin === 'int') {
			return `${issue.minimum}以上の数値を入力してください`;
		}
		if (origin === 'array') {
			return `${issue.minimum}件以上選択してください`;
		}
	}
	if (issue.code === 'invalid_type') {
		if (issue.input !== undefined && Number.isNaN(issue.input)) {
			return '数値を入力してください';
		}
		if (issue.expected === 'number') {
			return '数値を入力してください';
		}
	}
	if (/Too big|too_big/i.test(msg) && 'maximum' in issue) {
		return `${issue.maximum}以下の数値を入力してください`;
	}
	if (/Too small|too_small/i.test(msg) && 'minimum' in issue) {
		return `${issue.minimum}以上の数値を入力してください`;
	}
	return msg;
}

function issueMessage(issue: ZodIssue): string {
	return translateIssueMessage(issue);
}

/** フィールドパス → エラーメッセージ（入力のたびに表示用） */
export function getConfigFieldErrors(
	config: PartialConfig,
): Record<string, string> {
	const result = parsePartialConfig(config);
	if (result.success) return {};
	const errors: Record<string, string> = {};
	for (const issue of result.error.issues) {
		const key = pathKey(issue.path);
		if (!errors[key]) errors[key] = issueMessage(issue);
	}
	return errors;
}

export function validatePartialConfig(
	config: PartialConfig,
): { ok: true; data: PartialConfig } | { ok: false; errors: string[] } {
	const result = parsePartialConfig(config);
	if (result.success) {
		return { ok: true, data: result.data };
	}
	return { ok: false, errors: formatZodErrors(result.error) };
}

export function formatZodErrors(error: ZodError): string[] {
	return error.issues.map((issue) => {
		const label = labelForPath(issue.path);
		return `${label}: ${issueMessage(issue)}`;
	});
}

export function hasConfigFieldErrors(config: PartialConfig): boolean {
	return Object.keys(getConfigFieldErrors(config)).length > 0;
}

export function configValidationSummary(config: PartialConfig): string | null {
	if (!hasConfigFieldErrors(config)) return null;
	return messages.settings.validationFailed;
}
