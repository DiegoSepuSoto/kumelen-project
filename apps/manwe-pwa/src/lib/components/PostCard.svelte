<script lang="ts">
	import type { PostPreview } from '$lib/api/types';

	let { post }: { post: PostPreview } = $props();
</script>

<!-- PostCard de manwe: imagen, título, resumen y botón "Ver más". -->
<article class="card card-shadow">
	<div class="image"><img src={post.image_url} alt="" loading="lazy" decoding="async" /></div>
	<div class="body">
		<h3>{post.title}</h3>
		<p>{post.summary}</p>
		<a class="btn" href="/publicaciones/{post.id}" aria-label="Ver más: {post.title}">Ver más</a>
	</div>
</article>

<style>
	.card {
		flex: none;
		width: calc(65 * var(--wu));
		margin-right: var(--pad);
		padding-bottom: 2px;
	}
	.image {
		height: calc(19 * var(--vu));
		overflow: hidden;
		border-radius: 10px 10px 0 0;
		background: var(--white) url('/images/loading.gif') center / 60px no-repeat;
	}
	.image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.body {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		justify-content: space-between;
		height: calc(25 * var(--vu));
		padding: 12px;
		border-radius: 0 0 10px 10px;
		background: var(--white);
	}
	@media (max-height: 600px) {
		.image {
			height: calc(19 * var(--vu));
		}
		.body {
			height: calc(36.5 * var(--vu));
		}
	}
	h3 {
		width: 100%;
		font-size: 17px;
		font-weight: 700;
		color: #000;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	p {
		font-size: 16px;
		display: -webkit-box;
		-webkit-line-clamp: 3;
		line-clamp: 3;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
