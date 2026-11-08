<script lang="ts">
	import { api } from '$lib/api';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import LoadingCube from '$lib/components/LoadingCube.svelte';
	import Markdown from '$lib/components/Markdown.svelte';
	import { resource } from '$lib/resource.svelte';

	const notifications = resource((signal) => api.notifications(signal));
</script>

<svelte:head><title>Notificaciones · Kümelen</title></svelte:head>

<h1 class="sr-only">Notificaciones</h1>

{#if notifications.status === 'loading'}
	<LoadingCube />
{:else if notifications.status === 'success' && notifications.data}
	<ul class="list">
		{#each notifications.data.notifications as notification (notification.id)}
			<li class="card card-shadow">
				<article class="inner">
					<img class="logo" src={notification.service_image_url} alt="" loading="lazy" decoding="async" />
					<div class="text">
						<h2>{notification.title}</h2>
						<span class="service sr-only">De {notification.service_name}</span>
						<Markdown source={notification.body} variant="notification" />
					</div>
				</article>
			</li>
		{/each}
	</ul>
{:else}
	<ErrorState />
{/if}

<style>
	.list {
		display: flex;
		flex-direction: column;
		gap: 15px;
		margin: var(--pad) 0;
	}
	.card {
		margin: 0 var(--pad);
		padding-bottom: 2px;
	}
	.inner {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding: 10px;
		border-radius: 10px;
		background: var(--white);
	}
	.logo {
		flex: none;
		width: 70px;
		height: 70px;
		border-radius: 50%;
		object-fit: cover;
		background: var(--primary);
	}
	.text {
		flex: 1;
		min-width: 0;
	}
	h2 {
		margin-bottom: 5px;
		font-size: 16px;
		font-weight: 700;
	}
</style>
