<script lang="ts">
	import Icon from './Icon.svelte';
</script>

<!-- Equivale a ErrorScreen. -->
<div class="state" role="alert">
	<span style:color="var(--error)"><Icon name="error" size={100} /></span>
	<p>Hubo un error :(</p>
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
		font-size: 30px;
		font-weight: 700;
	}
</style>
