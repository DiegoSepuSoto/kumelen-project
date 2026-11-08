<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import AppBar from '$lib/components/AppBar.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import ContactButton from '$lib/components/ContactButton.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingCube from '$lib/components/LoadingCube.svelte';
	import Markdown from '$lib/components/Markdown.svelte';
	import ZoomableImage from '$lib/components/ZoomableImage.svelte';
	import { resource } from '$lib/resource.svelte';

	const service = resource((signal) => api.service(page.params.id!, signal));

	const tabs = ['Descripción', 'Profesionales', 'Contacto'];
	let selected = $state(0);

	function onKeydown(event: KeyboardEvent) {
		const last = tabs.length - 1;
		let next = selected;
		if (event.key === 'ArrowRight') next = selected === last ? 0 : selected + 1;
		else if (event.key === 'ArrowLeft') next = selected === 0 ? last : selected - 1;
		else if (event.key === 'Home') next = 0;
		else if (event.key === 'End') next = last;
		else return;

		event.preventDefault();
		selected = next;
		document.getElementById(`tab-${next}`)?.focus();
	}
</script>

<svelte:head><title>{service.data?.name ?? 'Servicio'} · Kümelen</title></svelte:head>

{#if service.status === 'loading'}
	<AppBar>{#snippet title()}{/snippet}</AppBar>
	<LoadingCube />
{:else if service.status === 'success' && service.data}
	{@const data = service.data}

	<AppBar>
		{#snippet title()}{data.name}{/snippet}
		{#snippet bottom()}
			<div class="tablist" role="tablist" aria-label="Información del servicio" tabindex="-1" onkeydown={onKeydown}>
				{#each tabs as tab, index}
					<button
						type="button"
						role="tab"
						id="tab-{index}"
						class="tab"
						class:selected={selected === index}
						aria-selected={selected === index}
						aria-controls="panel-{index}"
						tabindex={selected === index ? 0 : -1}
						onclick={() => (selected = index)}
					>
						{tab}
					</button>
				{/each}
			</div>
		{/snippet}
	</AppBar>

	<main>
		<div role="tabpanel" id="panel-0" aria-labelledby="tab-0" hidden={selected !== 0} class="panel description">
			<div class="logo">
				<ZoomableImage src={data.logo_url} alt="Logo de {data.name}">
					<Avatar src={data.logo_url} size={150} alt="Logo de {data.name}" />
				</ZoomableImage>
			</div>
			<div class="text"><Markdown source={data.description} /></div>
		</div>

		<div role="tabpanel" id="panel-1" aria-labelledby="tab-1" hidden={selected !== 1} class="panel">
			<ul class="persons">
				{#each data.persons as person (person.id)}
					<li class="person">
						<Avatar src={person.profile_picture_url} size={80} />
						<div class="person-text">
							<strong>{person.name}</strong>
							<span class="charge">{person.charge}</span>
							<span class="email">{person.email}</span>
						</div>
					</li>
				{/each}
			</ul>
		</div>

		<div role="tabpanel" id="panel-2" aria-labelledby="tab-2" hidden={selected !== 2} class="panel contacts">
			<ul>
				{#each data.contacts as contact (contact.id)}
					<li><ContactButton {contact} /></li>
				{/each}
			</ul>
		</div>
	</main>
{:else}
	<AppBar>{#snippet title()}{/snippet}</AppBar>
	<ErrorState />
{/if}

<style>
	.tablist {
		display: flex;
	}
	.tab {
		position: relative;
		flex: 1;
		min-height: 48px;
		padding: 0 16px;
		border: 0;
		background: transparent;
		color: rgb(255 255 255 / 0.75);
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}
	.tab.selected {
		color: var(--white);
	}
	.tab.selected::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 2px;
		background: var(--white);
	}
	.panel[hidden] {
		display: none;
	}
	.description {
		padding-top: var(--pad);
		padding-bottom: calc(var(--pad) * 2 + var(--safe-bottom));
	}
	.logo {
		display: flex;
		justify-content: center;
	}
	.text {
		padding: var(--pad);
	}
	.person {
		display: flex;
		align-items: center;
		gap: var(--pad);
		padding: var(--pad);
	}
	.persons li:not(:last-child)::after {
		content: '';
		display: block;
		height: 16px;
		background: linear-gradient(#000, #000) center / 100% 1px no-repeat;
	}
	.person-text {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}
	.person-text strong {
		font-size: 20px;
		font-weight: 700;
	}
	.charge {
		font-size: 18px;
	}
	.email {
		font-size: 16px;
		font-style: italic;
		overflow-wrap: anywhere;
	}
	.contacts {
		padding: var(--pad) 0 calc(var(--safe-bottom));
	}
	.contacts ul {
		display: flex;
		flex-direction: column;
		align-items: center;
	}
	.contacts li {
		margin-bottom: var(--pad);
	}
</style>
