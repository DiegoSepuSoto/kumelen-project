<script lang="ts">
	import { modal } from '$lib/modal';

	/** Equivale a showAlertDialogToOpenLink de manwe. */
	let { url, onclose }: { url: string; onclose: () => void } = $props();

	function proceed() {
		window.open(url, '_blank', 'noopener,noreferrer');
		onclose();
	}
</script>

<dialog use:modal {onclose} aria-labelledby="link-dialog-title" aria-describedby="link-dialog-text">
	<h2 id="link-dialog-title">Abrir enlace</h2>
	<p id="link-dialog-text">Serás redirigido a: {url}</p>
	<div class="actions">
		<button type="button" class="btn" onclick={onclose}>Cancelar</button>
		<button type="button" class="btn" onclick={proceed}>Continuar</button>
	</div>
</dialog>

<style>
	dialog {
		width: min(calc(100vw - 80px), 400px);
		padding: 24px 24px 8px;
		border: 0;
		border-radius: 4px;
		color: var(--text);
		box-shadow: 0 11px 15px -7px rgb(0 0 0 / 0.2), 0 24px 38px 3px rgb(0 0 0 / 0.14);
	}
	dialog::backdrop {
		background: rgb(0 0 0 / 0.54);
	}
	h2 {
		margin-bottom: 20px;
		font-size: 20px;
		font-weight: 600;
	}
	p {
		font-size: 16px;
		overflow-wrap: anywhere;
	}
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		padding: 16px 0 8px;
	}
</style>
