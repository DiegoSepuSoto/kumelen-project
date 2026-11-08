type Status = 'loading' | 'success' | 'error';

/**
 * Carga datos en el navegador y reacciona a los cambios de parámetros. Equivale a los cubits de
 * manwe (loading / success / error). Debe llamarse durante la inicialización de un componente.
 */
export function resource<T>(load: (signal: AbortSignal) => Promise<T>) {
	let status = $state<Status>('loading');
	let data = $state.raw<T | undefined>(undefined);
	let error = $state.raw<unknown>(undefined);
	let version = $state(0);

	$effect(() => {
		// Dependencia explícita para poder reintentar.
		version;
		const controller = new AbortController();
		status = 'loading';

		load(controller.signal).then(
			(value) => {
				if (controller.signal.aborted) return;
				data = value;
				error = undefined;
				status = 'success';
			},
			(reason) => {
				if (controller.signal.aborted) return;
				error = reason;
				status = 'error';
			}
		);

		return () => controller.abort();
	});

	return {
		get status() {
			return status;
		},
		get data() {
			return data;
		},
		get error() {
			return error;
		},
		reload() {
			version += 1;
		}
	};
}
