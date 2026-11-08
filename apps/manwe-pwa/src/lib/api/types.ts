/** Modelos de las respuestas del BFF; los nombres coinciden con los de manwe (lib/src/domain/models). */

export interface PostPreview {
	id: string;
	title: string;
	summary: string;
	image_url: string;
}

export interface ServiceWithPosts {
	id: string;
	abbreviation: string;
	posts_for_preview: PostPreview[];
}

export interface HomeScreen {
	services_with_posts: ServiceWithPosts[];
}

export interface Category {
	id: string;
	name: string;
}

export interface CatalogService {
	id: string;
	name: string;
	logo_url: string;
}

export interface CatalogScreen {
	services: CatalogService[];
	categories: Category[];
}

export interface PostsList {
	id: string;
	name: string;
	posts_for_preview: PostPreview[];
}

export interface PostsSearchResult {
	posts: PostPreview[];
}

export interface PostScreen {
	id: string;
	title: string;
	image: string;
	body: string;
	service_id: string;
	service_name: string;
	categories: Category[];
}

export interface Person {
	id: string;
	name: string;
	charge: string;
	email: string;
	profile_picture_url: string;
}

export interface Contact {
	id: string;
	name: string;
	type: string;
	link: string;
}

export interface ServiceScreen {
	id: string;
	logo_url: string;
	name: string;
	description: string;
	persons: Person[];
	contacts: Contact[];
}

export interface AppNotification {
	id: string;
	title: string;
	body: string;
	service_name: string;
	service_image_url: string;
}

export interface NotificationsScreen {
	notifications: AppNotification[];
}
