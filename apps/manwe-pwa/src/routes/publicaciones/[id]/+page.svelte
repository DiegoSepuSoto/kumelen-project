<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import AppBar from '$lib/components/AppBar.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import LoadingCube from '$lib/components/LoadingCube.svelte';
	import Markdown from '$lib/components/Markdown.svelte';
	import ZoomableImage from '$lib/components/ZoomableImage.svelte';
	import { resource } from '$lib/resource.svelte';

	const post = resource((signal) => api.post(page.params.id!, signal));

	function goBack() {
		if (history.length > 1) history.back();
		else goto('/', { replaceState: true });
	}
</script>

<svelte:head><title>{post.data?.title ?? 'Publicación'} · Kümelen</title></svelte:head>

{#if post.status === 'loading'}
	<AppBar>{#snippet title()}{/snippet}</AppBar>
	<LoadingCube />
{:else if post.status === 'success' && post.data}
	{@const data = post.data}

	<main class="post">
		<div class="hero">
			<ZoomableImage src={data.image} alt={data.title} class="hero-image">
				<img src={data.image} alt={data.title} />
			</ZoomableImage>
			<div class="safe-strip"></div>
			<div class="buttons">
				<button type="button" class="round" onclick={goBack} aria-label="Volver">
					<Icon name="backIos" />
				</button>
				<a class="round" href="/servicios/{data.service_id}" aria-label="Información de {data.service_name}">
					<Icon name="info" />
				</a>
			</div>
		</div>

		<p class="service">{data.service_name}</p>

		{#if data.categories.length > 0}
			<ul class="chips" aria-label="Categorías">
				{#each data.categories as category (category.id)}
					<li>{category.name}</li>
				{/each}
			</ul>
		{/if}

		<h1>{data.title}</h1>

		<article class="body"><Markdown source={data.body} /></article>
	</main>
{:else}
	<AppBar>{#snippet title()}{/snippet}</AppBar>
	<ErrorState />
{/if}

<style>
	.post {
		padding-bottom: calc(var(--pad) + var(--safe-bottom));
	}
	.hero {
		position: relative;
		height: 270px;
	}
	.hero :global(.hero-image),
	.hero img {
		width: 100%;
		height: 100%;
	}
	.hero img {
		object-fit: cover;
	}
	.safe-strip {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: var(--safe-top);
		background: var(--primary);
		pointer-events: none;
	}
	.buttons {
		position: absolute;
		top: max(60px, calc(var(--safe-top) + 16px));
		left: 0;
		right: 0;
		display: flex;
		justify-content: space-between;
		padding: 0 var(--pad);
		pointer-events: none;
	}
	.round {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 50px;
		height: 50px;
		padding: 0;
		border: 0;
		border-radius: 50%;
		background: var(--primary);
		color: var(--white);
		cursor: pointer;
		pointer-events: auto;
	}
	.service {
		padding-top: var(--pad);
		text-align: center;
		font-size: 24px;
		font-weight: 700;
		color: var(--primary);
	}
	.chips {
		display: flex;
		height: 45px;
		padding: var(--pad) var(--pad) 0 0;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.chips::-webkit-scrollbar {
		display: none;
	}
	.chips li {
		flex: none;
		margin-left: var(--pad);
		padding: 3px 15px;
		border-radius: 5px;
		background: var(--primary);
		color: var(--white);
		font-weight: 700;
		white-space: nowrap;
	}
	h1 {
		margin: 0 var(--pad);
		padding-top: var(--pad);
		text-align: center;
		font-size: 24px;
		font-weight: 700;
	}
	.body {
		padding: var(--pad);
	}
</style>
