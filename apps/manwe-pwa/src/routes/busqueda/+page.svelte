<script lang="ts">
	import { api } from '$lib/api';
	import type { PostPreview } from '$lib/api/types';
	import AppBar from '$lib/components/AppBar.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingCube from '$lib/components/LoadingCube.svelte';
	import PostList from '$lib/components/PostList.svelte';

	type State =
		| { kind: 'initial' }
		| { kind: 'loading' }
		| { kind: 'success'; posts: PostPreview[] }
		| { kind: 'error' };

	let term = $state('');
	let view = $state.raw<State>({ kind: 'initial' });
	let controller: AbortController | undefined;

	async function search(event: SubmitEvent) {
		event.preventDefault();
		if (!term.trim()) return;

		controller?.abort();
		controller = new AbortController();
		const { signal } = controller;

		view = { kind: 'loading' };
		try {
			const result = await api.searchPosts(term, signal);
			if (!signal.aborted) view = { kind: 'success', posts: result.posts };
		} catch {
			if (!signal.aborted) view = { kind: 'error' };
		}
	}
</script>

<svelte:head><title>Buscar por término · Kümelen</title></svelte:head>

<AppBar backFallback="/catalogo">
	{#snippet title()}
		<form onsubmit={search} role="search">
			<label class="sr-only" for="term">Término que deseas buscar</label>
			<input
				id="term"
				type="search"
				enterkeyhint="search"
				autocomplete="off"
				placeholder="Término que deseas buscar"
				bind:value={term}
			/>
		</form>
	{/snippet}
</AppBar>

<main>
	<h1 class="sr-only">Buscar por término</h1>
	<div aria-live="polite">
		{#if view.kind === 'initial'}
			<p class="hint">Ingresa un término para comenzar la búsqueda</p>
		{:else if view.kind === 'loading'}
			<LoadingCube />
		{:else if view.kind === 'success'}
			{#if view.posts.length > 0}
				<PostList posts={view.posts} />
			{:else}
				<EmptyState />
			{/if}
		{:else}
			<ErrorState />
		{/if}
	</div>
</main>

<style>
	form {
		width: 100%;
	}
	input {
		width: 100%;
		height: 48px;
		padding: 0;
		border: 0;
		outline: none;
		background: transparent;
		color: var(--white);
		font: inherit;
		font-size: 18px;
		font-weight: 700;
		caret-color: var(--white);
	}
	input::placeholder {
		color: var(--white);
		font-weight: 400;
		opacity: 1;
	}
	input::-webkit-search-cancel-button {
		filter: brightness(0) invert(1);
	}
	/* El foco se indica con una línea bajo el texto, ya que la barra no tiene recuadro. */
	form:focus-within {
		box-shadow: 0 2px 0 var(--white);
	}
	.hint {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 60dvh;
		padding: 0 var(--pad);
		font-size: 22px;
		text-align: center;
	}
</style>
