module.exports = ({ env }) => [
  'strapi::logger',
  'strapi::errors',
  'strapi::security',
  {
    // La API la consume iluvatar desde el servidor (sin CORS); solo el panel de administración vive en el navegador.
    name: 'strapi::cors',
    config: {
      origin: env.array('CORS_ORIGINS', []),
      methods: ['GET', 'HEAD', 'OPTIONS'],
    },
  },
  'strapi::poweredBy',
  'strapi::query',
  'strapi::body',
  'strapi::session',
  'strapi::favicon',
  'strapi::public',
];
