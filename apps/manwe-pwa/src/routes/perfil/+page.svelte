<script lang="ts">
	import { goto } from '$app/navigation';
	import AppBar from '$lib/components/AppBar.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import { appVersion } from '$lib/config';
	import { session } from '$lib/session.svelte';

	async function signOut() {
		await session.signOut();
		goto('/login', { replaceState: true });
	}
</script>

<svelte:head><title>Tu perfil · Kümelen</title></svelte:head>

<AppBar>
	{#snippet title()}Tu Perfil{/snippet}
</AppBar>

<main class="profile">
	<h1 class="sr-only">Tu perfil</h1>
	<Avatar src="/images/profile.png" size={112} ring={6} alt="" />

	<dl>
		<dt>Tu nombre:</dt>
		<dd>{session.user?.name || '—'}</dd>
		<dt>Tu correo:</dt>
		<dd>{session.user?.email || '—'}</dd>
	</dl>

	<p class="version">Versión de la app: {appVersion}</p>

	<button type="button" class="btn logout" onclick={signOut}>
		<Icon name="logout" size={30} />
		<span>Cerrar Sesión</span>
	</button>
</main>

<style>
	.profile {
		display: flex;
		flex: 1;
		flex-direction: column;
		align-items: flex-start;
		padding: calc(var(--pad) * 2);
		padding-bottom: calc(var(--pad) * 2 + var(--safe-bottom));
	}
	dl {
		margin: 20px 0 0;
	}
	dt {
		font-size: 20px;
		font-weight: 700;
	}
	dd {
		margin: 0 0 15px;
		font-size: 17px;
		overflow-wrap: anywhere;
	}
	.version {
		margin-top: auto;
		font-size: 20px;
		font-weight: 700;
	}
	.logout {
		gap: 8px;
		margin-top: 25px;
		padding: 7px 16px;
		background: var(--danger);
		color: var(--text);
		font-size: 20px;
	}
	.logout :global(svg) {
		color: var(--white);
	}
	.logout span {
		color: var(--white);
	}
</style>
