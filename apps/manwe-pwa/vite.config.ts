import adapter from '@sveltejs/adapter-node';
import { withPwa } from '@vite-pwa/sveltekit';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { defineConfig, loadEnv } from 'vite';

const pkg = JSON.parse(readFileSync(new URL('./package.json', import.meta.url), 'utf-8'));

const escapeRegExp = (value: string) => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), 'PUBLIC_');
	const apiBase = (env.PUBLIC_API_BASE ?? 'http://localhost:8787').replace(/\/$/, '');

	return {
		define: {
			__APP_VERSION__: JSON.stringify(pkg.version)
		},
		resolve: {
			alias: { $lib: fileURLToPath(new URL('./src/lib', import.meta.url)) }
		},
		plugins: [
			withPwa(
				{
					compilerOptions: {
						// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
						runes: ({ filename }) =>
							filename.split(/[/\\]/).includes('node_modules') ? undefined : true
					},
					// El renderizado en servidor (D10) corre sobre Node; el hosting se decide más adelante.
					adapter: adapter()
				},
				{
					registerType: 'autoUpdate',
					manifest: {
						name: 'Kümelen',
						short_name: 'Kümelen',
						description: 'Plataforma de apoyo extraacadémico',
						lang: 'es-CL',
						start_url: '/',
						scope: '/',
						display: 'standalone',
						theme_color: '#459951',
						background_color: '#f1f5f0',
						icons: [
							{ src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
							{ src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png' }
						]
					},
					generateSW: {
						globPatterns: ['client/**/*.{js,css,ico,png,svg,webp,jpg,gif,ttf,webmanifest}'],
						// La app se renderiza en servidor: no hay páginas prerenderizadas que precargar.
						globStrict: false,
						cleanupOutdatedCaches: true,
						clientsClaim: true,
						skipWaiting: true,
						runtimeCaching: [
							{
								// Páginas ya visitadas: abren sin conexión (el armazón y los datos vienen del caché).
								urlPattern: ({ request }) => request.mode === 'navigate',
								handler: 'NetworkFirst',
								options: {
									cacheName: 'manwe-pages',
									networkTimeoutSeconds: 4,
									expiration: { maxEntries: 30 }
								}
							},
							{
								// Respuestas del BFF. Se borra al cerrar sesión (ver lib/offline.ts).
								urlPattern: new RegExp(`^${escapeRegExp(apiBase)}/v1/`),
								handler: 'NetworkFirst',
								options: {
									cacheName: 'manwe-api',
									networkTimeoutSeconds: 5,
									expiration: { maxEntries: 150, maxAgeSeconds: 7 * 24 * 60 * 60 },
									cacheableResponse: { statuses: [200] }
								}
							},
							{
								urlPattern: ({ request }) => request.destination === 'image',
								handler: 'CacheFirst',
								options: {
									cacheName: 'manwe-images',
									expiration: { maxEntries: 120, maxAgeSeconds: 30 * 24 * 60 * 60 },
									cacheableResponse: { statuses: [0, 200] }
								}
							}
						]
					},
					devOptions: { enabled: false }
				}
			)
		]
	};
});
