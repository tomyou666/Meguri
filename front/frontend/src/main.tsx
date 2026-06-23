import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import { installExternalLinkDelegation } from '@/lib/externalLinkDelegation';
import App from './App';

installExternalLinkDelegation();

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
	<React.StrictMode>
		<App />
	</React.StrictMode>,
);
