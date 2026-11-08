// '' (variable declarada pero vacía, como deja un ARG sin valor en el Dockerfile) cuenta como sin definir:
// env.bool() la leería como `false` y desactivaría las cookies seguras sin que nadie lo pida.
const optionalBool = (value) => (value === '' ? undefined : value === 'true');

module.exports = ({ env }) => ({
  auth: {
    secret: env('ADMIN_JWT_SECRET'),
    // Sin definir se usa el valor de Strapi (cookies `secure` en producción). Solo para probar el panel por HTTP plano
    // en local: ADMIN_COOKIE_SECURE=false, tanto al construir la imagen (ARG) como al ejecutarla.
    cookie: { secure: optionalBool(env('ADMIN_COOKIE_SECURE', '')) },
  },
  apiToken: {
    salt: env('API_TOKEN_SALT'),
  },
  transfer: {
    token: {
      salt: env('TRANSFER_TOKEN_SALT'),
    },
  },
  secrets: {
    encryptionKey: env('ENCRYPTION_KEY'),
  },
  flags: {
    nps: env.bool('FLAG_NPS', true),
    promoteEE: env.bool('FLAG_PROMOTE_EE', true),
    docLinks: env.bool('FLAG_DOC_LINKS', true),
  },
});
