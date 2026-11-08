<script lang="ts">
	import { api } from '$lib/api';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingCube from '$lib/components/LoadingCube.svelte';
	import { resource } from '$lib/resource.svelte';

	const catalog = resource((signal) => api.catalog(signal));
</script>

<svelte:head><title>Buscar · Kümelen</title></svelte:head>

<h1 class="sr-only">Buscar</h1>

{#if catalog.status === 'loading'}
	<LoadingCube />
{:else if catalog.status === 'success' && catalog.data}
	{@const { services, categories } = catalog.data}

	<!-- SearchButton -->
	<a class="search" href="/busqueda">
		<span class="search-bg" aria-hidden="true"></span>
		<span class="search-text">Buscar por término</span>
	</a>

	<!-- ServiceCards -->
	<section aria-label="Servicios estudiantiles">
		<ul class="services">
			{#each services as service (service.id)}
				<li>
					<a class="service card-shadow" href="/servicios/{service.id}/publicaciones">
						<span class="service-image">
							<img src={service.logo_url} alt="" loading="lazy" decoding="async" />
						</span>
						<span class="service-name">{service.name}</span>
					</a>
				</li>
			{/each}
		</ul>
	</section>

	<!-- CategoryList -->
	<section aria-label="Categorías">
		<ul class="categories">
			{#each categories as category (category.id)}
				<li><a href="/categorias/{category.id}">{category.name}</a></li>
			{/each}
		</ul>
	</section>
{:else}
	<ErrorState />
{/if}

<style>
	.search {
		position: relative;
		display: block;
		height: calc(15 * var(--vu));
		margin: var(--pad) var(--pad) 0;
		overflow: hidden;
		border-radius: 10px;
		box-shadow: 0 5px 3px var(--shadow);
	}
	.search-bg {
		position: absolute;
		inset: -12px;
		background: url('/images/search_button.jpg') center / cover;
		filter: blur(5px);
	}
	.search-text {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: var(--white);
		font-size: 30px;
		font-weight: 700;
		text-align: center;
		/* Mantiene el contraste del texto sobre la foto. */
		text-shadow: 0 1px 6px rgb(0 0 0 / 0.55);
	}

	.services {
		display: flex;
		height: calc(17 * var(--vu));
		margin: var(--pad) 0 0 var(--pad);
		padding-right: var(--pad);
		overflow-x: auto;
		scrollbar-width: none;
	}
	.services::-webkit-scrollbar {
		display: none;
	}
	.service {
		display: flex;
		flex: none;
		flex-direction: column;
		width: calc(65 * var(--wu));
		height: calc(17 * var(--vu));
		margin-right: var(--pad);
	}
	.service-image {
		height: calc(10 * var(--vu));
		overflow: hidden;
		border-radius: 10px 10px 0 0;
		background: var(--white) url('/images/loading.gif') center / 40px no-repeat;
	}
	.service-image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.service-name {
		height: calc(6 * var(--vu));
		padding: 10px var(--pad) 0;
		border-radius: 0 0 10px 10px;
		background: var(--primary);
		color: var(--white);
		font-size: 17px;
		font-weight: 700;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.categories {
		height: calc(30 * var(--vu));
		margin: var(--pad) var(--pad) var(--pad);
		overflow-y: auto;
		border-radius: 10px;
		background: var(--white);
		box-shadow: 3px 3px 4px var(--shadow);
	}
	.categories li:not(:last-child)::after {
		content: '';
		display: block;
		height: 16px;
		background: linear-gradient(#000, #000) center / 100% 1px no-repeat;
	}
	.categories a {
		display: block;
		margin: 10px 0 10px var(--pad);
		font-size: 18px;
		font-weight: 700;
	}
</style>
