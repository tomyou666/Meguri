import { AppShell } from '@/components/layout/AppShell';
import { MaximizedNodeResultApp } from '@/components/layout/node-result/MaximizedNodeResultApp';
import { Toaster } from '@/components/ui/sonner';

const isMaximizedNodeResultView =
	new URLSearchParams(window.location.search).get('view') ===
	'maximized-node-result';

function App() {
	if (isMaximizedNodeResultView) {
		return <MaximizedNodeResultApp />;
	}

	return (
		<>
			<AppShell />
			<Toaster duration={5000} />
		</>
	);
}

export default App;
