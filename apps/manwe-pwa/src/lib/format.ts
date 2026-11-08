/** Primer nombre para el saludo ("¡Hola, Diego!"). */
export function firstName(fullName: string | null | undefined): string {
	return fullName?.trim().split(/\s+/)[0] || 'Estudiante';
}

/** Antepone la base del CMS a las rutas relativas; deja intactas las URL absolutas. */
export function resolveUrl(uri: string, base: string): string {
	if (/^(https?:)?\/\//i.test(uri) || uri.startsWith('data:')) return uri;
	return `${base}${uri.startsWith('/') ? '' : '/'}${uri}`;
}
