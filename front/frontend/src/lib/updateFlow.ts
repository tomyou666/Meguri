import { messages } from '@/i18n/messages';
import { openExternalBrowserUrl } from '@/lib/externalLinkDelegation';
import { notifyError } from '@/lib/notify';
import * as UpdateService from '../../bindings/meguri-app/internal/usecase/wails_service/updateservice';

export const PROMPT_ACTION_CONFIRMED = 'confirmed';
export const PROMPT_ACTION_OPEN_RELEASE = 'open_release';
export const PROMPT_ACTION_DISMISSED = 'dismissed';

export async function handleUpdatePromptResult(
	action: string | undefined,
	releaseURL: string | undefined,
): Promise<void> {
	if (!action || action === PROMPT_ACTION_DISMISSED) {
		return;
	}
	if (action === PROMPT_ACTION_OPEN_RELEASE) {
		if (releaseURL) {
			await openExternalBrowserUrl(releaseURL);
		}
		return;
	}
	if (action === PROMPT_ACTION_CONFIRMED) {
		try {
			await UpdateService.ApplyUpdate();
		} catch (e) {
			const msg = e instanceof Error ? e.message : String(e);
			notifyError(messages.update.applyFailed, { description: msg });
			throw e;
		}
	}
}
