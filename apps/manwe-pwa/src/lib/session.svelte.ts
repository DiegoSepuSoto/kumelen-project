import { createFirebaseAuth } from '$lib/auth/firebase';
import { createMockAuth } from '$lib/auth/mock';
import type { AuthProvider, AuthUser } from '$lib/auth/types';
import { authMode } from '$lib/config';
import { clearOfflineData } from '$lib/offline';

let provider: AuthProvider | undefined;

export function getAuthProvider(): AuthProvider {
	provider ??= authMode === 'firebase' ? createFirebaseAuth() : createMockAuth();
	return provider;
}

/** Estado de sesión compartido por toda la app. */
class Session {
	user = $state<AuthUser | null>(null);
	/** `false` hasta que el proveedor informa si hay una sesión guardada. */
	ready = $state(false);

	/** Empieza a escuchar al proveedor; devuelve la función para detenerse. */
	start(): () => void {
		return getAuthProvider().subscribe((user) => {
			this.user = user;
			this.ready = true;
		});
	}

	async signIn(): Promise<void> {
		await getAuthProvider().signIn();
	}

	async signOut(): Promise<void> {
		await clearOfflineData();
		await getAuthProvider().signOut();
	}
}

export const session = new Session();
