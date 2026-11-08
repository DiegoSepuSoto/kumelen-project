<script lang="ts">
	import { page } from '$app/state';
	import Icon, { type IconName } from './Icon.svelte';

	const tabs: { href: string; label: string; icon: IconName }[] = [
		{ href: '/', label: 'Publicaciones', icon: 'home' },
		{ href: '/catalogo', label: 'Buscar', icon: 'search' },
		{ href: '/notificaciones', label: 'Notificaciones', icon: 'notifications' }
	];

	const isActive = (href: string) => page.url.pathname === href;
</script>

<!-- Equivale a FancyBottomNavigation: la pestaña activa sube en un círculo verde con su título. -->
<nav class="nav" aria-label="Principal">
	{#each tabs as tab}
		{@const active = isActive(tab.href)}
		<a class="tab" class:active href={tab.href} aria-current={active ? 'page' : undefined}>
			<span class="circle"><Icon name={tab.icon} size={active ? 28 : 26} /></span>
			<span class="label" class:sr-only={!active}>{tab.label}</span>
		</a>
	{/each}
</nav>

<style>
	.nav {
		position: relative;
		z-index: 2;
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		flex: none;
		height: calc(65px + var(--safe-bottom));
		padding-bottom: var(--safe-bottom);
		background: var(--white);
	}
	.tab {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--primary);
	}
	.circle {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 48px;
		height: 48px;
		border-radius: 50%;
		transition:
			transform 0.3s ease,
			background-color 0.3s ease,
			color 0.3s ease;
	}
	.active .circle {
		width: 60px;
		height: 60px;
		transform: translateY(-22px);
		background: var(--primary);
		color: var(--white);
		box-shadow:
			0 0 0 6px var(--white),
			0 2px 8px 6px rgb(0 0 0 / 0.12);
	}
	.label {
		position: absolute;
		bottom: 5px;
		left: 0;
		right: 0;
		text-align: center;
		font-size: 12px;
		font-weight: 600;
		color: var(--text);
	}
	.tab:focus-visible {
		outline-offset: -4px;
	}
</style>
