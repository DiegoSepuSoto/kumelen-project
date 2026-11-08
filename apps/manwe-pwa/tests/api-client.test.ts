import { describe, expect, it, vi } from 'vitest';
import { ApiError, createApiClient } from '../src/lib/api/client';
import { createEndpoints } from '../src/lib/api/endpoints';

function jsonResponse(body: unknown, status = 200) {
	return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

function setup(responses: Response[], tokens: (string | null)[] = ['t1', 't2']) {
	const fetchMock = vi.fn(async () => responses.shift() ?? jsonResponse({}, 500));
	const getToken = vi.fn(async (forceRefresh?: boolean) => (forceRefresh ? tokens[1] : tokens[0]));
	const onUnauthorized = vi.fn();
	const client = createApiClient({ base: 'https://api.test', getToken, onUnauthorized, fetch: fetchMock as typeof fetch });
	return { client, fetchMock, getToken, onUnauthorized };
}

describe('createApiClient', () => {
	it('envía el token en el encabezado Authorization', async () => {
		const { client, fetchMock } = setup([jsonResponse({ ok: true })]);

		await client.get('/v1/home');

		const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
		expect(url).toBe('https://api.test/v1/home');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer t1');
	});

	it('no envía Authorization si no hay token', async () => {
		const { client, fetchMock } = setup([jsonResponse({})], [null, null]);

		await client.get('/v1/home');

		const init = fetchMock.mock.calls[0][1] as unknown as RequestInit;
		expect(init.headers).not.toHaveProperty('Authorization');
	});

	it('ante un 401 renueva el token y repite la petición una vez', async () => {
		const { client, fetchMock, getToken, onUnauthorized } = setup([
			jsonResponse({ message: 'expired' }, 401),
			jsonResponse({ services_with_posts: [] })
		]);

		const result = await client.get<{ services_with_posts: unknown[] }>('/v1/home');

		expect(result).toEqual({ services_with_posts: [] });
		expect(fetchMock).toHaveBeenCalledTimes(2);
		expect(getToken).toHaveBeenLastCalledWith(true);
		const retryInit = fetchMock.mock.calls[1][1] as unknown as RequestInit;
		expect((retryInit.headers as Record<string, string>).Authorization).toBe('Bearer t2');
		expect(onUnauthorized).not.toHaveBeenCalled();
	});

	it('si el token renovado también es rechazado, cierra la sesión', async () => {
		const { client, onUnauthorized } = setup([jsonResponse({}, 401), jsonResponse({}, 401)]);

		await expect(client.get('/v1/home')).rejects.toMatchObject({ status: 401 });
		expect(onUnauthorized).toHaveBeenCalledOnce();
	});

	it('lanza ApiError con el código cuando la respuesta no es correcta', async () => {
		const { client, onUnauthorized } = setup([jsonResponse({ message: 'boom' }, 502)]);

		const error = await client.get('/v1/home').catch((e) => e);

		expect(error).toBeInstanceOf(ApiError);
		expect(error.status).toBe(502);
		expect(onUnauthorized).not.toHaveBeenCalled();
	});
});

describe('createEndpoints', () => {
	it('usa las mismas rutas del BFF que manwe', async () => {
		const get = vi.fn(async () => ({}));
		const api = createEndpoints({ get } as never);

		await api.home();
		await api.catalog();
		await api.notifications();
		await api.post('p1');
		await api.service('s1');
		await api.servicePosts('s1');
		await api.categoryPosts('c1');

		expect(get.mock.calls.map((call) => (call as unknown[])[0])).toEqual([
			'/v1/home',
			'/v1/catalog',
			'/v1/notification',
			'/v1/post/p1',
			'/v1/service/s1',
			'/v1/service/s1/posts',
			'/v1/category/c1/posts'
		]);
	});

	it('normaliza y codifica el término de búsqueda', async () => {
		const get = vi.fn(async () => ({}));
		const api = createEndpoints({ get } as never);

		await api.searchPosts('  Hábitos de Estudio ');

		expect(get).toHaveBeenCalledWith('/v1/post/search/h%C3%A1bitos%20de%20estudio', undefined);
	});

	it('codifica los identificadores para que no alteren la ruta', async () => {
		const get = vi.fn(async () => ({}));
		const api = createEndpoints({ get } as never);

		await api.post('../admin');

		expect(get).toHaveBeenCalledWith('/v1/post/..%2Fadmin', undefined);
	});
});
