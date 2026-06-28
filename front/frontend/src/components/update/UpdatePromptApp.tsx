import { Events } from '@wailsio/runtime';
import { useCallback, useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { messages } from '@/i18n/messages';
import {
	PROMPT_ACTION_CONFIRMED,
	PROMPT_ACTION_DISMISSED,
	PROMPT_ACTION_OPEN_RELEASE,
} from '@/lib/updateFlow';
import * as UpdateService from '../../../bindings/meguri-app/internal/usecase/wails_service/updateservice';

const TOPIC_UPDATE_PROMPT_OPEN = 'update-prompt:open';

type UpdatePromptSnapshot = {
	version: string;
	releaseURL: string;
};

function snapshotFromEventData(data: unknown): UpdatePromptSnapshot | null {
	if (!data || typeof data !== 'object') return null;
	const raw = data as Record<string, unknown>;
	if (!raw.version) return null;
	return {
		version: String(raw.version),
		releaseURL: String(raw.releaseURL ?? ''),
	};
}

export function UpdatePromptApp() {
	const [snapshot, setSnapshot] = useState<UpdatePromptSnapshot | null>(null);
	const [submitting, setSubmitting] = useState(false);

	useEffect(() => {
		void UpdateService.GetUpdatePromptSnapshot().then((initial) => {
			if (initial.version) {
				setSnapshot({
					version: initial.version,
					releaseURL: initial.releaseURL ?? '',
				});
			}
		});

		const off = Events.On(TOPIC_UPDATE_PROMPT_OPEN, (ev) => {
			const next = snapshotFromEventData(ev.data);
			if (next) setSnapshot(next);
		});
		return () => off();
	}, []);

	const submit = useCallback(
		async (action: string) => {
			if (submitting) return;
			setSubmitting(true);
			try {
				await UpdateService.SubmitUpdatePrompt(action);
			} finally {
				setSubmitting(false);
			}
		},
		[submitting],
	);

	if (!snapshot) {
		return (
			<div className='flex h-screen items-center justify-center text-sm text-muted-foreground'>
				Loading…
			</div>
		);
	}

	return (
		<div className='flex h-screen flex-col bg-background p-6 text-foreground'>
			<h1 className='text-lg font-semibold'>{messages.update.promptTitle}</h1>
			<p className='mt-3 text-sm text-muted-foreground'>
				{messages.update.promptBody(snapshot.version)}
			</p>
			{snapshot.releaseURL ? (
				<p className='mt-4 text-sm'>
					<span className='text-muted-foreground'>
						{messages.update.promptReleaseLabel}:{' '}
					</span>
					<a
						href={snapshot.releaseURL}
						className='text-primary underline-offset-4 hover:underline'
					>
						{snapshot.releaseURL}
					</a>
				</p>
			) : null}
			<div className='mt-auto flex flex-row gap-2 pt-6'>
				<Button
					className='flex-1'
					disabled={submitting}
					onClick={() => void submit(PROMPT_ACTION_CONFIRMED)}
				>
					{messages.update.confirmAndRestart}
				</Button>
				<Button
					className='flex-1'
					variant='secondary'
					disabled={submitting}
					onClick={() => void submit(PROMPT_ACTION_OPEN_RELEASE)}
				>
					{messages.update.openRelease}
				</Button>
				<Button
					className='flex-1'
					variant='outline'
					disabled={submitting}
					onClick={() => void submit(PROMPT_ACTION_DISMISSED)}
				>
					{messages.update.later}
				</Button>
			</div>
		</div>
	);
}
