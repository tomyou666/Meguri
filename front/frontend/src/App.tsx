import { NodeDiffApp } from '@/components/diff/NodeDiffApp';
import { AppShell } from '@/components/layout/AppShell';
import { ExportApp } from '@/components/layout/export/ExportApp';
import { MaximizedNodeResultApp } from '@/components/layout/node-result/MaximizedNodeResultApp';
import { Toaster } from '@/components/ui/sonner';
import { UpdatePromptApp } from '@/components/update/UpdatePromptApp';

const viewParam = new URLSearchParams(window.location.search).get('view');
const isMaximizedNodeResultView = viewParam === 'maximized-node-result';
const isExportView = viewParam === 'export';
const isNodeDiffView = viewParam === 'node-diff';
const isUpdatePromptView = viewParam === 'update-prompt';

function App() {
	if (isMaximizedNodeResultView) {
		return <MaximizedNodeResultApp />;
	}
	if (isExportView) {
		return <ExportApp />;
	}
	if (isNodeDiffView) {
		return <NodeDiffApp />;
	}
	if (isUpdatePromptView) {
		return <UpdatePromptApp />;
	}

	return (
		<>
			<AppShell />
			<Toaster duration={5000} />
		</>
	);
}

export default App;
