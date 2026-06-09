import { useEffect, useState } from 'react';
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { messages } from '@/i18n/messages';
import { useAppStore } from '@/stores/appStore';

export function AppDialogs() {
	const showNewWorkspaceDialog = useAppStore((s) => s.showNewWorkspaceDialog);
	const closeNewWorkspaceDialog = useAppStore((s) => s.closeNewWorkspaceDialog);
	const createWorkspace = useAppStore((s) => s.createWorkspace);
	const workspaces = useAppStore((s) => s.workspaces);

	const showAddNodeDialog = useAppStore((s) => s.showAddNodeDialog);
	const closeAddNodeDialog = useAppStore((s) => s.closeAddNodeDialog);
	const addNode = useAppStore((s) => s.addNode);

	const showDeleteNodeDialog = useAppStore((s) => s.showDeleteNodeDialog);
	const closeDeleteNodeDialog = useAppStore((s) => s.closeDeleteNodeDialog);
	const deleteSelectedSubtree = useAppStore((s) => s.deleteSelectedSubtree);
	const selectedNode = useAppStore((s) => s.getSelectedNode());
	const ws = useAppStore((s) => s.getActiveWorkspace());

	const pendingDeleteWorkspaceId = useAppStore(
		(s) => s.pendingDeleteWorkspaceId,
	);
	const closeDeleteWorkspaceDialog = useAppStore(
		(s) => s.closeDeleteWorkspaceDialog,
	);
	const confirmDeleteWorkspace = useAppStore((s) => s.confirmDeleteWorkspace);

	const pendingDuplicateWorkspaceId = useAppStore(
		(s) => s.pendingDuplicateWorkspaceId,
	);
	const closeDuplicateWorkspaceDialog = useAppStore(
		(s) => s.closeDuplicateWorkspaceDialog,
	);
	const confirmDuplicateWorkspace = useAppStore(
		(s) => s.confirmDuplicateWorkspace,
	);

	const pendingDeleteWorkspace = workspaces.find(
		(w) => w.id === pendingDeleteWorkspaceId,
	);

	const [wsName, setWsName] = useState('My Workspace');
	const [wsUrl, setWsUrl] = useState('https://example.com/');
	const [nodeUrl, setNodeUrl] = useState('https://');
	const [duplicateName, setDuplicateName] = useState('');

	useEffect(() => {
		if (!pendingDuplicateWorkspaceId) return;
		const source = workspaces.find((w) => w.id === pendingDuplicateWorkspaceId);
		setDuplicateName(source?.name ?? '');
	}, [pendingDuplicateWorkspaceId, workspaces]);

	const mustShowNewWs = showNewWorkspaceDialog || workspaces.length === 0;

	return (
		<>
			<Dialog
				open={mustShowNewWs}
				onOpenChange={(open) => {
					if (!open && workspaces.length > 0) closeNewWorkspaceDialog();
				}}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{messages.dialog.newWorkspaceTitle}</DialogTitle>
					</DialogHeader>
					<div className='space-y-3'>
						<div>
							<Label>{messages.dialog.newWorkspaceName}</Label>
							<Input
								className='mt-1'
								value={wsName}
								onChange={(e) => setWsName(e.target.value)}
							/>
						</div>
						<div>
							<Label>{messages.dialog.newWorkspaceUrl}</Label>
							<Input
								className='mt-1'
								value={wsUrl}
								onChange={(e) => setWsUrl(e.target.value)}
							/>
						</div>
					</div>
					<DialogFooter>
						{workspaces.length > 0 && (
							<Button
								variant='outline'
								size='sm'
								onClick={closeNewWorkspaceDialog}
							>
								{messages.dialog.cancel}
							</Button>
						)}
						<Button size='sm' onClick={() => createWorkspace(wsName, wsUrl)}>
							{messages.dialog.create}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<Dialog open={showAddNodeDialog} onOpenChange={closeAddNodeDialog}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{messages.dialog.addNodeTitle}</DialogTitle>
					</DialogHeader>
					<Label>{messages.dialog.addNodeUrl}</Label>
					<Input
						className='mt-1'
						value={nodeUrl}
						onChange={(e) => setNodeUrl(e.target.value)}
					/>
					<DialogFooter>
						<Button variant='outline' size='sm' onClick={closeAddNodeDialog}>
							{messages.dialog.cancel}
						</Button>
						<Button size='sm' onClick={() => addNode(nodeUrl)} disabled={!ws}>
							{messages.dialog.add}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<Dialog open={showDeleteNodeDialog} onOpenChange={closeDeleteNodeDialog}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{messages.dialog.deleteNodeTitle}</DialogTitle>
					</DialogHeader>
					<p className='text-sm'>{messages.dialog.deleteNodeConfirm}</p>
					{selectedNode && (
						<p className='mt-2 truncate text-xs text-muted-foreground'>
							{selectedNode.urlNormalized}
						</p>
					)}
					<DialogFooter>
						<Button variant='outline' size='sm' onClick={closeDeleteNodeDialog}>
							{messages.dialog.cancel}
						</Button>
						<Button
							variant='destructive'
							size='sm'
							onClick={deleteSelectedSubtree}
						>
							{messages.dialog.delete}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<AlertDialog
				open={!!pendingDeleteWorkspaceId}
				onOpenChange={(open) => {
					if (!open) closeDeleteWorkspaceDialog();
				}}
			>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>
							{messages.dialog.deleteWorkspaceTitle}
						</AlertDialogTitle>
						<AlertDialogDescription>
							{messages.dialog.deleteWorkspaceConfirm}
							{pendingDeleteWorkspace && (
								<span className='mt-2 block font-medium text-foreground'>
									{pendingDeleteWorkspace.name}
								</span>
							)}
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>{messages.dialog.cancel}</AlertDialogCancel>
						<AlertDialogAction
							variant='destructive'
							onClick={() => void confirmDeleteWorkspace()}
						>
							{messages.dialog.delete}
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>

			<Dialog
				open={!!pendingDuplicateWorkspaceId}
				onOpenChange={(open) => {
					if (!open) closeDuplicateWorkspaceDialog();
				}}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{messages.dialog.duplicateWorkspaceTitle}</DialogTitle>
					</DialogHeader>
					<Label>{messages.dialog.duplicateWorkspaceName}</Label>
					<Input
						className='mt-1'
						value={duplicateName}
						onChange={(e) => setDuplicateName(e.target.value)}
					/>
					<DialogFooter>
						<Button
							variant='outline'
							size='sm'
							onClick={closeDuplicateWorkspaceDialog}
						>
							{messages.dialog.cancel}
						</Button>
						<Button
							size='sm'
							disabled={!duplicateName.trim()}
							onClick={() => void confirmDuplicateWorkspace(duplicateName)}
						>
							{messages.dialog.copy}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
}
