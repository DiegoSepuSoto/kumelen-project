import { svelte } from '@sveltejs/vite-plugin-svelte';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [svelte()],
	define: { __APP_VERSION__: JSON.stringify('0.0.0-test') },
	resolve: {
		alias: {
			$lib: fileURLToPath(new URL('./src/lib', import.meta.url)),
			'$app/env/public': fileURLToPath(new URL('./tests/stubs/env.ts', import.meta.url))
		},
		conditions: ['browser']
	},
	test: {
		environment: 'jsdom',
		include: ['tests/**/*.test.ts']
	}
});
