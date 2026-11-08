<script lang="ts">
	import type { PostPreview } from '$lib/api/types';
	import Avatar from './Avatar.svelte';

	let { posts }: { posts: PostPreview[] } = $props();
</script>

<!-- PostList + PostListItem de manwe: filas con imagen circular, título y resumen. -->
<ul class="list">
	{#each posts as post (post.id)}
		<li>
			<a class="item" href="/publicaciones/{post.id}">
				<span class="image"><Avatar src={post.image_url} size={90} /></span>
				<span class="text">
					<strong>{post.title}</strong>
					<span class="summary">{post.summary}</span>
				</span>
			</a>
		</li>
	{/each}
</ul>

<style>
	li:not(:last-child)::after {
		content: '';
		display: block;
		height: 16px;
		background: linear-gradient(#000, #000) center / 100% 1px no-repeat;
	}
	.item {
		display: flex;
		align-items: center;
		padding: var(--pad) 0;
	}
	.image {
		padding-left: var(--pad);
	}
	.text {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-width: 0;
		padding: 0 var(--pad);
	}
	strong {
		font-size: 20px;
		font-weight: 700;
	}
	.summary {
		font-size: 16px;
		font-style: italic;
		display: -webkit-box;
		-webkit-line-clamp: 4;
		line-clamp: 4;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
