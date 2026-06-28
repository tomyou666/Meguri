import { Events } from '@wailsio/runtime';
import { useCallback, useEffect, useState } from 'react';
import { notifyUpdateAvailable } from '@/lib/notify';
import {
	TOPIC_UPDATE_AVAILABLE,
	type UpdateAvailablePayload,
} from '@/lib/updateEvents';
import { handleUpdatePromptResult } from '@/lib/updateFlow';
import { isUpdateToastDismissedForVersion } from '@/lib/updatePreferences';
import * as UpdateService from '../../bindings/meguri-app/internal/usecase/wails_service/updateservice';

function isUpdatePending(status: string | undefined): boolean {
	return status === 'available' || status === 'ready';
}

export function useUpdateAvailability() {
	const [updateAvailable, setUpdateAvailable] = useState(false);

	const refreshStatus = useCallback(async () => {
		try {
			const status = await UpdateService.GetStatus();
			setUpdateAvailable(isUpdatePending(status.status));
		} catch {
			setUpdateAvailable(false);
		}
	}, []);

	const openUpdatePrompt = useCallback(async () => {
		const prompt = await UpdateService.PromptUpdate();
		await handleUpdatePromptResult(prompt.action, prompt.releaseURL);
		await refreshStatus();
	}, [refreshStatus]);

	useEffect(() => {
		void refreshStatus();
	}, [refreshStatus]);

	useEffect(() => {
		const off = Events.On(TOPIC_UPDATE_AVAILABLE, (ev) => {
			const data = (ev.data ?? ev) as UpdateAvailablePayload;
			const version = data.version;
			if (!version) {
				void refreshStatus();
				return;
			}
			setUpdateAvailable(true);
			if (isUpdateToastDismissedForVersion(version)) {
				return;
			}
			notifyUpdateAvailable(version, () => {
				void openUpdatePrompt();
			});
		});
		return off;
	}, [openUpdatePrompt, refreshStatus]);

	return { updateAvailable, refreshStatus, openUpdatePrompt };
}
