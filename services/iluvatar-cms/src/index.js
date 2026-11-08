'use strict';

const WEBHOOK_NAME = 'iluvatar-notificaciones';

module.exports = {
  register(/*{ strapi }*/) {},

  async bootstrap({ strapi }) {
    await ensureNotificationWebhook(strapi);
  },
};

// Registra (una sola vez) el webhook que avisa a iluvatar cuando se publica contenido.
// Si ILUVATAR_WEBHOOK_URL no está definida no hace nada: el webhook se puede crear a mano en el panel.
async function ensureNotificationWebhook(strapi) {
  const url = process.env.ILUVATAR_WEBHOOK_URL;
  if (!url) {
    strapi.log.warn('ILUVATAR_WEBHOOK_URL no está definida: no se registra el webhook de notificaciones.');
    return;
  }
  if (!process.env.WEBHOOK_SECRET) {
    strapi.log.warn('WEBHOOK_SECRET no está definida: iluvatar rechazará los webhooks sin secreto.');
  }

  const store = strapi.get('webhookStore');
  const existing = (await store.findWebhooks()).find((webhook) => webhook.name === WEBHOOK_NAME);
  const attributes = { name: WEBHOOK_NAME, url, headers: {}, events: ['entry.publish'], isEnabled: true };

  if (existing) {
    await store.updateWebhook(existing.id, { ...existing, ...attributes });
  } else {
    await store.createWebhook(attributes);
  }
}
