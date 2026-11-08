import { firebaseAuthEmulatorHost, firebaseConfig } from '$lib/config';
import type { AuthProvider, AuthUser } from './types';

function toAuthUser(user: { uid: string; displayName: string | null; email: string | null }): AuthUser {
	return { uid: user.uid, name: user.displayName ?? '', email: user.email ?? '' };
}

export interface FirebaseAuthOptions {
	config?: typeof firebaseConfig;
	/** `host:puerto` del emulador, tal como lo ve el navegador. Vacío usa Firebase real. */
	emulatorHost?: string;
}

/**
 * Firebase Auth con Google (D5, sin restringir dominio). El SDK se carga bajo demanda para no
 * penalizar el peso inicial de la app.
 *
 * Con `emulatorHost` definido se conecta al emulador de Auth antes de usarlo. El emulador no pide una
 * clave, proyecto ni dominio reales, así que se completan con valores de relleno si faltan; con el
 * emulador vacío la configuración se usa tal cual.
 */
export function createFirebaseAuth({
	config = firebaseConfig,
	emulatorHost = firebaseAuthEmulatorHost
}: FirebaseAuthOptions = {}): AuthProvider {
	const sdk = (async () => {
		const [{ initializeApp, getApps }, auth] = await Promise.all([
			import('firebase/app'),
			import('firebase/auth')
		]);

		const appConfig = emulatorHost
			? {
					...config,
					apiKey: config.apiKey || 'demo-api-key',
					projectId: config.projectId || 'demo-kumelen',
					// El SDK exige un authDomain para abrir el popup, aunque con el emulador no se use.
					authDomain: config.authDomain || 'localhost'
				}
			: config;
		const app = getApps()[0] ?? initializeApp(appConfig);
		const instance = auth.getAuth(app);

		if (emulatorHost) {
			auth.connectAuthEmulator(instance, `http://${emulatorHost}`, { disableWarnings: true });
		}

		return { auth: instance, ...auth };
	})();

	return {
		subscribe(onChange) {
			let stop = () => {};
			let cancelled = false;
			sdk.then(({ auth, onAuthStateChanged }) => {
				if (cancelled) return;
				stop = onAuthStateChanged(auth, (user) => onChange(user ? toAuthUser(user) : null));
			});
			return () => {
				cancelled = true;
				stop();
			};
		},
		async signIn() {
			const { auth, GoogleAuthProvider, signInWithPopup } = await sdk;
			const result = await signInWithPopup(auth, new GoogleAuthProvider());
			return toAuthUser(result.user);
		},
		async signOut() {
			const { auth, signOut } = await sdk;
			await signOut(auth);
		},
		async getToken(forceRefresh = false) {
			const { auth } = await sdk;
			// Si aún no se resolvió la sesión guardada, espera a que el SDK la restaure.
			await auth.authStateReady();
			return (await auth.currentUser?.getIdToken(forceRefresh)) ?? null;
		}
	};
}
