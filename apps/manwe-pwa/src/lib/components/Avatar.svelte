<script lang="ts">
	/** CircleAvatar anidado de Flutter: un anillo de color y, dentro, la imagen. */
	let {
		src,
		size,
		alt = '',
		ring = 5,
		ringColor = 'var(--primary)',
		fit = 'cover'
	}: {
		src: string;
		/** Diámetro exterior en px. */
		size: number;
		alt?: string;
		ring?: number;
		ringColor?: string;
		fit?: 'cover' | 'contain';
	} = $props();
</script>

<span class="avatar" style:--size="{size}px" style:--ring="{ring}px" style:--ring-color={ringColor}>
	<img {src} {alt} loading="lazy" decoding="async" style:object-fit={fit} />
</span>

<style>
	.avatar {
		display: inline-flex;
		flex: none;
		width: var(--size);
		height: var(--size);
		padding: var(--ring);
		border-radius: 50%;
		background: var(--ring-color);
	}
	img {
		width: 100%;
		height: 100%;
		border-radius: 50%;
		background: var(--primary);
	}
</style>
