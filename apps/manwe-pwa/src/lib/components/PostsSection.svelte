<script lang="ts">
	import type { PostPreview } from '$lib/api/types';
	import PostCard from './PostCard.svelte';

	let {
		id,
		abbreviation,
		posts
	}: { id: string; abbreviation: string; posts: PostPreview[] } = $props();
</script>

<section class="section" aria-labelledby="posts-of-{id}">
	<h2 id="posts-of-{id}">Últimas publicaciones de <em>{abbreviation}</em></h2>
	<ul class="list">
		{#each posts as post (post.id)}
			<li><PostCard {post} /></li>
		{/each}
	</ul>
</section>

<style>
	.section {
		padding-bottom: var(--pad);
	}
	h2 {
		padding: var(--pad) 0 0 var(--pad);
		font-size: 18px;
		font-weight: 400;
	}
	em {
		font-style: italic;
		font-weight: 700;
	}
	.list {
		display: flex;
		height: calc(47.5 * var(--vu));
		padding: var(--pad) 0 0 var(--pad);
		overflow-x: auto;
		overscroll-behavior-x: contain;
		scroll-snap-type: x proximity;
		scrollbar-width: none;
	}
	.list::-webkit-scrollbar {
		display: none;
	}
	li {
		display: flex;
		scroll-snap-align: start;
		scroll-margin-left: var(--pad);
	}
	@media (max-height: 600px) {
		.list {
			height: calc(60 * var(--vu));
		}
	}
</style>
