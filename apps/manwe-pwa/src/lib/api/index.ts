import { apiBase } from '$lib/config';
import { getAuthProvider, session } from '$lib/session.svelte';
import { createApiClient } from './client';
import { createEndpoints } from './endpoints';

const client = createApiClient({
	base: apiBase,
	getToken: (forceRefresh) => getAuthProvider().getToken(forceRefresh),
	onUnauthorized: () => session.signOut()
});

export const api = createEndpoints(client);
export { ApiError } from './client';
