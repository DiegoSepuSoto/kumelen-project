# Kümelen

App para que los servicios estudiantiles de la UTEM publiquen contenido y notificaciones a los estudiantes. Trabajo de título de Diego Sepúlveda. Este es un monorepo (commit inicial: nov 2026, copiado desde los repos independientes anteriores, sin su historial). El plan de migración y salida a producción está en `docs/PLAN_MIGRACION.md` (léelo antes de planificar cambios grandes).

## Mapa del monorepo

| Carpeta | Rol | Stack |
|---|---|---|
| `services/iluvatar` | API + BFF: verifica ID tokens de Firebase, sirve las pantallas, suscribe dispositivos y envía push | Go 1.27, Echo, Firebase Admin SDK |
| `services/laurelin` | Registro de estudiantes chilenos (pregrado y postgrado, varias instituciones); responde si un correo es de un estudiante activo | Go 1.27, Echo, PostgreSQL |
| `services/iluvatar-cms` | CMS headless para los servicios estudiantiles | Strapi 5, Node 22 a 26, SQLite (dev) / PostgreSQL (prod) |
| `apps/manwe-pwa` | App de los estudiantes (reemplaza a la app Flutter) | SvelteKit, PWA |
| `deploy` | `compose.yaml` local, `compose-env.sh`, config del emulador de Firebase | Docker Compose |
| `docs` | `PLAN_MIGRACION.md` y diagramas de la arquitectura actual (`docs/architecture`, `make diagrams`) | Markdown, Archify |

Fuera del monorepo (repos antiguos, archivados o por archivar): `kumelen-load-tests` (Artillery, excluido por decisión), `elentari` (BFF fusionado en `iluvatar`), `manwe` (Flutter) y el informe de la tesis (LaTeX, describe el diseño **original**). Cada carpeta de servicio tiene su propio `CLAUDE.md` con comandos y límites. Cada servicio se construye y prueba por separado: no hay workspaces de npm ni herramienta de monorepo; por ahora el monorepo no tiene CI (se agregará más adelante).

## Levantar todo

`make help` (raíz) lista los comandos de trabajo. `deploy/COMPOSE.md` explica `deploy/compose.yaml` (API/BFF, Laurelin con su PostgreSQL y PWA, más el emulador de Auth con `--profile emulator`); `deploy/compose-env.sh` genera su `.env`. **El CMS no está en el compose** (consume mucha memoria al construirse): se ejecuta aparte, en el equipo, con SQLite (`make cms-install`, `cms-env`, `cms-seed`, `cms-dev`), y `iluvatar` lo alcanza por `host.docker.internal`. Se ejecuta desde `deploy/`. El compose con el CMS adentro sí se ejecutó (Colima, 4 CPU y 6 GiB); sin él, está validado con `config` pero no ejecutado.

## Flujo de datos

```
cliente --Firebase Auth--> ID token
cliente --REST + Bearer ID token--> iluvatar --REST + API token--> iluvatar-cms
iluvatar-cms --webhook entry.publish (X-Webhook-Secret)--> iluvatar --FCM v1 (topics)--> dispositivos
cliente --POST /v1/push/subscribe--> iluvatar --> FCM (suscribe el token a los topics)
iluvatar --POST /internal/v1/students/lookup (token de servicio)--> laurelin   [solo si LAURELIN_URL está definida]
administradores --Firebase Auth + /v1/*--> laurelin
```

## Estado (nov 2026)

- Fases 0 a 4 del plan hechas; el monorepo tiene un solo commit inicial en `main`, sin push. Cada repo tiene sus comandos y límites en su propio `CLAUDE.md`.
- Verificado en local: PostgreSQL 17 en el CMS, imágenes de contenedor de ambos servicios (Podman), caché ante caída del CMS, CORS, rate limit y logs.
- No verificado: el envío FCM y la verificación de firma de tokens contra un proyecto de Firebase real; la carga en un entorno real; no hay CI todavía.
- Laurelin (`services/laurelin`) tiene el servicio hecho y probado (casos de uso, HTTP, PostgreSQL real, emulador de Auth, imagen), en la rama `feat/laurelin`; ya está en `deploy/compose.yaml` (validado, no ejecutado). `iluvatar` lo consulta (opcional, con `LAURELIN_URL`) para dejar entrar solo a estudiantes activos. Falta una interfaz de administración.
- Pendiente de decisión: hosting, dominio y TLS; almacenamiento de imágenes del CMS; PostgreSQL gestionado con respaldos.
- Toolchain local: Go 1.27, Node 24 vía nvm (el Node por defecto es 16 y no sirve para Strapi 5 ni para firebase-tools) y Colima como runtime de Docker (`make colima-up` / `make colima-down`; 4 CPU y 6 GiB bastan para construir todo; el compose usa `docker-compose` si falta el plugin `docker compose`).
- Los secretos van por variables de entorno; nunca escribir valores reales en archivos versionados.

## Reglas de trabajo

- Idioma: código y nombres en inglés, comentarios y docs en español, igual que el código existente.
- Arquitectura por capas en `iluvatar`; no mezclar capas (ver su `CLAUDE.md`).
- Los datos de estudiantes se tratan como sensibles: no loguear correos, tokens ni identificadores personales.
- Cada cambio de comportamiento lleva su test. Antes de decir que algo funciona, ejecutar `go build ./... && go vet ./... && go test -race ./...` y reportar el resultado real, incluyendo lo que no se pudo verificar.
- No cambiar dependencias ni la versión de Go sin avisar.
