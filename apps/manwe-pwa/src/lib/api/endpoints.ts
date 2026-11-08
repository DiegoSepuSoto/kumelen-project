import type { ApiClient } from './client';
import type {
	CatalogScreen,
	HomeScreen,
	NotificationsScreen,
	PostScreen,
	PostsList,
	PostsSearchResult,
	ServiceScreen
} from './types';

/** Una función por cada endpoint que manwe consume del BFF. */
export function createEndpoints(client: ApiClient) {
	return {
		home: (signal?: AbortSignal) => client.get<HomeScreen>('/v1/home', signal),
		catalog: (signal?: AbortSignal) => client.get<CatalogScreen>('/v1/catalog', signal),
		notifications: (signal?: AbortSignal) =>
			client.get<NotificationsScreen>('/v1/notification', signal),
		post: (id: string, signal?: AbortSignal) =>
			client.get<PostScreen>(`/v1/post/${encodeURIComponent(id)}`, signal),
		searchPosts: (term: string, signal?: AbortSignal) =>
			client.get<PostsSearchResult>(
				`/v1/post/search/${encodeURIComponent(term.trim().toLowerCase())}`,
				signal
			),
		service: (id: string, signal?: AbortSignal) =>
			client.get<ServiceScreen>(`/v1/service/${encodeURIComponent(id)}`, signal),
		servicePosts: (id: string, signal?: AbortSignal) =>
			client.get<PostsList>(`/v1/service/${encodeURIComponent(id)}/posts`, signal),
		categoryPosts: (id: string, signal?: AbortSignal) =>
			client.get<PostsList>(`/v1/category/${encodeURIComponent(id)}/posts`, signal)
	};
}
