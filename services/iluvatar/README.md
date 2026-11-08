# API Ilúvatar - Proyecto Kümelen

## Descripción del Proyecto

"Kümelen" nace desde la necesidad de implementar, con ayuda de elementos tecnológicos, nuevos canales de comunicación
que le permitan a los estudiantes de la UTEM conocer e informarse sobre los servicios estudiantiles que pueden apoyar
en las dimensiones extraacadémicas de su vida universitaria.

## Descripción del artefacto

El artefacto "Ilúvatar" es la API y el Backend for Frontend del proyecto: verifica la identidad de quien consulta
(ID tokens de Firebase Auth), construye el contenido de cada pantalla a partir del CMS (antes el artefacto "Elentari"),
suscribe dispositivos a notificaciones y envía las notificaciones push (FCM) cuando el CMS publica un post. No tiene base de datos propia.

## Variables de entorno

Las variables de entorno necesarias para la ejecución del artefacto son:

| Variable | Uso |
|---|---|
| `PORT` | Puerto (8080 por defecto). |
| `ENV` | `production` activa los valores seguros: Swagger cerrado, HSTS, sin CORS si no hay orígenes y rechazo de `PUSH_DISABLED`. |
| `APP_VERSION` | Se muestra en `/health`. |
| `FIREBASE_PROJECT_ID` | ID del proyecto de Firebase. **Obligatoria.** |
| `GOOGLE_APPLICATION_CREDENTIALS` | Ruta al JSON de la cuenta de servicio. El SDK de Firebase la necesita para crear sus clientes y FCM para enviar. Sin ella el servicio no arranca. |
| `ILUVATAR_CMS_HOST` | URL del CMS Strapi, sin `/api`. |
| `ILUVATAR_TOKEN` | API token de solo lectura del CMS. |
| `ILUVATAR_TOKEN_FILE` | Alternativa a `ILUVATAR_TOKEN`: ruta de un archivo que contiene el token (secretos montados, o el que deja el seed del CMS en el compose). |
| `CMS_PUBLIC_URL` | Dirección del CMS que ve el navegador, para las URLs de imágenes, cuando difiere de `ILUVATAR_CMS_HOST` (por ejemplo, red interna de Docker). Por defecto es `ILUVATAR_CMS_HOST`. |
| `WEBHOOK_SECRET` | Secreto compartido con el CMS; sin él, el webhook responde 503. |
| `CORS_ALLOWED_ORIGINS` | Orígenes web permitidos, separados por coma. En producción, sin esto no se permite ninguno. |
| `RATE_LIMIT_RPS` | Peticiones por segundo por IP (20 por defecto; `0` lo desactiva). |
| `TRUST_PROXY_HEADERS` | `true` solo detrás de un proxy propio: confía en `X-Forwarded-For` y `X-Forwarded-Proto`. Si no, la IP real sería la del proxy. |
| `SWAGGER_ENABLED` | Publica `/swagger`; por defecto, solo fuera de producción. |
| `CMS_CACHE_TTL` | Segundos de frescura de la caché del CMS (30 por defecto; `0` la desactiva). Si el CMS falla se sirve la última respuesta buena, hasta 24 horas. |
| `PUSH_DISABLED` | `true` no envía notificaciones; solo para desarrollo sin credenciales de FCM. |
| `FIREBASE_AUTH_EMULATOR_HOST` | Para desarrollo: usa el emulador de Auth (no firma los tokens ni sirve para FCM). |

## Contenedor

```bash
docker build -t iluvatar .
docker run -p 8080:8080 -e FIREBASE_PROJECT_ID=... -e ILUVATAR_CMS_HOST=... -e ILUVATAR_TOKEN=... -e WEBHOOK_SECRET=... \
  -e GOOGLE_APPLICATION_CREDENTIALS=/secrets/firebase.json -v /ruta/firebase.json:/secrets/firebase.json:ro iluvatar
```

Imagen de unos 50 MB sobre `distroless/static`, sin shell y con usuario no root. El `HEALTHCHECK` ejecuta `/iluvatar healthcheck`. Con Podman, construir con `--format docker` para conservar el `HEALTHCHECK`.

`GET /health` indica que el proceso vive; `GET /ready` comprueba además que el CMS responde (503 si no): úsalos como sondas de vida y de disponibilidad.

## Entorno de compilación/ejecución

Para compilar el código es necesario configurar un ambiente de Go en su versión 1.27

La ejecución del mismo es totalmente agnóstica, dado que de la compilación es generado un archivo binario ejecutable

## Generación/Visualización de documentación

Para la generación/actualización de la documentación del proyecto se requiere ejecutar el siguiente comando:

```console
foo@bar:~$ swag init -g src/main.go
```

Luego, para su visualización, es necesario entrar a la siguiente ruta:

http://servidor.com/swagger/index.html

### Nombre en clave

El nombre en clave utilizado para este artefacto, "Ilúvatar", viene de la mitología creada por el autor J.R.R. Tolkien,
donde Ilúvatar representa al ente creador de todo el universo donde se llevan a cabo las aventuras de libros como "El
Hobbit" y "El Señor de los Anillos" [Ver Más](https://es.wikipedia.org/wiki/Il%C3%BAvatar)