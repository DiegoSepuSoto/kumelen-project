import { describe, expect, it } from 'vitest';
import { renderMarkdown } from '../src/lib/markdown';

const CMS = 'https://cms.test';

describe('renderMarkdown', () => {
	it('convierte encabezados, énfasis y listas', () => {
		const html = renderMarkdown('# Título\n\nTexto **fuerte**\n\n- uno\n- dos', CMS);

		expect(html).toContain('<h1>Título</h1>');
		expect(html).toContain('<strong>fuerte</strong>');
		expect(html).toContain('<li>uno</li>');
	});

	it('elimina scripts y manejadores de eventos del contenido del CMS', () => {
		const html = renderMarkdown(
			'<script>alert(1)</script>\n\n<img src="x" onerror="alert(2)">\n\n[x](javascript:alert(3))',
			CMS
		);

		expect(html).not.toContain('<script');
		expect(html).not.toContain('onerror');
		expect(html).not.toContain('javascript:');
	});

	it('resuelve las imágenes relativas contra el CMS y deja intactas las absolutas', () => {
		const html = renderMarkdown('![a](/uploads/foto.png)\n\n![b](https://otro.test/b.png)', CMS);

		expect(html).toContain('src="https://cms.test/uploads/foto.png"');
		expect(html).toContain('src="https://otro.test/b.png"');
		expect(html).toContain('loading="lazy"');
	});

	it('abre los enlaces en otra pestaña sin dar acceso a la ventana original', () => {
		const html = renderMarkdown('[UTEM](https://www.utem.cl)', CMS);

		expect(html).toContain('rel="noopener noreferrer"');
		expect(html).toContain('target="_blank"');
	});

	it('devuelve una cadena vacía para un texto vacío', () => {
		expect(renderMarkdown('', CMS)).toBe('');
	});
});
