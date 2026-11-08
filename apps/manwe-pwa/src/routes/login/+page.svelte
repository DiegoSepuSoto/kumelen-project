<script lang="ts">
	import Icon from '$lib/components/Icon.svelte';
	import { termsUrl } from '$lib/config';
	import { session } from '$lib/session.svelte';

	let submitting = $state(false);
	let failed = $state(false);
	let toastTimer: ReturnType<typeof setTimeout> | undefined;

	async function signIn() {
		submitting = true;
		failed = false;
		try {
			await session.signIn();
			// El layout raíz redirige al inicio cuando la sesión queda iniciada.
		} catch {
			failed = true;
			clearTimeout(toastTimer);
			toastTimer = setTimeout(() => (failed = false), 4000);
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head><title>Ingresar · Kümelen</title></svelte:head>

<div class="safe-top"></div>

<main class="login">
	<img class="utem" src="/images/marca-utem-horizontal.png" alt="Universidad Tecnológica Metropolitana del Estado de Chile" />

	<header class="header">
		<h1>App Kümelen</h1>
		<p>Plataforma de apoyo extraacadémico</p>
	</header>

	<section class="form" aria-label="Ingreso">
		<p class="lead">Ingresa con tu cuenta de Google</p>

		<button type="button" class="btn sign-in" onclick={signIn} disabled={submitting}>
			{#if submitting}
				<span class="spinner" role="status"><span class="sr-only">Ingresando…</span></span>
			{:else}
				<Icon name="google" size={22} />
				Ingresar con Google
			{/if}
		</button>

		<a class="terms" href={termsUrl} target="_blank" rel="noopener noreferrer">Ver términos y condiciones</a>
	</section>
</main>

{#if failed}
	<div class="toast" role="alert">No se pudo ingresar a la aplicación</div>
{/if}

<style>
	.safe-top {
		flex: none;
		height: var(--safe-top);
		background: var(--primary);
	}
	.login {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-height: calc(85 * var(--vu));
	}
	.utem {
		width: calc(80 * var(--wu));
		margin: calc(var(--pad) * 2) auto 0;
	}
	.header {
		display: flex;
		flex-direction: column;
		justify-content: center;
		height: calc(20 * var(--vu));
		min-height: 120px;
		margin: 0 calc(var(--pad) * 2);
	}
	h1 {
		font-size: 42px;
		font-weight: 700;
	}
	.header p {
		font-size: 19px;
	}
	.form {
		display: flex;
		flex: 1;
		flex-direction: column;
		align-items: center;
		justify-content: flex-start;
		gap: calc(5 * var(--vu));
		min-height: calc(45 * var(--vu));
		padding-top: calc(2 * var(--vu));
		margin: 0 calc(var(--pad) * 2) calc(var(--pad) * 2);
		padding-bottom: var(--safe-bottom);
		text-align: center;
	}
	.lead {
		font-size: 18px;
		font-style: italic;
	}
	.terms {
		color: var(--primary);
		font-size: 17px;
		font-weight: 700;
	}
	.sign-in {
		gap: 10px;
		min-width: 230px;
		min-height: 52px;
		padding: 7px 20px;
		font-size: 20px;
		font-weight: 600;
		background: var(--white);
		color: var(--text);
		border: 1.5px solid var(--primary);
	}
	.spinner {
		width: 28px;
		height: 28px;
		border: 3px solid var(--primary);
		border-right-color: transparent;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
	.toast {
		position: fixed;
		bottom: calc(16px + var(--safe-bottom));
		left: 50%;
		z-index: 20;
		width: min(calc(100vw - 32px), 448px);
		padding: 14px 16px;
		transform: translateX(-50%);
		border-radius: 4px;
		background: #323232;
		color: var(--white);
		font-size: 14px;
	}
</style>
