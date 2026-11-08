# iluvatar-cms — CMS de Kümelen

CMS headless (Strapi 5) donde los servicios estudiantiles crean y editan el contenido que ve la app: servicios, personas, contactos, categorías, publicaciones (`post`) y avisos.

El servicio `iluvatar` lee este contenido por la API REST (`/api/...`) con un API token de solo lectura, y recibe de aquí el webhook que dispara las notificaciones push.

## Requisitos

- Node.js 22 a 26 (probado con 24.20) y npm.
- PostgreSQL en producción. En desarrollo se usa SQLite sin configurar nada.

## Puesta en marcha (desarrollo)

```bash
npm install
cp .env.example .env     # y reemplaza todos los "tobemodified" por valores aleatorios
npm run seed             # datos de prueba y un API token de solo lectura (se muestra una sola vez)
npm run develop          # panel de administración en http://localhost:1337/admin
```

`npm run seed` no corre en producción (`NODE_ENV=production`). Imprime `ILUVATAR_TOKEN=...`: es el token que usa `iluvatar` para leer el CMS. Con `SEED_TOKEN_FILE=/ruta` lo escribe en ese archivo en vez de mostrarlo (así lo entrega el compose de desarrollo; si el archivo se pierde y el token sigue en la base, lo regenera).

## Variables de entorno

| Variable | Uso |
|---|---|
| `APP_KEYS`, `API_TOKEN_SALT`, `ADMIN_JWT_SECRET`, `TRANSFER_TOKEN_SALT`, `JWT_SECRET`, `ENCRYPTION_KEY` | Secretos de Strapi. Obligatorios y sin valor por defecto. |
| `DATABASE_CLIENT` | `sqlite` (por defecto) o `postgres`. |
| `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_NAME`, `DATABASE_USERNAME`, `DATABASE_PASSWORD`, `DATABASE_SSL` | Conexión a PostgreSQL. |
| `ILUVATAR_WEBHOOK_URL` | URL de `iluvatar` (`.../v1/hooks/post`). Al arrancar se registra un webhook `entry.publish` hacia ella. |
| `WEBHOOK_SECRET` | Secreto compartido; Strapi lo envía en la cabecera `X-Webhook-Secret` y `iluvatar` lo exige. |

## Tipos de contenido

| Tipo | Ruta REST | Borrador y publicación |
|---|---|---|
| `servicio` | `/api/servicios` | sí |
| `persona` | `/api/personas` | sí |
| `contacto` | `/api/contactos` | sí |
| `categoria` | `/api/categorias` | no |
| `post` | `/api/posts` | sí |
| `aviso` | `/api/avisos` | sí |

Los esquemas están en `src/api/*/content-types/*/schema.json`. Las consultas usan el `documentId` de cada documento como identificador.

## Contenedor

```bash
docker build -t iluvatar-cms .
docker run -p 1337:1337 --env-file .env -e NODE_ENV=production -e DATABASE_CLIENT=postgres -e DATABASE_HOST=... iluvatar-cms
```

Corre como usuario `node` con el panel ya compilado. Probado contra PostgreSQL 17. Las sondas pueden usar `GET /_health` (204).

Tras un proxy con TLS define `PUBLIC_URL`, `IS_PROXIED=true` y `CORS_ORIGINS` (ver `.env.example`).

## Producción

- Crear un API token de **solo lectura** en *Settings → API Tokens* y entregarlo a `iluvatar` como `ILUVATAR_TOKEN`.
- Las imágenes se guardan en `public/uploads` (disco local): en un contenedor se pierden al reiniciar. Hay que configurar un proveedor de almacenamiento (por ejemplo S3) antes de salir a producción; `iluvatar` ya acepta URLs absolutas de imágenes.
- El rol público no tiene permisos: todo el acceso a la API pasa por el token.
