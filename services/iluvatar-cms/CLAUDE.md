# iluvatar-cms — CMS de Kümelen

Strapi 5 (JavaScript) con PostgreSQL en producción y SQLite en desarrollo. Contenido: `servicio`, `persona`, `contacto`, `categoria`, `post`, `aviso`. Los consume `iluvatar` por REST con un API token de solo lectura. Ver `README.md` para variables de entorno y puesta en marcha.

## Comandos

```bash
nvm use 24                 # Node 22 a 26; con Node 16 o 20 Strapi no arranca
npm install
npm run seed               # datos de prueba + API token de solo lectura (solo desarrollo)
npm run develop            # panel en http://localhost:1337/admin
npm run start              # sin panel recompilado; suficiente para la API
```

## Estructura

- `src/api/<tipo>/content-types/<tipo>/schema.json`: esquema de cada tipo. Rutas, controladores y servicios son los de fábrica (`factories.createCore*`), sin lógica propia.
- `src/index.js`: al arrancar registra el webhook `iluvatar-notificaciones` (evento `entry.publish`) si existe `ILUVATAR_WEBHOOK_URL`.
- `config/server.js`: `webhooks.populateRelations` en verdadero (el webhook necesita `servicio.Nombre` y `Sigla`) y cabecera `X-Webhook-Secret` desde `WEBHOOK_SECRET`.
- `config/admin.js`: `ADMIN_COOKIE_SECURE` (vacía = valor de Strapi) es una salida de emergencia para probar el panel por HTTP plano; el panel la lee al compilarse (`ARG` del Dockerfile). Una variable vacía se trata como sin definir porque `env.bool` la leería como `false`.
- `scripts/seed.js`: datos de prueba (3 servicios, 7 posts, categorías, avisos). Es idempotente y se niega a correr con `NODE_ENV=production`. Con `SEED_TOKEN_FILE` escribe el API token en un archivo (`/shared` en la imagen está para eso).

## Reglas

- Al cambiar un esquema, regrabar los fixtures de `iluvatar` (`go test -tags record ...`) y correr su prueba `e2e`.
- Los campos conservan los nombres en español de Strapi 3 (`Titulo`, `Resumen`, `Cuerpo`, `Imagen`...).
- Nunca versionar `.env`, `.tmp/` ni `public/uploads/`.
- PostgreSQL probado (Strapi 5.61 con PostgreSQL 17: seed, API y e2e de `iluvatar`). Las imágenes se guardan en disco local, que no persiste en contenedores: falta decidir un proveedor de almacenamiento.
- `Dockerfile` multi-etapa (Node 24, usuario `node`, panel compilado). Con Podman: `podman build --format docker -t iluvatar-cms .`.
