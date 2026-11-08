# Kümelen completo con Docker Compose

`compose.yaml` levanta en tu máquina la API/BFF, Laurelin (registro de estudiantes, con su base) y la PWA. El **CMS (Strapi) no está en el compose**: pide mucha memoria al construirse, así que se ejecuta aparte, directamente en el equipo (ver "El CMS, aparte"). Es para **desarrollo y pruebas locales**, no para producción (puertos en `localhost`, datos de prueba, sin TLS).

> **Estado: escrito y validado, no ejecutado.** Ver "Qué está verificado" al final.

## Qué levanta

| Servicio | Qué es | Dirección |
|---|---|---|
| `iluvatar` | API + BFF (Go) | http://localhost:8080 (`/swagger/index.html`) |
| `laurelin-db` | PostgreSQL 17 propio de Laurelin (volumen `laurelin_pgdata`) | solo red interna |
| `laurelin` | Registro de estudiantes (Go), solo API | http://localhost:8090 |
| `pwa` | La app web (`manwe-pwa`, SvelteKit) | http://localhost:3000 |
| `firebase-emulator` | Emulador de Firebase Auth, **solo con** `--profile emulator` | http://localhost:9099 |
| CMS (fuera del compose) | Strapi 5 con SQLite, `make cms-dev` | http://localhost:1337 (panel en `/admin`) |

Orden de arranque (lo resuelve Compose): `laurelin-db` → `laurelin` → `iluvatar` → `pwa`. `iluvatar` consulta al CMS en `http://host.docker.internal:1337`, así que el CMS debe estar corriendo en el equipo para que las pantallas respondan (sin él, `/ready` de `iluvatar` da 503 y `/v1/home` da 502).

## Requisitos

- Docker con Compose v2 (usa `depends_on` con condiciones y perfiles). Validado con Compose 5.6.
- Estar dentro del monorepo: el compose construye `../services/iluvatar`, `../services/laurelin` y `../apps/manwe-pwa`. Todos los comandos se ejecutan desde `deploy/` (o con `make` desde la raíz).
- Node 24 para el CMS (con nvm: `nvm use 24`; el Node por defecto no sirve para Strapi 5).
- Puertos libres: 1337 (CMS), 8080, 8090, 3000 (y 9099 con el emulador). Se pueden cambiar con `CMS_PORT`, `ILUVATAR_PORT`, `LAURELIN_PORT`, `PWA_PORT` en `.env`.

## Primer arranque

Desde la raíz del monorepo (el `Makefile` encadena estos pasos; `make help` lista todo):

```bash
make env            # deploy/.env con secretos aleatorios (no sobrescribe uno existente)
make cms-install    # dependencias del CMS (npm ci)
make cms-env        # services/iluvatar-cms/.env (SQLite, secretos, webhook hacia iluvatar)
make cms-seed       # datos de prueba + API token en deploy/secrets/cms_token (con el CMS detenido)
make up             # iluvatar, laurelin y pwa con el emulador de Auth (PROFILE= vacío: sin emulador)
make cms-dev        # en otra terminal: el CMS en http://localhost:1337/admin
```

Sin `make`, desde `deploy/`: `./compose-env.sh`, `./cms-env.sh`, `docker compose --profile emulator up --build` (o `docker-compose` si no tienes el plugin), y los comandos `npm` del CMS que están en el `Makefile`.

La primera vez construye dos imágenes de Go y una de la PWA. Después:

1. Abre http://localhost:1337/admin y crea la primera cuenta de administración del CMS.
2. Los datos de prueba (3 servicios, 7 publicaciones, categorías y avisos) ya están cargados por `make cms-seed`.
3. Abre http://localhost:3000 para la PWA.

## El CMS, aparte

El CMS usa **SQLite** (`services/iluvatar-cms/.tmp/data.db`) y las imágenes subidas quedan en `services/iluvatar-cms/public/uploads`; nada de eso está en volúmenes de Docker. `make cms-dev` lo ejecuta con recarga en caliente; `make cms-build` y `make cms-start` lo compilan y ejecutan sin ella (la compilación del panel es lo que consume memoria).

- **Token:** `make cms-seed` crea el API token de solo lectura y lo escribe en `deploy/secrets/cms_token`; el compose monta esa carpeta en `iluvatar` (`ILUVATAR_TOKEN_FILE=/secrets/cms_token`). Ejecútalo con el CMS detenido (SQLite, un solo proceso) y **antes** de `make up`. Si borras el archivo con la base intacta, el seed lo regenera.
- **Webhook:** al publicar, el CMS avisa a `http://localhost:8080/v1/hooks/post` con el `WEBHOOK_SECRET` de `deploy/.env` (`make cms-env` lo copia, y deben coincidir). Con `PUSH_DISABLED=true` no se envía ningún push.
- **Otra dirección del CMS:** si el CMS corre en otro equipo o puerto, define `CMS_URL` en `deploy/.env` (por defecto `http://host.docker.internal:1337`).

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
docker compose down -v               # detiene y BORRA la base de Laurelin
```

Los datos del CMS no los toca el compose: para empezar de cero borra `services/iluvatar-cms/.tmp/` y `public/uploads/*` y vuelve a ejecutar `make cms-seed`.

## Laurelin en el compose

`laurelin` no tiene interfaz: se opera por API. Para probarlo con el emulador de Auth:

1. Crea en el emulador una cuenta con el correo de `LAURELIN_BOOTSTRAP_ADMIN_EMAIL` (por defecto `admin@laurelin.test`) y márcala como verificada. El emulador acepta `POST http://localhost:9099/identitytoolkit.googleapis.com/v1/accounts:signUp?key=fake` y luego `POST .../v1/projects/demo-kumelen/accounts:update` con `Authorization: Bearer owner` y `{"localId": "...", "emailVerified": true}`. Vuelve a iniciar sesión (`accounts:signInWithPassword`) para obtener un ID token con el correo verificado.
2. Llama a `http://localhost:8090/v1/me` con `Authorization: Bearer <ID token>`. El resto de las rutas están en `services/laurelin/CLAUDE.md`.
3. La consulta que hace `iluvatar` se prueba con `POST /internal/v1/students/lookup` y `Authorization: Bearer $LAURELIN_SERVICE_TOKEN` (el valor está en `.env`).

### Exigir que quien entra sea estudiante

Por defecto `iluvatar` **no** consulta a Laurelin: entra cualquier cuenta de Firebase con token válido (así funciona la PWA con el emulador sin cargar datos). Para exigir que sea un estudiante activo:

1. Con el administrador de arriba, crea una institución, un programa y el estudiante cuyo correo usarás para entrar (`POST /v1/institutions`, `/v1/institutions/:id/programs` y `/v1/students`).
2. En `.env` agrega `ILUVATAR_LAURELIN_URL=http://laurelin:8090` y reinicia `iluvatar` (`docker compose up -d iluvatar`).
3. Desde ese momento las rutas `/v1/*` de `iluvatar` responden 403 si la cuenta no tiene el correo verificado o no es un estudiante activo, y 503 si Laurelin no responde y esa persona no había sido verificada antes. Las respuestas se guardan 10 minutos (1 minuto si no es estudiante) y, si Laurelin cae, quien ya fue verificado sigue entrando hasta 24 horas.

El botón de Google de la PWA con el emulador crea cuentas con el correo sin verificar: contra Laurelin necesitan un proyecto real de Firebase (opción B) o cuentas verificadas en el emulador.

Con un proyecto real de Firebase, `LAURELIN_BOOTSTRAP_ADMIN_EMAIL` debe ser tu correo real (verificado por Google al iniciar sesión).

## Por qué hay dos direcciones para el CMS

`iluvatar` consulta al CMS desde el contenedor en `http://host.docker.internal:1337` (el equipo anfitrión), pero las imágenes que devuelve las carga **el navegador**, que usa `localhost`. Por eso `iluvatar` recibe `CMS_PUBLIC_URL=http://localhost:1337` para armar las URLs de imágenes. Lo mismo vale para la PWA: sus `PUBLIC_API_BASE` y `PUBLIC_CMS_BASE` apuntan a `localhost`.

## Si algo falla

- **Falta una variable:** `docker compose` se detiene con un mensaje que dice cuál; ejecuta `./compose-env.sh`.
- **`iluvatar` no arranca y habla del token:** falta `deploy/secrets/cms_token`; ejecuta `make cms-seed` (con el CMS detenido) y luego `make up`.
- **`/v1/home` da 502 o `/ready` da 503:** el CMS no está corriendo en el equipo (`make cms-dev`) o el contenedor no lo alcanza. Prueba desde el contenedor: `docker compose exec iluvatar wget -qO- http://host.docker.internal:1337/_health`. Si `host.docker.internal` no resuelve (Docker en Linux), define `CMS_URL` con la IP del equipo.
- **El CMS no deja iniciar sesión en el panel:** con `make cms-dev` corre en desarrollo y no debería pasar. Con `make cms-start` (producción sobre HTTP plano) agrega `ADMIN_COOKIE_SECURE=false` al `.env` del CMS y vuelve a ejecutar `make cms-build`.
- **`iluvatar` no arranca con `could not find default credentials`:** estás usando un proyecto real sin montar la cuenta de servicio, o el emulador sin `FIREBASE_AUTH_EMULATOR_HOST`.
- **La PWA muestra las imágenes rotas:** revisa que `CMS_PORT` coincida en `.env` y en el `PORT` del `.env` del CMS.
- **`pwa` se queda esperando:** depende de que `iluvatar` esté saludable, y este de que `laurelin` lo esté. `docker compose logs laurelin` muestra el motivo.
- **Con Podman:** Podman de macOS no ejecuta los healthchecks, y con ellos `depends_on: service_healthy` no avanza. Usa Docker, o levanta los servicios por etapas.

## Qué está verificado

Verificado:
- El compose **completo (con CMS dentro) se ejecutó** en Colima (4 CPU, 6 GiB): los 8 contenedores quedaron sanos, `iluvatar` respondió con datos del CMS y la PWA abrió la ventana del emulador al iniciar sesión.
- Tras sacar el CMS del compose, `docker compose config` valida con y sin el perfil `emulator` y los servicios son `iluvatar`, `laurelin-db`, `laurelin`, `pwa` y `firebase-emulator`.
- `npm run seed` con `SEED_TOKEN_FILE` (SQLite): primera vez, repetición, archivo perdido con la base intacta y sin la variable. `iluvatar` lee el token desde archivo.

**No verificado (nuevo arreglo, no ejecutado):**
- Que el CMS en el equipo (`make cms-dev`) se alcance desde el contenedor de `iluvatar` por `host.docker.internal` en Colima.
- `make cms-env`, `make cms-seed` y `make cms-dev` de punta a punta, y que el webhook del CMS llegue a `iluvatar` en `localhost:8080`.
- El login de la PWA hasta la pantalla de inicio, y que el service worker se registre en un navegador.

## Pendientes (fuera del compose)

- Probar de punta a punta el inicio de sesión de la PWA con el emulador, en un navegador completo.
