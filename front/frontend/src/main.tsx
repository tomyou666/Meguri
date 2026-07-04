import React from 'react';
import ReactDOM from 'react-dom/client';
import { scan } from 'react-scan';
import './index.css';
import { installExternalLinkDelegation } from '@/lib/externalLinkDelegation';
import App from './App';

if (import.meta.env.DEV) {
	scan({ enabled: true });
}

installExternalLinkDelegation();

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
	<React.StrictMode>
		<App />
	</React.StrictMode>,
);
