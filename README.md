# Kümelen

Plataforma para que los servicios estudiantiles de la UTEM publiquen contenido y notificaciones a los estudiantes. Es el trabajo de título de Diego Sepúlveda.

Este monorepo reúne el backend (API/BFF y registro de estudiantes), el CMS, la app de los estudiantes (PWA) y el entorno local con Docker Compose. Se creó en nov 2026 copiando los repos independientes anteriores, sin su historial.

> **Estado:** desarrollo local. Nada ha estado en producción y no hay CI todavía. Ver [Estado y límites](#estado-y-límites).

## Arquitectura

```
cliente --Firebase Auth--> ID token
cliente --REST + Bearer ID token--> iluvatar --REST + API token--> iluvatar-cms
iluvatar-cms --webhook entry.publish (X-Webhook-Secret)--> iluvatar --FCM v1 (topics)--> dispositivos
cliente --POST /v1/push/subscribe--> iluvatar --> FCM (suscribe el token a los topics)
iluvatar --POST /internal/v1/students/lookup (token de servicio)--> laurelin   [solo si LAURELIN_URL está definida]
administradores --Firebase Auth + /v1/*--> laurelin
```

| Carpeta | Rol | Stack |
|---|---|---|
| [`services/iluvatar`](services/iluvatar) | API + BFF: verifica ID tokens de Firebase, sirve las pantallas leyendo el CMS, suscribe dispositivos y envía push | Go 1.27, Echo, Firebase Admin SDK |
| [`services/laurelin`](services/laurelin) | Registro de estudiantes chilenos (varias instituciones, pregrado y postgrado); responde si un correo es de un estudiante activo | Go 1.27, Echo, PostgreSQL |
| [`services/iluvatar-cms`](services/iluvatar-cms) | CMS headless donde los servicios estudiantiles publican | Strapi 5, Node 22 a 26, SQLite (dev) / PostgreSQL (prod) |
| [`apps/manwe-pwa`](apps/manwe-pwa) | App de los estudiantes (reemplaza a la app Flutter `manwe`) | SvelteKit 3, Svelte 5, PWA |
| [`deploy`](deploy) | `compose.yaml` local, generadores de `.env` y config del emulador de Firebase | Docker Compose |
| [`docs`](docs) | [Plan de migración](docs/PLAN_MIGRACION.md) y diagramas de la arquitectura actual (`make diagrams`) | Markdown, Archify |

Cada servicio tiene su propio `CLAUDE.md` (comandos, arquitectura, límites) y, salvo Laurelin, un `README.md`. No hay workspaces de npm ni herramienta de monorepo: cada pieza se construye y prueba por separado.

Fuera del monorepo quedan `elentari` (el BFF antiguo, fusionado en `iluvatar`), `manwe` (Flutter), `kumelen-load-tests` (Artillery) y el informe de la tesis (LaTeX, describe el diseño original).

## Requisitos

- **Docker** (en macOS se usa [Colima](https://github.com/abiosoft/colima)) y Docker Compose v2. El `Makefile` usa `docker compose` o, si falta el plugin, `docker-compose`.
- **Go 1.27** para compilar y probar los servicios Go.
- **Node 24** para el CMS, la PWA y el emulador de Firebase. El `Makefile` lo activa con [nvm](https://github.com/nvm-sh/nvm) si está instalado; el Node por defecto de tu equipo probablemente no sirve para Strapi 5.
- Recursos de la VM de Docker: **4 CPU y 6 GiB** (`make colima-up`). El stack ya en ejecución usa ~1,2 GiB; los 6 GiB son para construir el panel de Strapi.

## Inicio rápido

Desde la raíz del monorepo (`make help` lista todos los comandos):

```bash
make colima-up      # solo macOS con Colima
make env            # deploy/.env con secretos aleatorios (no sobrescribe uno existente)
make cms-install    # dependencias del CMS
make cms-env        # services/iluvatar-cms/.env (SQLite, secretos, webhook hacia iluvatar)
make cms-seed       # datos de prueba + API token en deploy/secrets/cms_token (con el CMS detenido)
make up             # iluvatar, laurelin y PWA, con el emulador de Firebase Auth
make cms-dev        # en otra terminal: el CMS con recarga en caliente
```

El CMS **no corre en el compose** porque construirlo consume mucha memoria: se ejecuta aparte, en tu equipo, y `iluvatar` lo alcanza por `host.docker.internal`. `cms-seed` va antes de `make up` porque `iluvatar` necesita el archivo del token al arrancar.

| Servicio | Dirección |
|---|---|
| PWA | http://localhost:3000 |
| iluvatar | http://localhost:8080 (Swagger en `/swagger/index.html`) |
| Laurelin | http://localhost:8090 |
| CMS (panel) | http://localhost:1337/admin (la primera vez se crea la cuenta de administración) |
| Emulador de Auth | http://localhost:9099 (con `PROFILE=emulator`, el valor por defecto) |

Con el emulador, el botón "Ingresar con Google" de la PWA abre una ventana falsa donde eliges o creas una cuenta. Para usar un proyecto real de Firebase de pruebas (nunca el de producción), ver [`deploy/COMPOSE.md`](deploy/COMPOSE.md).

## Comandos del Makefile

| Grupo | Comandos |
|---|---|
| Entorno | `colima-up`, `colima-down`, `colima-status`, `stop-all` (baja el stack y Colima) |
| Stack | `env`, `up`, `up-fg`, `build`, `down`, `clean` (borra la base de Laurelin), `restart S=<svc>`, `ps`, `logs [S=<svc>]`, `shell S=<svc>`, `config`, `urls`, `health` |
| CMS | `cms-install`, `cms-env`, `cms-seed`, `cms-dev`, `cms-build`, `cms-start` |
| Código | `test`, `test-go`, `test-pwa`, `vet`, `fmt`, `lint-go`, `install-pwa` |

Variables: `PROFILE=` (vacío = sin emulador, usa Firebase real), `S=<servicio>`, `COLIMA_CPU`, `COLIMA_MEM`.

## Desarrollar

```bash
make test           # Go (build, vet, test -race) y PWA (vitest)
make test-go
make test-pwa
```

Antes de dar un cambio por bueno, en cada servicio Go: `go build ./... && go vet ./... && go test -race ./...`. En la PWA: `npm run check && npm test && npm run build`. Las pruebas con etiqueta (`e2e`, `firebase_emulator`, `integration`, `record`) necesitan servicios locales y no corren con `go test ./...`; se explican en el `CLAUDE.md` de cada servicio.

La PWA también corre sola contra un BFF simulado (`npm run mock` y `npm run dev` en `apps/manwe-pwa`), sin Docker.

### Reglas del proyecto

- Código y nombres en inglés; comentarios y documentación en español.
- Arquitectura por capas en los servicios Go (`domain` ← `application/usecase` ← `infrastructure`); no mezclar capas.
- Los datos de estudiantes son sensibles: no loguear correos, tokens ni identificadores personales. Laurelin no guarda RUT, teléfono ni dirección.
- Cada cambio de comportamiento lleva su test.
- Los secretos van por variables de entorno; nunca valores reales en archivos versionados (`.env` y `deploy/secrets/*` están en `.gitignore`).
- No cambiar dependencias ni la versión de Go sin avisar.

## Qué hace cada pieza

**iluvatar** no tiene base de datos. Todas las rutas `/v1/*` exigen `Authorization: Bearer <ID token de Firebase>`: pantallas (`home`, `post`, `service`, `catalog`, `category`, `notification`), suscripción de dispositivos (`/v1/push/subscribe` y `unsubscribe`) y el webhook del CMS (`/v1/hooks/post`, protegido con `X-Webhook-Secret`), que dispara **un** mensaje FCM por publicación. Con `LAURELIN_URL` solo deja entrar a estudiantes activos; en producción se niega a arrancar sin esa verificación (salvo `STUDENT_CHECK_DISABLED=true`). Si el CMS cae, sirve la última respuesta buena desde una caché en memoria. Inventario de datos personales: [`PRIVACIDAD.md`](services/iluvatar/PRIVACIDAD.md).

**laurelin** guarda `institutions` → `programs` → `students` (correo institucional, nombre, año de ingreso, estado) y `admins` (`super_admin` e `institution_admin`), con auditoría. La API `/v1/*` es para administradores autenticados con Firebase; `/internal/v1/students/lookup` es para otros servicios, con un token de servicio, y no devuelve correo ni nombre. **Aún no tiene interfaz de administración**: se opera por API. Detalle en [`services/laurelin/CLAUDE.md`](services/laurelin/CLAUDE.md).

**iluvatar-cms** define los tipos `servicio`, `persona`, `contacto`, `categoria`, `post` y `aviso`. Al arrancar registra el webhook hacia `iluvatar` si existe `ILUVATAR_WEBHOOK_URL`. `iluvatar` lo lee con un API token de solo lectura.

**manwe-pwa** es una copia visual 1:1 de la app Flutter, sin funcionalidades nuevas; solo cambia el inicio de sesión (de Pasaporte.UTEM a Firebase Auth). Renderiza en servidor, el token vive en el cliente y el Markdown del CMS siempre se sanea. Detalle en [`apps/manwe-pwa/README.md`](apps/manwe-pwa/README.md).

## Estado y límites

Fases 0 a 4 del [plan de migración](docs/PLAN_MIGRACION.md) hechas. Todo está en la rama `feat/laurelin` sobre `main`, sin push.

**Verificado**
- Tests de los servicios Go y de la PWA (37 pruebas con Vitest en la última corrida).
- El compose completo, con el CMS dentro, se ejecutó en Colima: 8 contenedores sanos, la API respondió con un token del emulador y el popup de login de la PWA abre el emulador.
- PostgreSQL 17 en el CMS y en Laurelin; imágenes de contenedor de los servicios; caída del CMS, CORS, rate limit y logs.

**No verificado**
- Envío FCM y verificación de firma de tokens contra un proyecto real de Firebase (el emulador no firma tokens y no existe emulador de FCM).
- Con el CMS fuera del compose: que el contenedor de `iluvatar` alcance el CMS por `host.docker.internal` en Colima, los targets `cms-*` de punta a punta, el webhook CMS → `iluvatar` y que el `Makefile` completo funcione (solo se revisó con `make -n`).
- Que el login en la PWA con el emulador llegue hasta la pantalla de inicio.
- Modo sin conexión (service worker) en un navegador completo, y Safari/lectores de pantalla.
- Cargas en un entorno real; los workflows de CI no existen.

**Pendiente**
- Notificaciones push en la PWA (el endpoint de suscripción ya existe en `iluvatar`; falta el cliente FCM web y su proyecto de Firebase).
- Interfaz de administración de Laurelin, carga masiva por CSV, lectura de la auditoría y revisión legal (Ley 19.628 y 21.719).
- Mensaje propio en la PWA para el 403 de "no eres estudiante activo".
- Decisión del contraste del verde primario de la PWA.
- Decisiones abiertas: hosting, dominio y TLS; almacenamiento de imágenes del CMS; PostgreSQL gestionado con respaldos.
- Archivar `elentari`, `manwe` y los demás repos antiguos; agregar CI.

## Documentación

- [`deploy/COMPOSE.md`](deploy/COMPOSE.md): entorno local, Firebase real vs. emulador, operación diaria y problemas frecuentes.
- [`docs/PLAN_MIGRACION.md`](docs/PLAN_MIGRACION.md): decisiones (D1 a D12), fases y estado.
- [`docs/architecture`](docs/architecture): diagramas de la arquitectura actual hechos con Archify (HTML interactivo, regenerables con `make diagrams`). Los C4 en PlantUML del diseño original ya no se mantienen.
- `CLAUDE.md` en la raíz y en cada servicio: guía de trabajo para el desarrollo con Claude Code.
