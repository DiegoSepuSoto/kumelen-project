# iluvatar — API y BFF de Kümelen

Servicio Go (Echo) sin base de datos propia. Verifica ID tokens de Firebase Auth (y, con `LAURELIN_URL`, que la cuenta sea de un estudiante activo según Laurelin), sirve las pantallas de la app (BFF, antes el repo `elentari`, leídas del CMS Strapi 5 por REST), suscribe dispositivos a topics de notificaciones y envía el push por FCM v1 cuando el CMS publica un post. El repo `elentari` queda obsoleto.

## Comandos

```bash
go build ./...
go vet ./...
go test -race -cover ./...
go test -tags e2e -run TestBFFAgainstRealCMS ./src/   # Strapi real (`npm run seed`) + emulador de Firebase Auth; ver el comentario del archivo para las variables
go test -tags firebase_emulator -run TestEmulator ./src/infrastructure/firebase/   # verificador real del SDK contra el emulador (FIREBASE_AUTH_EMULATOR_HOST)
go test -tags record -run TestRecord ./src/infrastructure/strapi/   # regraba los fixtures desde un CMS real (STRAPI_URL, STRAPI_TOKEN)
swag init -g src/main.go          # regenera docs/ (Swagger); probado con swag v1.16.6
govulncheck ./...                 # 3 vulnerabilidades alcanzables: 2 de Echo sin parche publicado y 1 de la biblioteca estándar (corregida en Go 1.27.3)
```

Swagger en `/swagger/index.html`. Variables de entorno: ver `README.md` (`FIREBASE_PROJECT_ID` es obligatoria; el servicio no arranca sin credenciales de FCM).

Emulador de Auth: `npx firebase-tools emulators:start --only auth --project demo-kumelen` desde una carpeta con un `firebase.json` que habilite solo `auth`. Con Node 24 funciona; con el Node por defecto (16) no.

## Arquitectura (`src/`)

- `main.go`: composición de dependencias a mano (`run`); `registerBFFHandlers` monta las pantallas. Falla al arrancar si no puede crear los clientes de Firebase o si la configuración es insegura (`PUSH_DISABLED` en producción). `iluvatar healthcheck` consulta `/health` y devuelve el código de salida (lo usa el `HEALTHCHECK` de la imagen).
- `domain/`: modelos (`models/`) e **interfaces** de repositorio (`repository/`). No importa de otras capas.
- `application/usecase/`: una carpeta por área, una operación por archivo.
  - Pantallas de la app: `home`, `post`, `service`, `catalog`, `category`, `notification` (leer notificaciones).
  - Notificaciones push: `push` (aviso de un post publicado y suscripción de dispositivos a topics). No confundir con `notification`.
- `infrastructure/`:
  - `strapi/` (`WithPublicURL` separa la dirección interna del CMS de la que ve el navegador): cliente REST de Strapi 5 (con caché que sirve la última respuesta buena si el CMS cae, y `Ping` para `/ready`) y los repositorios de servicios, posts, categorías y avisos. Los identificadores son los `documentId` del CMS.
  - `laurelin/`: cliente de `POST /internal/v1/students/lookup` de `services/laurelin` (implementa `repository.StudentDirectory`). Caché en memoria por hash del correo (estudiante 10 min, no estudiante 1 min), una sola consulta por correo a la vez y, si Laurelin cae, acepta hasta 24 h a quien ya fue verificado como estudiante; a cualquier otro lo rechaza con 503 (falla cerrado).
  - `firebase/`: verificador de ID tokens (`auth_verifier.go`) y repositorio de push FCM v1 (`push_repository.go`) sobre el SDK Admin. Ambos dependen de una interfaz mínima del cliente del SDK para poder probarlos.
  - `http/server/`: configuración de producción (`ConfigFromEnv`, `Apply`): CORS, cabeceras de seguridad, límite de cuerpo, rate limit por IP, timeouts, ID de petición y log JSON (`slog`) sin datos personales (usa la ruta registrada, no la URL).
  - `http/handlers/rest/`: handlers Echo y el middleware `RequireAuth`, que deja el `uid` de Firebase en el contexto (`middleware.UID(c)`).
- `shared/`: `validations`, `testdoubles` (dobles de prueba a mano).

## Rutas

- Autenticación: no hay login en este servicio. El cliente inicia sesión con Firebase y envía `Authorization: Bearer <ID token>`; cualquier fallo es 401. Con `LAURELIN_URL` se aplica además `RequireStudent`: 403 si el correo no está verificado o no es de un estudiante activo, 503 si Laurelin no responde (el correo nunca va al log ni a la respuesta). Sin `LAURELIN_URL` la verificación está desactivada, y `ENV=production` se niega a arrancar así salvo `STUDENT_CHECK_DISABLED=true`. El webhook del CMS no pasa por esto.
- `GET /v1/home`, `/v1/post/:id`, `/v1/post/search/:searchTerm`, `/v1/service/:id`, `/v1/service/:id/posts`, `/v1/catalog`, `/v1/category/:id/posts`, `/v1/notification`, todas con `RequireAuth`.
- `POST /v1/push/subscribe` y `POST /v1/push/unsubscribe` (con `RequireAuth`): cuerpo `{"device_token": "...", "topics": ["all", "service.<documentId>"]}`. Los topics se validan (`all` o `service.` + id alfanumérico, máximo 20). El servidor no guarda el token.
- `POST /v1/hooks/post` (webhook `entry.publish` de Strapi): exige la cabecera `X-Webhook-Secret` igual a `WEBHOOK_SECRET` (401 si no coincide, 503 si el servicio no la tiene configurada). Ignora con 200 los eventos que no son la publicación de un post. Envía **un solo mensaje FCM** con la condición `'all' in topics || 'service.<id>' in topics`, así un dispositivo suscrito a ambos lo recibe una vez.
- `GET /health` (vive) y `GET /ready` (el CMS responde; 503 si no, sin detalle al cliente). `GET /swagger/*` solo si `SWAGGER_ENABLED`.
- Todas las demás rutas: máximo 1 MB de cuerpo y rate limit por IP (429).

## Tests

- Dobles de prueba en `src/shared/testdoubles` (campos función; llamar a uno sin configurar provoca un pánico explícito).
- `handlers/rest/v1/routes_test.go` fija el contrato REST de las pantallas: rutas, 401, 502 si el CMS falla y la forma de `/v1/home`.
- `middleware` (incluye `RequireStudent`), `laurelin` (cliente con servidor falso: caché, caída, concurrencia), `push` (casos de uso y handler) y `hooks` tienen tests propios; el webhook usa la carga real de Strapi (`hooks/testdata`).
- `infrastructure/strapi` se prueba con fixtures de respuestas reales de Strapi 5.61 (`testdata/`) y un CMS falso con `httptest`.
- `infrastructure/firebase` se prueba con un cliente del SDK falso. Lo que **no** está verificado: la verificación de firma de tokens y el envío FCM reales (requieren un proyecto de Firebase); el emulador de Auth no firma los tokens y no existe emulador de FCM.
- `http/server` prueba CORS, cabeceras, rate limit (incluido que un `X-Forwarded-For` falso no lo evade), límite de cuerpo y que el log no filtra datos personales; `strapi/cache_test.go` prueba la caché con reloj falso.
- Pruebas con etiqueta (`e2e`, `firebase_emulator`, `record`) necesitan servicios locales y no corren en `go test ./...`.
- Verificado a mano con el binario y la imagen (podman): CORS, 429, `/ready` 503 con el CMS caído mientras `/v1/home` responde desde la caché, apagado ordenado con SIGTERM, usuario no root. El runtime de este Podman no ejecuta los healthchecks, pero el comando funciona dentro del contenedor.

## Deuda y riesgos conocidos (no arreglar sin pedirlo)

- Sin métricas ni alertas: solo logs JSON y las sondas `/health` y `/ready`.
- La caché y el rate limit viven en memoria de cada instancia: con varias réplicas no se comparten (el límite efectivo es por réplica).
- El BFF responde 502 también cuando un recurso no existe en el CMS (el CMS contesta 404).
- `docs/` (Swagger) se regenera a mano con `swag init`.

## Al agregar código

- Nuevo caso de uso: interfaz en `domain/repository`, implementación en `infrastructure`, caso de uso en `application/usecase`, cableado en `main.go`.
- Tests: tabla de casos por función, con los dobles de `testdoubles`. Los repositorios que llaman a servicios externos se prueban con `httptest`.
- No loguear correos, contraseñas ni tokens.
