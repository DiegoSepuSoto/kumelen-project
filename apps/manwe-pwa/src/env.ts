import { defineEnvVars } from '@sveltejs/kit/env';

const withDefault = (fallback: string) => (value: string | undefined) => value ?? fallback;

/** Variables que la app lee en el navegador (ver .env.example). */
export const variables = defineEnvVars({
	PUBLIC_API_BASE: { public: true, schema: withDefault('http://localhost:8787'), description: 'Base del BFF (iluvatar)' },
	PUBLIC_CMS_BASE: { public: true, schema: withDefault(''), description: 'Base del CMS para imágenes relativas; vacía usa PUBLIC_API_BASE' },
	PUBLIC_AUTH_MODE: { public: true, schema: withDefault('mock'), description: '`mock` o `firebase`' },
	PUBLIC_FIREBASE_API_KEY: { public: true, schema: withDefault('') },
	PUBLIC_FIREBASE_AUTH_DOMAIN: { public: true, schema: withDefault('') },
	PUBLIC_FIREBASE_PROJECT_ID: { public: true, schema: withDefault('') },
	PUBLIC_FIREBASE_APP_ID: { public: true, schema: withDefault('') },
	PUBLIC_FIREBASE_MESSAGING_SENDER_ID: { public: true, schema: withDefault('') },
	PUBLIC_FIREBASE_AUTH_EMULATOR_HOST: {
		public: true,
		schema: withDefault(''),
		description: 'host:puerto del emulador de Firebase Auth tal como lo ve el navegador; vacío usa Firebase real'
	}
});
