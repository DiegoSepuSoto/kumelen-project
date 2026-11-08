<script lang="ts">
	import '../app.css';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import LoadingCube from '$lib/components/LoadingCube.svelte';
	import { getShowOnboarding } from '$lib/prefs';
	import { session } from '$lib/session.svelte';
	import { pwaInfo } from 'virtual:pwa-info';
	import { onMount } from 'svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();

	/** Pantallas que no exigen sesión. */
	const publicPaths = ['/login', '/bienvenida'];
	const isPublic = $derived(publicPaths.includes(page.url.pathname));

	onMount(() => {
		const stop = session.start();

		if (pwaInfo) {
			import('virtual:pwa-register').then(({ registerSW }) => registerSW({ immediate: true }));
		}

		return stop;
	});

	// Misma lógica de arranque que ManweApp: sin sesión, onboarding la primera vez y luego login.
	$effect(() => {
		if (!session.ready) return;

		if (!session.user && !isPublic) {
			goto(getShowOnboarding() ? '/bienvenida' : '/login', { replaceState: true });
		} else if (session.user && isPublic) {
			goto('/', { replaceState: true });
		}
	});

	const blocked = $derived(!session.ready || (!session.user && !isPublic) || (!!session.user && isPublic));
</script>

<svelte:head>
	{#if pwaInfo}{@html pwaInfo.webManifest.linkTag}{/if}
</svelte:head>

<a class="skip-link" href="#contenido">Saltar al contenido</a>

<div class="shell" id="contenido" tabindex="-1">
	{#if blocked}
		<LoadingCube />
	{:else}
		{@render children()}
	{/if}
</div>

<style>
	.shell {
		position: relative;
		display: flex;
		flex-direction: column;
		width: 100%;
		max-width: var(--app-max);
		min-height: 100dvh;
		margin: 0 auto;
		background: var(--bg);
		box-shadow: 0 0 24px rgb(0 0 0 / 0.12);
		outline: none;
	}
	.skip-link {
		position: absolute;
		top: -100px;
		left: 8px;
		z-index: 100;
		padding: 8px 12px;
		border-radius: 4px;
		background: var(--white);
		color: var(--text);
		font-weight: 700;
	}
	.skip-link:focus {
		top: 8px;
	}
</style>
