<script lang="ts">
	import { api } from '$lib/api';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingCube from '$lib/components/LoadingCube.svelte';
	import PostsSection from '$lib/components/PostsSection.svelte';
	import { resource } from '$lib/resource.svelte';

	const home = resource((signal) => api.home(signal));
</script>

<svelte:head><title>Publicaciones · Kümelen</title></svelte:head>

<h1 class="sr-only">Publicaciones</h1>

{#if home.status === 'loading'}
	<LoadingCube />
{:else if home.status === 'success' && home.data}
	{#each home.data.services_with_posts as service (service.id)}
		<PostsSection id={service.id} abbreviation={service.abbreviation} posts={service.posts_for_preview} />
	{/each}
{:else}
	<ErrorState />
{/if}
