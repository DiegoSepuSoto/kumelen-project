<!-- Equivale a SpinKitCubeGrid(color: kPrimaryColor, size: 100). -->
<div class="center" role="status">
	<span class="sr-only">Cargando…</span>
	<div class="grid" aria-hidden="true">
		{#each [0.2, 0.3, 0.4, 0.1, 0.2, 0.3, 0, 0.1, 0.2] as delay}
			<span style:animation-delay="{delay}s"></span>
		{/each}
	</div>
</div>

<style>
	.center {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 200px;
		height: 100%;
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		width: 100px;
		height: 100px;
	}
	span:not(.sr-only) {
		background: var(--primary);
		animation: cube-scale 1.3s infinite ease-in-out;
	}
	@keyframes cube-scale {
		0%,
		70%,
		100% {
			transform: scale3d(1, 1, 1);
		}
		35% {
			transform: scale3d(0, 0, 1);
		}
	}
</style>
