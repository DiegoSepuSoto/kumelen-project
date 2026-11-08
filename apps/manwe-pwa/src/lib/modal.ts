/** Acción de Svelte: abre el <dialog> como modal al montarse (con foco atrapado y Escape para cerrar). */
export function modal(node: HTMLDialogElement) {
	if (!node.open) node.showModal();
	return {
		destroy() {
			if (node.open) node.close();
		}
	};
}
