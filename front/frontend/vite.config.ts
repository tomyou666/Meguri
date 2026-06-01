import path from 'node:path';
import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react-swc';
import wails from '@wailsio/runtime/plugins/vite';
import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
	server: {
		host: '127.0.0.1',
		port: Number(process.env.WAILS_VITE_PORT) || 9245,
		strictPort: true,
	},
	build: {
		sourcemap: mode === 'development',
		rolldownOptions: {
			output: {
				codeSplitting: {
					groups: [
						{
							name: 'vendor',
							entries: ['react', 'react-dom'],
						},
					],
				},
			},
		},
	},
	plugins: [react(), tailwindcss(), wails('./bindings')],
	resolve: {
		alias: {
			'@': path.resolve(__dirname, './src'),
		},
	},
}));
