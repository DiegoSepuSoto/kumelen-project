<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import AppBar from '$lib/components/AppBar.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingCube from '$lib/components/LoadingCube.svelte';
	import PostList from '$lib/components/PostList.svelte';
	import { resource } from '$lib/resource.svelte';

	const categoryPosts = resource((signal) => api.categoryPosts(page.params.id!, signal));
</script>

<svelte:head><title>{categoryPosts.data?.name ?? 'Categoría'} · Kümelen</title></svelte:head>

{#if categoryPosts.status === 'loading'}
	<AppBar backFallback="/catalogo">{#snippet title()}{/snippet}</AppBar>
	<LoadingCube />
{:else if categoryPosts.status === 'success' && categoryPosts.data}
	{@const data = categoryPosts.data}

	<AppBar backFallback="/catalogo">{#snippet title()}{data.name}{/snippet}</AppBar>
	<main>
		<h1 class="sr-only">{data.name}</h1>
		<PostList posts={data.posts_for_preview} />
	</main>
{:else}
	<AppBar backFallback="/catalogo">{#snippet title()}{/snippet}</AppBar>
	<ErrorState />
{/if}

<style>
	main {
		padding-bottom: var(--safe-bottom);
	}
</style>
