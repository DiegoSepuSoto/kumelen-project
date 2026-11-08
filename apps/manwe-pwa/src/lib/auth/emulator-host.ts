/**
 * Deja un `host:puerto` limpio: acepta que venga con `http://` o con una barra final. Una cadena vacía
 * (o solo espacios) significa "sin emulador".
 */
export function normalizeEmulatorHost(value: string): string {
	return value
		.trim()
		.replace(/^https?:\/\//i, '')
		.replace(/\/+$/, '');
}
