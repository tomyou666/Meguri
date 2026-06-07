import { X } from 'lucide-react';
import { toast } from 'sonner';

export const TOAST_DURATION_MS = 5_000;
export const TOAST_ERROR_DURATION_MS = 10_000;

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
		action: {
			label: <X className='size-5' />,
			onClick: () => toast.dismiss(toastId),
		},
	});
}

export function notifyError(
	title: string,
	options?: { description?: string },
): void {
	showNotifyToast(
		'error',
		title,
		TOAST_ERROR_DURATION_MS,
		options?.description,
	);
}

export function notifySuccess(
	title: string,
	options?: { description?: string },
): void {
	showNotifyToast('success', title, TOAST_DURATION_MS, options?.description);
}
