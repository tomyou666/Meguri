import { AppShell } from '@/components/layout/AppShell';
import { Toaster } from '@/components/ui/sonner';

function App() {
	return (
		<>
			<AppShell />
			<Toaster duration={5000} />
		</>
	);
}

export default App;
