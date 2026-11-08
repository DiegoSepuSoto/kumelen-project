<script lang="ts">
	import { firstName } from '$lib/format';
	import { session } from '$lib/session.svelte';
</script>

<!-- Cabecera de las pestañas principales (MainAppBar de manwe). Lleva al perfil. -->
<a class="bar" href="/perfil">
	<span class="avatar"><img src="/images/profile.png" alt="" /></span>
	<span class="greeting">¡Hola, {firstName(session.user?.name)}!</span>
	<span class="hint">Ver Perfil</span>
</a>

<style>
	.bar {
		display: flex;
		flex: none;
		flex-direction: column;
		align-items: center;
		width: 100%;
		height: calc(25 * var(--vu));
		padding-top: calc(5 * var(--vu) + var(--safe-top));
		color: var(--white);
		font-weight: 700;
		background: linear-gradient(to right, var(--primary) 37%, var(--secondary) 95%);
		box-shadow: 0 3px 3px #9e9e9e;
		position: relative;
		z-index: 1;
	}
	@media (min-height: 751px) {
		.bar {
			height: calc(22 * var(--vu));
		}
	}
	.avatar {
		display: flex;
		width: 52px;
		height: 52px;
		padding: 4px;
		margin-bottom: 10px;
		border-radius: 50%;
		background: var(--white);
	}
	@media (min-height: 601px) {
		.avatar {
			width: 76px;
			height: 76px;
			padding: 4px;
		}
	}
	.avatar img {
		width: 100%;
		height: 100%;
		border-radius: 50%;
		background: var(--white);
	}
	.greeting {
		font-size: 18px;
	}
	.hint {
		font-size: 14px;
	}
	.bar:focus-visible {
		outline-color: var(--white);
		outline-offset: -4px;
	}
</style>
