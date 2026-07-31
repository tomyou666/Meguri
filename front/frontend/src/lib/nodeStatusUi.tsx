import {
	AlertCircle,
	CheckCircle2,
	Circle,
	Loader2,
	SkipForward,
} from 'lucide-react';
import { createElement, type ReactNode } from 'react';
import { messages } from '@/i18n/messages';
import type { NodeStatus } from '@/types/graph';

export type NodeStatusUi = {
	icon: ReactNode;
	border: string;
	textClass: string;
	label: string;
};

const iconClass = 'size-3 shrink-0';

/** グラフ・ツリー共通の status 見た目。 */
export const nodeStatusUi: Record<NodeStatus, NodeStatusUi> = {
	idle: {
		icon: createElement(Circle, {
			className: `${iconClass} text-muted-foreground`,
		}),
		border: 'border-border',
		textClass: 'text-muted-foreground',
		label: messages.status.idle,
	},
	running: {
		icon: createElement(Loader2, {
			className: `${iconClass} animate-spin text-blue-400`,
		}),
		border: 'border-blue-500',
		textClass: 'text-blue-400',
		label: messages.status.running,
	},
	success: {
		icon: createElement(CheckCircle2, {
			className: `${iconClass} text-emerald-400`,
		}),
		border: 'border-emerald-500',
		textClass: 'text-emerald-400',
		label: messages.status.success,
	},
	error: {
		icon: createElement(AlertCircle, {
			className: `${iconClass} text-destructive`,
		}),
		border: 'border-destructive',
		textClass: 'text-destructive',
		label: messages.status.error,
	},
	skipped: {
		icon: createElement(SkipForward, {
			className: `${iconClass} text-amber-400`,
		}),
		border: 'border-amber-500',
		textClass: 'text-amber-400',
		label: messages.status.skipped,
	},
};
