export interface AuthUser {
	uid: string;
	name: string;
	email: string;
}

/**
 * Contrato entre la app y el proveedor de identidad. El backend solo verifica el ID token
 * (D5), así que cambiar de proveedor no afecta a las pantallas.
 */
export interface AuthProvider {
	/** Resuelve la sesión guardada y avisa de cada cambio; devuelve la función para dejar de escuchar. */
	subscribe(onChange: (user: AuthUser | null) => void): () => void;
	signIn(): Promise<AuthUser>;
	signOut(): Promise<void>;
	/** Token que se envía al BFF; `forceRefresh` lo renueva tras un 401. */
	getToken(forceRefresh?: boolean): Promise<string | null>;
}
