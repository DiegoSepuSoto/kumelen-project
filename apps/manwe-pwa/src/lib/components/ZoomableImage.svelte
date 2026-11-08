<script lang="ts">
	import { modal } from '$lib/modal';
	import type { Snippet } from 'svelte';
	import Icon from './Icon.svelte';

	/** Envuelve una imagen para verla a tamaño completo (FullSizeImageScreen de manwe). */
	let {
		src,
		alt,
		children,
		class: className = ''
	}: { src: string; alt: string; children: Snippet; class?: string } = $props();

	let open = $state(false);
</script>

<button type="button" class="trigger {className}" onclick={() => (open = true)} aria-label="Ver imagen a tamaño completo: {alt}">
	{@render children()}
</button>

{#if open}
	<dialog use:modal onclose={() => (open = false)} aria-label={alt}>
		<button type="button" class="close" onclick={() => (open = false)} aria-label="Cerrar imagen">
			<Icon name="close" />
		</button>
		<!-- Un clic sobre la imagen la cierra, como el GestureDetector de manwe. -->
		<button type="button" class="full" onclick={() => (open = false)} aria-label="Cerrar imagen">
			<img {src} {alt} />
		</button>
	</dialog>
{/if}

<style>
	.trigger {
		display: block;
		padding: 0;
		border: 0;
		background: none;
		cursor: zoom-in;
	}
	dialog {
		width: 100vw;
		max-width: none;
		height: 100dvh;
		max-height: none;
		padding: 0;
		border: 0;
		background: var(--bg);
	}
	.full {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		height: 100%;
		padding: 0;
		border: 0;
		background: none;
		cursor: zoom-out;
	}
	.full img {
		max-width: 100%;
		max-height: 100%;
		object-fit: contain;
	}
	.close {
		position: absolute;
		top: calc(8px + var(--safe-top));
		right: 8px;
		z-index: 1;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 48px;
		height: 48px;
		padding: 0;
		border: 0;
		border-radius: 50%;
		background: var(--primary);
		color: var(--white);
		cursor: pointer;
	}
</style>
