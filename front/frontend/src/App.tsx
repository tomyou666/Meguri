import { AppShell } from '@/components/layout/AppShell';
import { ExportApp } from '@/components/layout/export/ExportApp';
import { MaximizedNodeResultApp } from '@/components/layout/node-result/MaximizedNodeResultApp';
import { Toaster } from '@/components/ui/sonner';

const viewParam = new URLSearchParams(window.location.search).get('view');
const isMaximizedNodeResultView = viewParam === 'maximized-node-result';
const isExportView = viewParam === 'export';

function App() {
	if (isMaximizedNodeResultView) {
		return <MaximizedNodeResultApp />;
	}
	if (isExportView) {
		return <ExportApp />;
	}

	return (
		<>
			<AppShell />
			<Toaster duration={5000} />
		</>
	);
}

export default App;
