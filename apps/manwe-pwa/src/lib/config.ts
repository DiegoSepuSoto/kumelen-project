import {
	PUBLIC_API_BASE,
	PUBLIC_AUTH_MODE,
	PUBLIC_CMS_BASE,
	PUBLIC_FIREBASE_API_KEY,
	PUBLIC_FIREBASE_APP_ID,
	PUBLIC_FIREBASE_AUTH_DOMAIN,
	PUBLIC_FIREBASE_AUTH_EMULATOR_HOST,
	PUBLIC_FIREBASE_MESSAGING_SENDER_ID,
	PUBLIC_FIREBASE_PROJECT_ID
} from '$app/env/public';
import { normalizeEmulatorHost } from './auth/emulator-host';

const trimSlash = (value: string) => value.replace(/\/$/, '');

/** Base del BFF (`iluvatar`): reemplaza a ELENTARI_HOST de manwe. */
export const apiBase = trimSlash(PUBLIC_API_BASE);

/** Base del CMS: se antepone a las imágenes relativas del cuerpo de una publicación (ILUVATAR_CMS_HOST en manwe). */
export const cmsBase = trimSlash(PUBLIC_CMS_BASE || PUBLIC_API_BASE);

/** `firebase` en producción; `mock` permite desarrollar sin proyecto de Firebase. */
export const authMode: 'firebase' | 'mock' = PUBLIC_AUTH_MODE === 'firebase' ? 'firebase' : 'mock';

export const firebaseConfig = {
	apiKey: PUBLIC_FIREBASE_API_KEY,
	authDomain: PUBLIC_FIREBASE_AUTH_DOMAIN,
	projectId: PUBLIC_FIREBASE_PROJECT_ID,
	appId: PUBLIC_FIREBASE_APP_ID,
	messagingSenderId: PUBLIC_FIREBASE_MESSAGING_SENDER_ID
};

/** Emulador de Firebase Auth (solo desarrollo y pruebas). Vacío = Firebase real. */
export const firebaseAuthEmulatorHost = normalizeEmulatorHost(PUBLIC_FIREBASE_AUTH_EMULATOR_HOST);

export const termsUrl = 'https://github.com/DiegoSepuSoto/manwe/blob/master/TERMS_AND_CONDITIONS.md';

export const feedbackEmail = 'diego.sepulvedas@utem.cl';

export const appVersion = __APP_VERSION__;
