<script lang="ts">
	import { goto } from '$app/navigation';
	import { setShowOnboarding } from '$lib/prefs';

	let track: HTMLUListElement | undefined = $state();
	let current = $state(0);

	const slides = [
		{
			key: 'content',
			title: 'Infórmate',
			text: 'Con las publicaciones y notificaciones de los servicios estudiantiles de la universidad podrás conocer información relevante sobre tu formación extraacadémica',
			image: '/images/onboarding_content.png',
			imageFirst: true
		},
		{
			key: 'communication',
			title: 'Estamos contigo',
			text: 'Podrás contactar a los diferentes servicios estudiantiles en caso que necesites apoyo',
			image: '/images/onboarding_communication.png',
			imageFirst: false
		}
	];

	const isLast = $derived(current === slides.length - 1);

	function onScroll() {
		if (!track) return;
		current = Math.round(track.scrollLeft / track.clientWidth);
	}

	function show(index: number) {
		track?.scrollTo({ left: index * (track?.clientWidth ?? 0), behavior: 'smooth' });
	}

	function finish() {
		setShowOnboarding(false);
		goto('/login', { replaceState: true });
	}
</script>

<svelte:head><title>Bienvenida · Kümelen</title></svelte:head>

<main class="onboarding">
	<h1 class="sr-only">Bienvenida a Kümelen</h1>

	<!-- Un contenedor desplazable debe poder enfocarse para recorrerse con el teclado (regla de axe). -->
	<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
	<ul class="track" bind:this={track} onscroll={onScroll} aria-label="Presentación" tabindex="0">
		{#each slides as slide, index (slide.key)}
			<li class="slide" aria-label="{index + 1} de {slides.length}">
				{#if slide.imageFirst}
					<img src={slide.image} alt="" />
				{/if}
				<h2 style:padding-top={slide.imageFirst ? '20px' : '40px'}>{slide.title}</h2>
				<p>{slide.text}</p>
				{#if !slide.imageFirst}
					<img src={slide.image} alt="" />
				{/if}
			</li>
		{/each}
	</ul>

	<footer class="footer">
		<div class="dots" role="group" aria-label="Página">
			{#each slides as slide, index (slide.key)}
				<button
					type="button"
					class="dot"
					class:active={index === current}
					aria-label="Ir a la página {index + 1}"
					aria-current={index === current ? 'true' : undefined}
					onclick={() => show(index)}
				></button>
			{/each}
		</div>
		<button type="button" class="btn action" onclick={finish}>{isLast ? 'Salir' : 'Saltar'}</button>
	</footer>
</main>

<style>
	.onboarding {
		display: flex;
		flex: 1;
		flex-direction: column;
		height: 100dvh;
	}
	.track {
		display: flex;
		flex: 1;
		min-height: 0;
		overflow-x: auto;
		scroll-snap-type: x mandatory;
		scrollbar-width: none;
	}
	.track::-webkit-scrollbar {
		display: none;
	}
	.slide {
		display: flex;
		flex: none;
		flex-direction: column;
		align-items: center;
		width: 100%;
		padding: var(--safe-top) var(--pad) 0;
		overflow-y: auto;
		scroll-snap-align: start;
		text-align: center;
	}
	.slide img {
		width: 100%;
		max-height: 38dvh;
		margin: 20px 0;
		object-fit: contain;
	}
	.slide:first-child img {
		margin-top: 40px;
	}
	h2 {
		padding-bottom: 20px;
		font-size: 40px;
		font-weight: 700;
	}
	.slide p {
		padding: 20px 0;
		font-size: 20px;
	}
	.footer {
		display: flex;
		flex: none;
		flex-direction: column;
		align-items: center;
		gap: 14px;
		padding: 12px var(--pad) calc(16px + var(--safe-bottom));
	}
	.dots {
		display: flex;
		gap: 12px;
	}
	.dot {
		width: 24px;
		height: 24px;
		padding: 0;
		border: 2px solid var(--primary);
		border-radius: 50%;
		background: var(--primary);
		cursor: pointer;
	}
	.dot.active {
		background: var(--bg);
	}
	.action {
		align-self: flex-end;
		min-width: 110px;
		min-height: 48px;
		font-size: 20px;
		font-weight: 400;
	}
</style>
