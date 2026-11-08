<script lang="ts">
	import BottomNav from '$lib/components/BottomNav.svelte';
	import MainAppBar from '$lib/components/MainAppBar.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<!-- Navigation de manwe: cabecera fija, contenido desplazable y barra inferior. -->
<div class="tabs">
	<header><MainAppBar /></header>
	<main>{@render children()}</main>
	<BottomNav />
</div>

<style>
	.tabs {
		display: flex;
		flex-direction: column;
		height: 100dvh;
	}
	header {
		flex: none;
	}
	main {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		overscroll-behavior-y: contain;
	}
</style>
