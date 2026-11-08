// BFF simulado para desarrollar y probar la PWA sin `iluvatar` ni Strapi.
// Responde con la misma forma que los endpoints reales (ver src/lib/api/types.ts).
// Uso: npm run mock
import { createServer } from 'node:http';

const PORT = Number(process.env.MOCK_PORT ?? 8787);

const colors = ['#459951', '#00417c', '#915fb5', '#ca436b', '#e0a030', '#2a7f9e'];

/** Imagen de relleno generada al vuelo: /img/<texto>.svg */
function image(label, index = 0, size = [640, 360]) {
	const [w, h] = size;
	const color = colors[index % colors.length];
	return `<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}" viewBox="0 0 ${w} ${h}">
<rect width="100%" height="100%" fill="${color}"/>
<circle cx="${w * 0.8}" cy="${h * 0.25}" r="${h * 0.28}" fill="#ffffff22"/>
<text x="50%" y="52%" text-anchor="middle" font-family="sans-serif" font-size="${Math.round(h / 7)}" font-weight="700" fill="#fff">${label}</text>
</svg>`;
}

const img = (label, i, size) => `http://localhost:${PORT}/img/${encodeURIComponent(label)}.svg?i=${i}${size ? `&w=${size[0]}&h=${size[1]}` : ''}`;

const lorem =
	'Esta es una publicación de ejemplo. Su contenido llega desde el CMS como Markdown y se muestra con el mismo estilo que la app original.';

const services = [
	{
		id: 'pae',
		abbreviation: 'PAE',
		name: 'Plan de Apoyo Estudiantil',
		posts: [
			['hab', 'Hábitos de estudio', 'Para terminar la semana y empezar con un buen ritmo, revisa estos consejos sobre hábitos de estudio.'],
			['qs', '¿Qué es PAE?', 'Conoce el Plan de Apoyo Estudiantil y cómo acompaña tu proceso en la universidad.'],
			['org', 'Organiza tu tiempo', 'Una guía corta para repartir tus horas entre clases, estudio y descanso.']
		]
	},
	{
		id: 'sesaes',
		abbreviation: 'SESAES',
		name: 'Servicio de Salud Estudiantil',
		posts: [
			['aut', 'Autocuidado en época de pruebas', 'Pequeñas rutinas que ayudan a mantener la calma cuando se acumulan las evaluaciones.'],
			['sal', 'Atención de salud', 'Horarios y formas de agendar una hora con el equipo de salud estudiantil.']
		]
	},
	{
		id: 'deportes',
		abbreviation: 'Deportes',
		name: 'Servicio de Educación Física, Deporte y Recreación',
		posts: [['sel', 'Selecciones deportivas', 'Inscríbete en las selecciones y talleres deportivos del semestre.']]
	}
];

const categories = [
	{ id: 'apoyo', name: 'Apoyo Académico' },
	{ id: 'habitos', name: 'Hábitos' },
	{ id: 'primeros', name: 'Primeros Años' },
	{ id: 'ramos', name: 'Ramos difíciles' },
	{ id: 'salud', name: 'Salud' },
	{ id: 'deporte', name: 'Deporte' }
];

const allPosts = services.flatMap((service, si) =>
	service.posts.map(([id, title, summary], pi) => ({
		id,
		title,
		summary,
		image_url: img(title, si + pi),
		service,
		categories: categories.filter((_, ci) => (ci + si + pi) % 3 === 0)
	}))
);

const preview = (post) => ({ id: post.id, title: post.title, summary: post.summary, image_url: post.image_url });

const postBody = (title) => `# ${title}

${lorem}

## Qué debes saber

- Es un contenido **de ejemplo** para probar la maqueta.
- Admite *énfasis*, listas y [enlaces externos](https://www.utem.cl).
- Las imágenes del cuerpo usan rutas relativas al CMS.

1. Primer paso
2. Segundo paso
3. Tercer paso

![Imagen de ejemplo](/img/Ilustración.svg?i=2)

Para más información escribe a [correo@example.com](mailto:correo@example.com).`;

const notifications = [
	{ id: 'n1', title: 'Nuevo taller de hábitos', body: '# Taller de hábitos\nEste jueves a las **15:00** en el campus. [Más información](https://www.utem.cl)', service_name: 'PAE', service_image_url: img('PAE', 0, [200, 200]) },
	{ id: 'n2', title: 'Campaña de vacunación', body: 'Ya puedes agendar tu hora con el servicio de salud.', service_name: 'SESAES', service_image_url: img('SESAES', 1, [200, 200]) },
	{ id: 'n3', title: 'Inscripciones abiertas', body: 'Se abrió el proceso de inscripción a selecciones deportivas.', service_name: 'Deportes', service_image_url: img('Deportes', 2, [200, 200]) }
];

const serviceDetail = (service, si) => ({
	id: service.id,
	logo_url: img(service.abbreviation, si, [400, 400]),
	name: service.name,
	description: `# ${service.name}\n\n${lorem}\n\n## ¿Qué hacemos?\n\nAcompañamos a las y los estudiantes durante su formación.`,
	persons: [
		{ id: `${service.id}-1`, name: 'Camila Rojas', charge: 'Coordinadora', email: `camila.rojas@${service.id}.example.com`, profile_picture_url: img('CR', si, [200, 200]) },
		{ id: `${service.id}-2`, name: 'Diego Fuentes', charge: 'Profesional de apoyo', email: `diego.fuentes@${service.id}.example.com`, profile_picture_url: img('DF', si + 1, [200, 200]) }
	],
	contacts: [
		{ id: `${service.id}-c1`, name: 'Facebook', type: 'Facebook', link: 'https://www.facebook.com' },
		{ id: `${service.id}-c2`, name: 'Instagram', type: 'Instagram', link: 'https://www.instagram.com' },
		{ id: `${service.id}-c3`, name: 'Correo', type: 'Correo', link: `contacto@${service.id}.example.com` },
		{ id: `${service.id}-c4`, name: 'Sitio web', type: 'Web', link: 'https://www.utem.cl' }
	]
});

function route(pathname) {
	let match;
	if (pathname === '/v1/home') {
		return {
			services_with_posts: services.map((service) => ({
				id: service.id,
				abbreviation: service.abbreviation,
				posts_for_preview: allPosts.filter((post) => post.service.id === service.id).map(preview)
			}))
		};
	}
	if (pathname === '/v1/catalog') {
		return {
			services: services.map((service, i) => ({ id: service.id, name: service.name, logo_url: img(service.abbreviation, i, [400, 200]) })),
			categories
		};
	}
	if (pathname === '/v1/notification') return { notifications };
	if ((match = pathname.match(/^\/v1\/post\/search\/(.+)$/))) {
		const term = decodeURIComponent(match[1]).toLowerCase();
		return {
			posts: allPosts
				.filter((post) => `${post.title} ${post.summary}`.toLowerCase().includes(term))
				.map(preview)
		};
	}
	if ((match = pathname.match(/^\/v1\/post\/([^/]+)$/))) {
		const post = allPosts.find((p) => p.id === decodeURIComponent(match[1]));
		if (!post) return null;
		return {
			id: post.id,
			title: post.title,
			image: img(post.title, 0, [900, 540]),
			body: postBody(post.title),
			service_id: post.service.id,
			service_name: post.service.name,
			categories: post.categories
		};
	}
	if ((match = pathname.match(/^\/v1\/service\/([^/]+)\/posts$/))) {
		const si = services.findIndex((s) => s.id === decodeURIComponent(match[1]));
		if (si < 0) return null;
		const service = services[si];
		return {
			id: service.id,
			name: service.name,
			posts_for_preview: allPosts.filter((post) => post.service.id === service.id).map(preview)
		};
	}
	if ((match = pathname.match(/^\/v1\/service\/([^/]+)$/))) {
		const si = services.findIndex((s) => s.id === decodeURIComponent(match[1]));
		return si < 0 ? null : serviceDetail(services[si], si);
	}
	if ((match = pathname.match(/^\/v1\/category\/([^/]+)\/posts$/))) {
		const category = categories.find((c) => c.id === decodeURIComponent(match[1]));
		if (!category) return null;
		return {
			id: category.id,
			name: category.name,
			posts_for_preview: allPosts.filter((post) => post.categories.some((c) => c.id === category.id)).map(preview)
		};
	}
	return undefined;
}

const cors = {
	'Access-Control-Allow-Origin': '*',
	'Access-Control-Allow-Headers': 'Authorization, Content-Type',
	'Access-Control-Allow-Methods': 'GET, OPTIONS'
};

createServer((req, res) => {
	const url = new URL(req.url ?? '/', `http://localhost:${PORT}`);

	if (req.method === 'OPTIONS') {
		res.writeHead(204, cors).end();
		return;
	}

	if (url.pathname.startsWith('/img/')) {
		const label = decodeURIComponent(url.pathname.slice(5).replace(/\.svg$/, ''));
		const size = url.searchParams.has('w') ? [Number(url.searchParams.get('w')), Number(url.searchParams.get('h'))] : undefined;
		res.writeHead(200, { ...cors, 'Content-Type': 'image/svg+xml', 'Cache-Control': 'public, max-age=3600' });
		res.end(image(label.replace(/[<>&]/g, ''), Number(url.searchParams.get('i') ?? 0), size));
		return;
	}

	// Igual que el BFF real, exige un token (acepta cualquiera) y responde 401 sin él.
	if (!req.headers.authorization?.startsWith('Bearer ')) {
		res.writeHead(401, { ...cors, 'Content-Type': 'application/json' }).end(JSON.stringify({ message: 'empty token' }));
		return;
	}

	const body = route(url.pathname);
	if (body === undefined) {
		res.writeHead(404, { ...cors, 'Content-Type': 'application/json' }).end(JSON.stringify({ message: 'not found' }));
	} else if (body === null) {
		res.writeHead(404, { ...cors, 'Content-Type': 'application/json' }).end(JSON.stringify({ message: 'resource not found' }));
	} else {
		res.writeHead(200, { ...cors, 'Content-Type': 'application/json' }).end(JSON.stringify(body));
	}
}).listen(PORT, () => console.log(`BFF simulado en http://localhost:${PORT}`));
