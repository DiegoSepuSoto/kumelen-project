<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Snippet } from 'svelte';
	import Icon from './Icon.svelte';

	/** AppBar de Material en verde primario, con flecha para volver, acciones y pestañas opcionales. */
	let {
		title,
		actions,
		bottom,
		backFallback = '/'
	}: {
		title: Snippet;
		actions?: Snippet;
		bottom?: Snippet;
		/** Destino si no hay historial al que volver (por ejemplo, al abrir un enlace directo). */
		backFallback?: string;
	} = $props();

	function goBack() {
		if (history.length > 1) history.back();
		else goto(backFallback, { replaceState: true });
	}
</script>

<header class="bar">
	<div class="row">
		<button type="button" class="icon-button" onclick={goBack} aria-label="Volver">
			<Icon name="back" />
		</button>
		<div class="title">{@render title()}</div>
		{#if actions}<div class="actions">{@render actions()}</div>{/if}
	</div>
	{#if bottom}{@render bottom()}{/if}
</header>

<style>
	.bar {
		position: sticky;
		top: 0;
		z-index: 5;
		padding-top: var(--safe-top);
		background: var(--primary);
		color: var(--white);
		box-shadow:
			0 2px 4px -1px rgb(0 0 0 / 0.2),
			0 4px 5px 0 rgb(0 0 0 / 0.14),
			0 1px 10px 0 rgb(0 0 0 / 0.12);
	}
	.row {
		display: flex;
		align-items: center;
		height: 56px;
		padding: 0 4px;
	}
	.title {
		flex: 1;
		min-width: 0;
		padding-left: 16px;
		font-size: 20px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.actions {
		display: flex;
		align-items: center;
		padding-right: 8px;
	}
	.icon-button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 48px;
		height: 48px;
		padding: 0;
		border: 0;
		border-radius: 50%;
		background: transparent;
		color: inherit;
		cursor: pointer;
	}
	.bar :global(:focus-visible) {
		outline-color: var(--white);
	}
</style>
