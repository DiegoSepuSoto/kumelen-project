module.exports = ({ env }) => ({
  host: env('HOST', '0.0.0.0'),
  port: env.int('PORT', 1337),
  // Detrás de un proxy con TLS: URL pública del CMS y confianza en X-Forwarded-*, necesarias para las cookies seguras del panel.
  url: env('PUBLIC_URL', ''),
  proxy: env.bool('IS_PROXIED', false),
  app: {
    keys: env.array('APP_KEYS'),
  },
  webhooks: {
    // iluvatar necesita el servicio (Nombre, Sigla) dentro del payload para armar la notificación.
    populateRelations: env.bool('WEBHOOKS_POPULATE_RELATIONS', true),
    // iluvatar rechaza los webhooks que no traen este secreto.
    defaultHeaders: {
      'X-Webhook-Secret': env('WEBHOOK_SECRET', ''),
    },
  },
});
