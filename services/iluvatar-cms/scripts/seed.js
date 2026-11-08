'use strict';

// Datos de prueba para desarrollo: servicios, personas, contactos, categorías, posts y avisos,
// más un API token de solo lectura para el BFF. No se ejecuta en producción.
//
//   npm run seed
//   SEED_TOKEN_FILE=/ruta/token npm run seed   # escribe el token en el archivo en vez de mostrarlo

const fs = require('fs');
const os = require('os');
const path = require('path');
const { createStrapi, compileStrapi } = require('@strapi/strapi');

// PNG de 1x1 píxel.
const PIXEL_PNG = Buffer.from(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==',
  'base64'
);

const TOKEN_NAME = 'iluvatar-bff-dev';

async function uploadImage(strapi, name) {
  const filepath = path.join(os.tmpdir(), `${name}.png`);
  fs.writeFileSync(filepath, PIXEL_PNG);
  const [file] = await strapi.plugin('upload').service('upload').upload({
    data: { fileInfo: { name, alternativeText: name } },
    files: { filepath, originalFilename: `${name}.png`, mimetype: 'image/png', size: PIXEL_PNG.length },
  });
  return file.id;
}

const create = (strapi, uid, data) => strapi.documents(uid).create({ data, status: 'published' });

async function seedContent(strapi) {
  if ((await strapi.documents('api::servicio.servicio').count({})) > 0) {
    strapi.log.info('Ya hay servicios: no se cargan datos de prueba.');
    return;
  }

  const salud = await create(strapi, 'api::categoria.categoria', { Titulo: 'Salud' });
  const deporte = await create(strapi, 'api::categoria.categoria', { Titulo: 'Deporte' });
  const beneficios = await create(strapi, 'api::categoria.categoria', { Titulo: 'Beneficios' });

  const servicios = [
    { Nombre: 'Plan de Apoyo Estudiantil', Sigla: 'PAE', Descripcion: 'Acompaña a los estudiantes en su **formación integral**.', posts: 6, categorias: [beneficios, salud] },
    { Nombre: 'Servicio de Salud Estudiantil', Sigla: 'SESAES', Descripcion: 'Atención de salud para estudiantes.', posts: 1, categorias: [salud] },
    { Nombre: 'Servicio de Deportes', Sigla: 'Deportes', Descripcion: 'Actividad física y recreación.', posts: 0, categorias: [deporte] },
  ];

  for (const [index, definition] of servicios.entries()) {
    const servicio = await create(strapi, 'api::servicio.servicio', {
      Nombre: definition.Nombre,
      Sigla: definition.Sigla,
      Descripcion: definition.Descripcion,
      Logo: await uploadImage(strapi, `logo-${definition.Sigla.toLowerCase()}`),
    });

    await create(strapi, 'api::persona.persona', {
      Nombre: `Profesional ${definition.Sigla}`,
      Cargo: 'Coordinación',
      Correo: `${definition.Sigla.toLowerCase()}@example.com`,
      servicio: servicio.documentId,
    });
    await create(strapi, 'api::contacto.contacto', {
      Tipo: 'Web',
      Nombre: `Sitio de ${definition.Sigla}`,
      Link: `https://example.com/${definition.Sigla.toLowerCase()}`,
      servicio: servicio.documentId,
    });
    await create(strapi, 'api::aviso.aviso', {
      Titulo: `Aviso de ${definition.Sigla}`,
      Cuerpo: 'Texto del aviso.',
      servicio: servicio.documentId,
    });

    for (let n = 1; n <= definition.posts; n++) {
      await create(strapi, 'api::post.post', {
        Titulo: `${definition.Sigla}: publicación ${n}`,
        Resumen: `Resumen de la publicación ${n} de ${definition.Sigla}.`,
        Cuerpo: `# Publicación ${n}\n\nContenido de prueba de **${definition.Sigla}**.`,
        Imagen: await uploadImage(strapi, `post-${index}-${n}`),
        servicio: servicio.documentId,
        categorias: definition.categorias.map((categoria) => categoria.documentId),
      });
    }
  }

  strapi.log.info('Datos de prueba cargados.');
}

async function seedApiToken(strapi) {
  const tokens = strapi.service('admin::api-token');
  // SEED_TOKEN_FILE: archivo donde dejar el token para que otro contenedor lo lea (docker compose de desarrollo).
  const tokenFile = process.env.SEED_TOKEN_FILE;
  const existing = await tokens.getByName(TOKEN_NAME);

  if (existing && (!tokenFile || fs.existsSync(tokenFile))) {
    strapi.log.info(`El API token "${TOKEN_NAME}" ya existe (solo se muestra al crearlo).`);
    return;
  }

  // Si el archivo se perdió pero el token sigue en la base, se regenera: la clave solo se conoce al crearla.
  const token = existing
    ? await tokens.regenerate(existing.id)
    : await tokens.create({
        name: TOKEN_NAME,
        description: 'Lectura para el BFF en desarrollo',
        type: 'read-only',
        lifespan: null,
      });

  if (tokenFile) {
    // 0644: lo lee otro contenedor con otro usuario. Es un token de solo lectura de un entorno de desarrollo.
    fs.writeFileSync(tokenFile, token.accessKey, { mode: 0o644 });
    strapi.log.info(`API token escrito en ${tokenFile}.`);
    return;
  }
  // eslint-disable-next-line no-console
  console.log(`\nILUVATAR_TOKEN=${token.accessKey}\n`);
}

async function main() {
  if (process.env.NODE_ENV === 'production') {
    throw new Error('El seed es solo para desarrollo (NODE_ENV=production).');
  }
  const app = await createStrapi(await compileStrapi()).load();
  try {
    await seedContent(app);
    await seedApiToken(app);
  } finally {
    await app.destroy();
  }
}

main().then(
  () => process.exit(0),
  (error) => {
    // eslint-disable-next-line no-console
    console.error(error);
    process.exit(1);
  }
);
