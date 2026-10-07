import { X } from 'lucide-react';
import { toast } from 'sonner';
import { summaryBadgeLabels } from '@/components/diff/diffSummaryUtils';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { messages } from '@/i18n/messages';
import { setDismissedUpdateToastVersion } from '@/lib/updatePreferences';
import type { WorkspaceDiff } from '@/types/adapter';

export const TOAST_DURATION_MS = 5_000;
export const TOAST_ERROR_DURATION_MS = 10_000;

/** 既知のバックエンド文言をユーザー向け日本語に差し替える */
const KNOWN_ERROR_MESSAGES: Record<string, string> = {
	'export preview session expired': messages.export.previewSessionExpired,
};

function toastDismissCancel(getToastId: () => string | number) {
	return {
		label: (
			<>
				<span className='sr-only'>{messages.toast.dismissAria}</span>
				<X className='size-4' aria-hidden />
			</>
		),
		onClick: () => toast.dismiss(getToastId()),
	};
}

function DiffToastBadges({ summary }: { summary: WorkspaceDiff['summary'] }) {
	const badges = summaryBadgeLabels(summary);
	return (
		<div className='flex flex-wrap gap-1 pt-0.5'>
			<Badge variant='outline' className='text-[10px]'>
				{badges.content}
			</Badge>
			<Badge variant='outline' className='text-[10px]'>
				{badges.links}
			</Badge>
			<Badge variant='outline' className='text-[10px]'>
				{badges.fetch}
			</Badge>
		</div>
	);
}

/** Wails RuntimeError の JSON 文字列から message を取り出す */
function extractRuntimeErrorMessage(raw: string): string | null {
	const trimmed = raw.trim();
	if (!trimmed.startsWith('{')) return null;
	try {
		const parsed = JSON.parse(trimmed) as { message?: unknown };
		return typeof parsed.message === 'string' && parsed.message.trim()
			? parsed.message.trim()
			: null;
	} catch {
		return null;
	}
}

/** 通知に出す文言を読みやすく整形する */
export function formatNotifyMessage(raw: string): string {
	const extracted = extractRuntimeErrorMessage(raw) ?? raw.trim();
	if (!extracted) return messages.error.unknown;
	return KNOWN_ERROR_MESSAGES[extracted] ?? extracted;
}

function showNotifyToast(
	type: 'error' | 'success',
	title: string,
	duration: number,
	description?: string,
): void {
	const show = type === 'error' ? toast.error : toast.success;
	let toastId: string | number;
	toastId = show(title, {
		description,
		duration,
		cancel: toastDismissCancel(() => toastId),
	});
}

export function notifyError(
	title: string,
	options?: { description?: string },
): void {
	showNotifyToast(
		'error',
		formatNotifyMessage(title),
		TOAST_ERROR_DURATION_MS,
		options?.description != null
			? formatNotifyMessage(options.description)
			: undefined,
	);
}

export function notifySuccess(
	title: string,
	options?: { description?: string },
): void {
	showNotifyToast(
		'success',
		title,
		TOAST_DURATION_MS,
		options?.description != null
			? formatNotifyMessage(options.description)
			: undefined,
	);
}

export function notifyDiffDetected(
	title: string,
	actionLabel: string,
	onViewDetails: () => void,
	summary: WorkspaceDiff['summary'],
): void {
	let toastId: string | number;
	toastId = toast.warning(title, {
		description: <DiffToastBadges summary={summary} />,
		duration: TOAST_DURATION_MS,
		action: {
			label: actionLabel,
			onClick: () => {
				onViewDetails();
				toast.dismiss(toastId);
			},
		},
		cancel: toastDismissCancel(() => toastId),
	});
}

function UpdateToastDescription({
	dismissRef,
}: {
	dismissRef: React.MutableRefObject<boolean>;
}) {
	const checkboxId = 'update-toast-dismiss';
	return (
		<div className='flex items-center gap-2 pt-1 text-muted-foreground text-xs'>
			<Checkbox
				id={checkboxId}
				onCheckedChange={(checked) => {
					dismissRef.current = checked === true;
				}}
			/>
			<label htmlFor={checkboxId} className='cursor-pointer'>
				{messages.update.toastDismissLabel}
			</label>
		</div>
	);
}

export function notifyUpdateAvailable(
	version: string,
	onDetails: () => void,
): void {
	const dismissRef = { current: false };
	let toastId: string | number;
	toastId = toast.info(messages.update.toastTitle(version), {
		description: <UpdateToastDescription dismissRef={dismissRef} />,
		duration: Number.POSITIVE_INFINITY,
		action: {
			label: messages.update.toastAction,
			onClick: () => {
				onDetails();
				toast.dismiss(toastId);
			},
		},
		cancel: toastDismissCancel(() => toastId),
		onDismiss: () => {
			if (dismissRef.current) {
				setDismissedUpdateToastVersion(version);
			}
		},
	});
}
