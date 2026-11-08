/** Nombre del caché de respuestas del BFF; debe coincidir con vite.config.ts. */
export const API_CACHE = 'manwe-api';

/**
 * Borra los datos guardados para uso sin conexión. Se llama al cerrar sesión para que, en un
 * dispositivo compartido, la siguiente persona no vea contenido de la anterior.
 */
export async function clearOfflineData(): Promise<void> {
	if (typeof caches === 'undefined') return;
	try {
		await caches.delete(API_CACHE);
	} catch {
		// Un fallo al limpiar no debe impedir cerrar sesión.
	}
}
