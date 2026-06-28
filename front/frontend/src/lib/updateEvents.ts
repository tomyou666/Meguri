export const TOPIC_UPDATE_AVAILABLE = 'meguri:update:available';

export type UpdateAvailablePayload = {
	version?: string;
	releaseURL?: string;
};
