<script lang="ts">
	import { cmsBase } from '$lib/config';
	import { renderMarkdown } from '$lib/markdown';
	import ConfirmLinkDialog from './ConfirmLinkDialog.svelte';

	/** MarkdownBody de manwe: `post` para publicaciones y descripciones, `notification` para tarjetas. */
	let { source, variant = 'post' }: { source: string; variant?: 'post' | 'notification' } = $props();

	const html = $derived(renderMarkdown(source, cmsBase));
	let pendingLink = $state<string | null>(null);

	function interceptLinks(event: MouseEvent) {
		const link = (event.target as Element | null)?.closest('a[href]');
		if (!link) return;
		event.preventDefault();
		pendingLink = (link as HTMLAnchorElement).href;
	}
</script>

<!-- Los enlaces son <a> reales: el clic con ratón y la tecla Enter llegan a este mismo manejador. -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div class="md {variant}" onclick={interceptLinks}>{@html html}</div>

{#if pendingLink}
	<ConfirmLinkDialog url={pendingLink} onclose={() => (pendingLink = null)} />
{/if}

<style>
	.md {
		--base: 17px;
		--h1: 22px;
		overflow-wrap: anywhere;
	}
	.md.notification {
		--base: 14px;
		--h1: 15px;
	}
	.md :global(p) {
		margin: 0 0 var(--base);
		font-size: var(--base);
	}
	.md :global(h1) {
		margin: 0 0 calc(var(--base) * 0.6);
		font-size: var(--h1);
		font-weight: 700;
	}
	.md :global(h2) {
		margin: 0 0 calc(var(--base) * 0.6);
		font-size: calc(var(--h1) * 0.9);
		font-weight: 700;
	}
	.md :global(h3),
	.md :global(h4) {
		margin: 0 0 calc(var(--base) * 0.5);
		font-size: calc(var(--h1) * 0.8);
		font-weight: 700;
	}
	.md :global(ul),
	.md :global(ol) {
		margin: 0 0 var(--base);
		padding-left: 1.4em;
		font-size: var(--base);
	}
	.md :global(ul) {
		list-style: disc;
	}
	.md :global(ol) {
		list-style: decimal;
	}
	.md :global(a) {
		color: var(--secondary);
		text-decoration: underline;
	}
	.md :global(img) {
		max-width: 100%;
		height: auto;
		margin: 0 auto var(--base);
	}
	.md :global(blockquote) {
		margin: 0 0 var(--base);
		padding: 0 0 0 12px;
		border-left: 4px solid #cfd8cd;
	}
	.md :global(:last-child) {
		margin-bottom: 0;
	}
</style>
