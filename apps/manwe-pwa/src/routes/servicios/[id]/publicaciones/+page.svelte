<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import AppBar from '$lib/components/AppBar.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import LoadingCube from '$lib/components/LoadingCube.svelte';
	import PostList from '$lib/components/PostList.svelte';
	import { resource } from '$lib/resource.svelte';

	const servicePosts = resource((signal) => api.servicePosts(page.params.id!, signal));
</script>

<svelte:head><title>{servicePosts.data?.name ?? 'Publicaciones'} · Kümelen</title></svelte:head>

{#if servicePosts.status === 'loading'}
	<AppBar backFallback="/catalogo">{#snippet title()}{/snippet}</AppBar>
	<LoadingCube />
{:else if servicePosts.status === 'success' && servicePosts.data}
	{@const data = servicePosts.data}

	<AppBar backFallback="/catalogo">
		{#snippet title()}{data.name}{/snippet}
		{#snippet actions()}
			<a class="info" href="/servicios/{data.id}" aria-label="Información de {data.name}">
				<Icon name="info" />
			</a>
		{/snippet}
	</AppBar>

	<main>
		<h1 class="sr-only">{data.name}</h1>
		<PostList posts={data.posts_for_preview} />
	</main>
{:else}
	<AppBar backFallback="/catalogo">{#snippet title()}{/snippet}</AppBar>
	<ErrorState />
{/if}

<style>
	.info {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 48px;
		height: 48px;
	}
	main {
		padding-bottom: var(--safe-bottom);
	}
</style>
