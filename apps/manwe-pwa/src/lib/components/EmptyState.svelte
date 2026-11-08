<script lang="ts">
	import Icon from './Icon.svelte';
</script>

<!-- Equivale a NoContentScreen. -->
<div class="state" role="status">
	<span style:color="var(--warning)"><Icon name="warning" size={100} /></span>
	<p>No hay nada para mostrar</p>
</div>

<style>
	.state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 240px;
		height: 100%;
		text-align: center;
	}
	p {
		font-size: 28px;
		font-weight: 700;
	}
</style>
