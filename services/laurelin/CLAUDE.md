# laurelin — registro de estudiantes universitarios chilenos

Servicio Go (Echo + PostgreSQL) que guarda estudiantes de pregrado y postgrado de varias instituciones. Lo operan administradores (central o por institución) y le responde a otros servicios, sobre todo `iluvatar`, si un correo corresponde a un estudiante activo (reemplaza la verificación que daba Mi UTEM). Decisiones de diseño: `docs/PLAN_MIGRACION.md`, Fase L.

## Comandos

```bash
go build ./... && go vet ./... && go test -race -cover ./...
# PostgreSQL real (borra las tablas de Laurelin de esa base: usa una desechable)
podman run -d --rm --name laurelin-pg -e POSTGRES_PASSWORD=test -e POSTGRES_DB=laurelin -p 55433:5432 postgres:17
LAURELIN_TEST_DATABASE_URL=postgres://postgres:test@localhost:55433/laurelin go test -race -tags integration ./src/infrastructure/postgres/
podman build --format docker -t laurelin:dev .   # --format docker conserva el HEALTHCHECK
```

Emulador de Auth para probar a mano: igual que en `iluvatar` (`deploy/firebase/firebase.json`, Node 24). Los tokens del emulador traen `email_verified` solo si se marca la cuenta como verificada.

## Variables de entorno

| Variable | Uso |
|---|---|
| `DATABASE_URL` | Obligatoria. Conexión a PostgreSQL. Las migraciones se aplican solas al arrancar (con un candado, es seguro con varias réplicas). |
| `FIREBASE_PROJECT_ID` | Obligatoria. Con `FIREBASE_AUTH_EMULATOR_HOST` usa el emulador; sin él necesita credenciales (`GOOGLE_APPLICATION_CREDENTIALS`). |
| `SERVICE_TOKEN` o `SERVICE_TOKEN_FILE` | Token que presentan los otros servicios. Obligatorio con `ENV=production`; sin él `/internal/*` responde 503. |
| `BOOTSTRAP_ADMIN_EMAIL` | Crea (si no existe) una cuenta `super_admin` con ese correo. Es la única forma de tener el primer administrador. |
| `PORT` (8090), `ENV`, `CORS_ALLOWED_ORIGINS`, `RATE_LIMIT_RPS`, `TRUST_PROXY_HEADERS` | Igual que en `iluvatar`. No hay Swagger. |

## Modelo y reglas

- `institutions` → `programs` (nivel: `undergraduate`, `master`, `doctorate`, `other_postgraduate`) → `students`. La institución de un estudiante es la de su programa; puede cambiar de programa pero no de institución.
- **Un estudiante es solo**: correo institucional (minúsculas, único en todo Laurelin), nombre, año de ingreso y estado (`active`, `on_leave`, `graduated`, `withdrawn`). **No se guarda RUT, teléfono ni dirección** (minimización, Ley 21.719). No agregar datos personales sin una decisión explícita.
- Administradores (`admins`): `super_admin` (todo) e `institution_admin` (solo su institución). Se identifican por el **correo verificado** de su cuenta de Firebase; un token con correo sin verificar es 403.
- Alcance: lo decide la capa de casos de uso, no los handlers. Un recurso de otra institución responde 404 (no 403) en lecturas y cambios, para no revelar que existe; crear en otra institución es 403.
- `audit_log` guarda quién hizo qué (ids y acción, nunca correos ni nombres) y se escribe **después** del cambio: si falla, queda un error en el log y el cambio no se deshace.

## Rutas

- `/v1/*` con `Authorization: Bearer <ID token de Firebase>` (401 token inválido, 403 no es administrador): `GET /me`, CRUD de `/institutions` (crear, editar y borrar solo `super_admin`), `/institutions/:id/programs` y `/programs/:id`, `/students` (filtros `institution_id`, `program_id`, `level`, `status`, `q` por correo o nombre, `limit` máx. 200, `offset`) y `/admins` (solo `super_admin`; nadie puede borrarse a sí mismo). Borrar algo que aún se usa es 409.
- `POST /internal/v1/students/lookup` con `Authorization: Bearer <SERVICE_TOKEN>` y cuerpo `{"email": "..."}`. Responde `is_student` (true solo si `active`), `status`, `entry_year`, `level`, institución y programa, **sin devolver correo ni nombre**; 404 si el correo no existe. El correo va en el cuerpo para que no quede en logs de proxys.
- `GET /health` y `GET /ready` (PostgreSQL responde).

## Arquitectura (`src/`)

Igual que `iluvatar`: `domain` (modelos e interfaces, sin dependencias de otras capas) ← `application/usecase` (reglas) ← `infrastructure` (`postgres/`, `firebase/`, `http/`). `main.go` cablea a mano. `http/server` es la configuración de producción copiada de `iluvatar` (CORS, cabeceras, límites, rate limit por IP, logs JSON sin datos personales). No hay módulo compartido entre servicios: lo repetido es a propósito.

## Tests

- `shared/testdoubles/memory.go`: repositorios en memoria con las reglas que importan (correo y código únicos, no borrar lo que se usa). Los casos de uso y las rutas HTTP completas (`v1/routes_test.go`) se prueban sobre ellos.
- `postgres/integration_test.go` (etiqueta `integration`): migraciones concurrentes, filtros, comodines `%` y `_` en la búsqueda, claves foráneas y restricciones de la base.
- Verificado a mano: binario real + PostgreSQL 17 + emulador de Auth (401/403, correo sin verificar, CRUD, lookup, auditoría, logs sin correos ni tokens, cierre ordenado) y la imagen (usuario no root, `/ready`, `healthcheck`, rechazo de producción sin token). **No verificado**: la firma de tokens contra un proyecto de Firebase real.

## Deuda conocida (no arreglar sin pedirlo)

- `iluvatar` todavía **no** consulta a Laurelin: falta cliente, configuración y decidir qué pasa con un estudiante que no existe.
- No hay interfaz de administración, ni carga masiva (CSV), ni endpoint para leer `audit_log`. Está en `deploy/compose.yaml` (con su propia base `laurelin-db`), validado pero no ejecutado.
- No hay Swagger; las rutas están documentadas aquí.
- El rate limit vive en memoria de cada réplica. Falta revisión legal de qué se guarda y por cuánto tiempo (ver `docs/PLAN_MIGRACION.md`).
