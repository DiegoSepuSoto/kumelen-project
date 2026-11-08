<script lang="ts">
	import type { Contact } from '$lib/api/types';
	import Icon, { type IconName } from './Icon.svelte';

	let { contact }: { contact: Contact } = $props();

	/** Misma correspondencia tipo → botón que contactButtonBuilder de manwe. */
	const kind = $derived.by(() => {
		switch (contact.type) {
			case 'Facebook':
				return { icon: 'facebook' as IconName, label: 'Facebook', href: contact.link, width: 160, className: 'facebook' };
			case 'Instagram':
				return { icon: 'instagram' as IconName, label: 'Instagram', href: contact.link, width: 160, className: 'instagram' };
			case 'Correo':
				return { icon: 'mail' as IconName, label: 'Enviar Correo', href: `mailto:${contact.link}`, width: 200, className: 'mail' };
			default:
				return { icon: 'web' as IconName, label: 'Ver Sitio', href: contact.link, width: 150, className: 'web' };
		}
	});
</script>

<a
	class="button {kind.className}"
	style:width="{kind.width}px"
	href={kind.href}
	target={kind.href.startsWith('mailto:') ? undefined : '_blank'}
	rel="noopener noreferrer"
	aria-label="{kind.label}{contact.name ? `: ${contact.name}` : ''}"
>
	<Icon name={kind.icon} />
	<span>{kind.label}</span>
</a>

<style>
	.button {
		display: flex;
		align-items: center;
		justify-content: space-between;
		min-height: 36px;
		padding: 8px 16px;
		border-radius: 5px;
		color: var(--white);
		font-size: 20px;
		font-weight: 700;
		box-shadow: 0 4px 5px rgb(0 0 0 / 0.26);
	}
	.facebook {
		background: #0177fc;
	}
	.instagram {
		background: linear-gradient(to bottom right, #915fb5, #ca436b);
	}
	.mail {
		background: #ea4335;
	}
	.web {
		background: var(--primary);
	}
</style>
