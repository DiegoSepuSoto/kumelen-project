import DOMPurify from 'dompurify';
import { Marked } from 'marked';
import { resolveUrl } from './format';

const parser = new Marked({ gfm: true });

/**
 * Convierte el Markdown de una publicación a HTML seguro. El contenido lo escriben los equipos
 * de cada servicio desde el CMS, así que siempre se sanea antes de insertarlo.
 *
 * Las imágenes relativas se resuelven contra el CMS (igual que imageBuilder en manwe) y los
 * enlaces se marcan para que la interfaz pida confirmación antes de abrirlos.
 */
export function renderMarkdown(source: string, cmsBase: string): string {
	if (typeof window === 'undefined') return '';

	const html = parser.parse(source, { async: false });
	const fragment = DOMPurify.sanitize(html, {
		USE_PROFILES: { html: true },
		RETURN_DOM_FRAGMENT: true
	});

	for (const image of fragment.querySelectorAll('img')) {
		image.setAttribute('src', resolveUrl(image.getAttribute('src') ?? '', cmsBase));
		image.setAttribute('loading', 'lazy');
		if (!image.hasAttribute('alt')) image.setAttribute('alt', '');
	}

	for (const link of fragment.querySelectorAll('a[href]')) {
		link.setAttribute('rel', 'noopener noreferrer');
		link.setAttribute('target', '_blank');
	}

	const container = document.createElement('div');
	container.append(fragment);
	return container.innerHTML;
}
