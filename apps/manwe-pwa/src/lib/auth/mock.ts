import type { AuthProvider, AuthUser } from './types';

const STORAGE_KEY = 'manwe:mock-user';

const mockUser: AuthUser = {
	uid: 'mock-user',
	name: 'Estudiante UTEM',
	email: 'estudiante@example.com'
};

function read(): AuthUser | null {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		return raw ? (JSON.parse(raw) as AuthUser) : null;
	} catch {
		return null;
	}
}

/** Proveedor para desarrollo y pruebas: no contacta a ningún servicio externo. */
export function createMockAuth(): AuthProvider {
	const listeners = new Set<(user: AuthUser | null) => void>();
	const notify = (user: AuthUser | null) => listeners.forEach((listener) => listener(user));

	return {
		subscribe(onChange) {
			listeners.add(onChange);
			onChange(read());
			return () => listeners.delete(onChange);
		},
		async signIn() {
			try {
				localStorage.setItem(STORAGE_KEY, JSON.stringify(mockUser));
			} catch {
				// Sin almacenamiento la sesión dura lo que dure la página.
			}
			notify(mockUser);
			return mockUser;
		},
		async signOut() {
			try {
				localStorage.removeItem(STORAGE_KEY);
			} catch {
				// Nada que limpiar.
			}
			notify(null);
		},
		async getToken() {
			return read() ? 'mock-token' : null;
		}
	};
}
