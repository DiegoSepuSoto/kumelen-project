# Kümelen completo con Docker Compose

`compose.yaml` levanta todo el sistema en tu máquina: PostgreSQL, el CMS, la API/BFF, Laurelin (registro de estudiantes) y la PWA. Es para **desarrollo y pruebas locales**, no para producción (puertos en `localhost`, datos de prueba, sin TLS).

> **Estado: escrito y validado, no ejecutado.** Ver "Qué está verificado" al final.

## Qué levanta

| Servicio | Qué es | Dirección |
|---|---|---|
| `postgres` | PostgreSQL 17 (volumen `pgdata`) | solo red interna |
| `cms` | Strapi 5 (`iluvatar-cms`) | http://localhost:1337 (panel en `/admin`) |
| `cms-seed` | Tarea que corre una vez: carga datos de prueba y deja el API token en un volumen compartido | — |
| `iluvatar` | API + BFF (Go) | http://localhost:8080 (`/swagger/index.html`) |
| `laurelin-db` | PostgreSQL 17 propio de Laurelin (volumen `laurelin_pgdata`) | solo red interna |
| `laurelin` | Registro de estudiantes (Go), solo API | http://localhost:8090 |
| `pwa` | La app web (`manwe-pwa`, SvelteKit) | http://localhost:3000 |
| `firebase-emulator` | Emulador de Firebase Auth, **solo con** `--profile emulator` | http://localhost:9099 |

Orden de arranque (lo resuelve Compose): `postgres` → `cms` → `cms-seed` → `iluvatar` → `pwa`; en paralelo, `laurelin-db` → `laurelin` (por ahora `iluvatar` no lo consulta).

## Requisitos

- Docker con Compose v2 (usa `depends_on` con condiciones y perfiles). Validado con Compose 5.6.
- Estar dentro del monorepo: el compose construye `../services/iluvatar`, `../services/iluvatar-cms` y `../apps/manwe-pwa`. Todos los comandos se ejecutan desde `deploy/`.
- Puertos libres: 1337, 8080, 8090, 3000 (y 9099 con el emulador). Se pueden cambiar con `CMS_PORT`, `ILUVATAR_PORT`, `LAURELIN_PORT`, `PWA_PORT` en `.env`.

## Primer arranque

```bash
./compose-env.sh               # crea .env con secretos aleatorios (no sobrescribe uno existente)
docker compose up --build      # con un proyecto de Firebase real (opción B, abajo)
# o
docker compose --profile emulator up --build   # con el emulador de Auth (opción A, abajo)
```

La primera vez construye tres imágenes y tarda unos minutos. Después:

1. Abre http://localhost:1337/admin y crea la primera cuenta de administración del CMS.
2. Los datos de prueba (3 servicios, 7 publicaciones, categorías y avisos) ya están cargados por `cms-seed`.
3. Abre http://localhost:3000 para la PWA.

## Firebase: dos opciones

**A) Emulador de Auth (sin proyecto real).** Es lo que deja `./compose-env.sh` por defecto. El botón de Google de la PWA abre una ventana falsa del emulador donde eliges o creas una cuenta; también puedes crear usuarios y llamar a `iluvatar` directamente con sus tokens. La ventana del emulador es un popup nuevo: pruébalo en un navegador completo. Los tokens del emulador no están firmados, así que solo sirve para desarrollo.

**B) Proyecto real de Firebase de pruebas (nunca el de producción).** Para usar la PWA de verdad:

1. En `.env`: `FIREBASE_PROJECT_ID=<tu proyecto>`, `FIREBASE_AUTH_EMULATOR_HOST=` y `PUBLIC_FIREBASE_AUTH_EMULATOR_HOST=` (ambas vacías), `GOOGLE_APPLICATION_CREDENTIALS=/secrets/firebase.json` y las cinco variables `PUBLIC_FIREBASE_*` de la configuración web de la app.
2. Guarda la cuenta de servicio en `./secrets/firebase.json` (la carpeta `secrets/` se monta de solo lectura y no se versiona).
3. En la consola de Firebase, agrega `localhost` a los dominios autorizados de Authentication y habilita el acceso con Google.
4. `docker compose up --build` (sin el perfil `emulator`).

Las notificaciones push están **desactivadas** (`PUSH_DISABLED=true`) para no enviar mensajes reales desde un entorno de pruebas. Con la opción B puedes poner `PUSH_DISABLED=false`.

## Operación diaria

```bash
docker compose up -d                 # en segundo plano
docker compose logs -f iluvatar      # logs de un servicio (JSON)
docker compose ps                    # estado y salud
docker compose down                  # detiene (conserva los datos)
docker compose down -v               # detiene y BORRA datos, imágenes subidas y el token
```

Tras `down -v` el siguiente `up` vuelve a cargar los datos de prueba y genera un token nuevo. Si solo borras el volumen `shared` y conservas la base, `cms-seed` regenera el token.

## Laurelin en el compose

`laurelin` no tiene interfaz: se opera por API. Para probarlo con el emulador de Auth:

1. Crea en el emulador una cuenta con el correo de `LAURELIN_BOOTSTRAP_ADMIN_EMAIL` (por defecto `admin@laurelin.test`) y márcala como verificada. El emulador acepta `POST http://localhost:9099/identitytoolkit.googleapis.com/v1/accounts:signUp?key=fake` y luego `POST .../v1/projects/demo-kumelen/accounts:update` con `Authorization: Bearer owner` y `{"localId": "...", "emailVerified": true}`. Vuelve a iniciar sesión (`accounts:signInWithPassword`) para obtener un ID token con el correo verificado.
2. Llama a `http://localhost:8090/v1/me` con `Authorization: Bearer <ID token>`. El resto de las rutas están en `services/laurelin/CLAUDE.md`.
3. La consulta que hará `iluvatar` se prueba con `POST /internal/v1/students/lookup` y `Authorization: Bearer $LAURELIN_SERVICE_TOKEN` (el valor está en `.env`).

Con un proyecto real de Firebase, `LAURELIN_BOOTSTRAP_ADMIN_EMAIL` debe ser tu correo real (verificado por Google al iniciar sesión).

## Por qué hay dos direcciones para el CMS

`iluvatar` consulta al CMS por la red interna (`http://cms:1337`), pero las imágenes que devuelve las carga **el navegador**, que no resuelve `cms`. Por eso `iluvatar` recibe `CMS_PUBLIC_URL=http://localhost:1337` para armar las URLs de imágenes. Lo mismo vale para la PWA: sus `PUBLIC_API_BASE` y `PUBLIC_CMS_BASE` apuntan a `localhost`, no a los nombres internos.

## Si algo falla

- **Falta una variable:** `docker compose` se detiene con un mensaje que dice cuál; ejecuta `./compose-env.sh`.
- **No puedo iniciar sesión en el panel del CMS:** el CMS corre en modo producción sobre HTTP plano. El login por API se probó, pero no el panel con un navegador. Si el panel vuelve al formulario sin error, agrega `ADMIN_COOKIE_SECURE=false` a `.env` y reconstruye (`docker compose up --build cms`): el panel lee esa opción al compilarse. Es solo para local.
- **`iluvatar` no arranca con `could not find default credentials`:** estás usando un proyecto real sin montar la cuenta de servicio, o el emulador sin `FIREBASE_AUTH_EMULATOR_HOST`.
- **La PWA muestra las imágenes rotas:** revisa que `CMS_PORT` coincida en `.env` (se usa en dos servicios).
- **`pwa` se queda esperando:** depende de que `iluvatar` esté saludable, y este de que `cms-seed` termine. `docker compose logs cms-seed` muestra el motivo.
- **Con Podman:** Podman de macOS no ejecuta los healthchecks, y con ellos `depends_on: service_healthy` no avanza. Usa Docker, o levanta los servicios por etapas.

## Qué está verificado

Verificado:
- La imagen de la PWA (su propio `Dockerfile` en `manwe-pwa`) se construyó con Podman y sirve la página, `sw.js`, el manifiesto, los íconos y las fuentes. Lo comprobó la sesión que mantiene la PWA y lo reportó; yo revisé el `Dockerfile`, no la construí de nuevo.
- La PWA conecta al emulador solo cuando `PUBLIC_FIREBASE_AUTH_EMULATOR_HOST` está definida (13 tests nuevos en `manwe-pwa`).
- El archivo es válido: `docker compose config` lo resuelve sin avisos, con y sin el perfil `emulator`, y falla con un mensaje claro si falta un secreto.
- El orden de arranque, las variables resueltas de `iluvatar` y los argumentos de build de la PWA salen como se espera.
- `npm run seed` con `SEED_TOKEN_FILE`: se probó contra SQLite en los cuatro casos (primera vez, repetición, archivo perdido con la base intacta, y sin la variable).
- `iluvatar`: leer el token desde archivo y separar la URL pública del CMS tienen tests.
- Las imágenes de `iluvatar` y del CMS se construyeron y ejecutaron con Podman en la fase anterior, y el CMS contra PostgreSQL 17.

**No verificado:**
- **Nunca se ejecutó `docker compose up`.** Es probable que aparezca algún ajuste al correrlo por primera vez.
- El CMS se modificó después de construirlo la última vez (carpeta `/shared` y el argumento de cookies) y no se reconstruyó.
- Que `cms-seed` funcione dentro de la imagen de producción del CMS (se probó en el repo, con todas las dependencias).
- El panel del CMS con un navegador sobre HTTP, y el login de la PWA contra Firebase real **o contra el emulador** (no se probó de punta a punta: el botón de Google, la sesión al recargar y que `iluvatar` acepte el token de la PWA).
- Que el service worker se registre en un navegador (solo se comprobó que `sw.js` y el manifiesto se sirven).

## Pendientes (fuera del compose)

- Probar de punta a punta el inicio de sesión de la PWA con el emulador, en un navegador completo.
- 