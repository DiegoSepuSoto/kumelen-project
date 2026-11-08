export class ApiError extends Error {
	constructor(
		readonly status: number,
		message: string
	) {
		super(message);
		this.name = 'ApiError';
	}
}

export interface ApiClientDeps {
	base: string;
	/** Token vigente; con `forceRefresh` el proveedor lo renueva (equivale a refresh-token en manwe). */
	getToken: (forceRefresh?: boolean) => Promise<string | null>;
	/** Se llama cuando el servidor rechaza incluso el token renovado: la sesión ya no sirve. */
	onUnauthorized: () => void | Promise<void>;
	fetch?: typeof fetch;
}

export interface ApiClient {
	get<T>(path: string, signal?: AbortSignal): Promise<T>;
}

/**
 * Cliente del BFF. Reproduce el interceptor de manwe: ante un 401 renueva el token y repite la
 * petición una vez; si vuelve a fallar, cierra la sesión.
 */
export function createApiClient(deps: ApiClientDeps): ApiClient {
	const doFetch = deps.fetch ?? fetch;

	async function send(path: string, signal: AbortSignal | undefined, forceRefresh: boolean) {
		const token = await deps.getToken(forceRefresh);
		return doFetch(`${deps.base}${path}`, {
			signal,
			headers: {
				Accept: 'application/json',
				...(token ? { Authorization: `Bearer ${token}` } : {})
			}
		});
	}

	return {
		async get<T>(path: string, signal?: AbortSignal): Promise<T> {
			let response = await send(path, signal, false);

			if (response.status === 401) {
				response = await send(path, signal, true);
				if (response.status === 401) {
					await deps.onUnauthorized();
					throw new ApiError(401, 'No autorizado');
				}
			}

			if (!response.ok) {
				throw new ApiError(response.status, `Error ${response.status} al consultar ${path}`);
			}

			return (await response.json()) as T;
		}
	};
}
