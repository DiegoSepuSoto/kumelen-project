# Plan de migración y salida a producción — Kümelen

Estado de partida: ningún servicio ha estado en producción, no hay datos que conservar, ningún repo tiene tests y `iluvatar` / `elentari` no compilan (`httpclient-call-go@v1.1.1` inaccesible).

## Decisiones tomadas

| # | Decisión |
|---|---|
| D1 | Go 1.27 en `iluvatar` y `elentari` |
| D2 | Eliminar GraphQL: el BFF consumirá Strapi por REST |
| D3 | Cambiar de motor si hace falta: Strapi 5 con PostgreSQL (Strapi 5 no soporta MongoDB) |
| D4 | Se apunta a producción real, no a un MVP |
| D5 | Firebase Auth, sin restricción de dominio |
| D6 | La app Flutter (`manwe`) se trata aparte |
| D7 | Unir `elentari` (BFF) en `iluvatar`: queda un solo servicio Go llamado `iluvatar` (BFF + verificación de tokens Firebase + webhook a FCM). `elentari` se archiva |
| D8 | Dejar Flutter: `manwe` se reescribe como PWA con un framework JavaScript, copia visual 1:1, sin features nuevas (ver Fase F) |
| D9 | La PWA se construye con **SvelteKit** (TypeScript, `@vite-pwa/sveltekit`) |
| D11 | El look and feel se gestiona en **Penpot** autoalojado con `docker-compose`; los tokens viven en el código y el look actual de `manwe` se carga en Penpot con un plugin (ver Fase F) |
| D12 | Penpot se integrará con agentes de IA mediante MCP y los tokens se intercambiarán en formato DTCG; implementación posterior, tras una prueba corta con la instancia local (ver Fase F) |
| D10 | La PWA existe solo como web: no se publica en tiendas. Usa renderizado en servidor (SSR) de SvelteKit. Dominio y hosting se deciden más adelante |

Qué sobrevive de cada servicio en `iluvatar`: de `elentari`, las pantallas (`home`, `post`, `service`, `catalog`, `category`, `notification`) con su capa de dominio, casos de uso y handlers REST; de `iluvatar`, solo el webhook y el envío de notificaciones. Se eliminan el login Mi UTEM, el JWT propio, GORM/PostgreSQL de estudiantes, `ADMIN_ACCOUNT_*` y las rutas `validate-token`/`refresh-token` (el refresco lo hace Firebase en el cliente). El servicio queda sin base de datos propia; PostgreSQL es solo de Strapi.

Consecuencias de que no haya datos ni producción previa: no hay migración de datos (se crean los tipos de contenido en Strapi 5 y se siembran), no hay compatibilidad hacia atrás ni rollback que cuidar, y se puede cambiar la forma de las rutas si conviene.

## Hallazgos que condicionan el plan

- **Push roto:** `iluvatar` envía con la API legacy de FCM (`POST /fcm/send` con `key=`), que Google dio de baja. Además `SendNotificationsToUsers` lee `response.StatusCode` antes de revisar el error (pánico si la llamada falla) y hace un envío por dispositivo en un bucle.
- **Login:** `Login` hace `defer response.Body.Close()` sin revisar el error de `Do()`.
- **Sin dominio, sin verificación de estudiante:** cualquiera con Google/Apple/correo podrá entrar. Los datos (UID de Firebase + token de dispositivo) deben minimizarse; se pierde `Career`, que hoy viene de Mi UTEM.
- **El modelo de datos real de `iluvatar` es mínimo:** la tabla `STUDENT` solo tiene `ID`, `Email` (hash MD5), `Career`, `DeviceID`. No existen las "preferencias" que describe el informe.
- **Webhook:** `/v1/hooks/post` no tiene autenticación y no distingue el evento (solo el modelo).
- **Strapi:** el secreto de admin tiene un valor por defecto (`12345`) y el playground de GraphQL está siempre activo.

## Fase 0 — Base verificable

Solo sobre el código que sobrevive a la fusión (`elentari` y la parte de notificaciones de `iluvatar`); no se escriben tests del login Mi UTEM, el JWT propio ni el MD5, porque se eliminan.

1. Hacer que `elentari` compile: reemplazar `httpclient-call-go` por `net/http`.
2. Tests de caracterización (tabla de casos + mocks a mano de las interfaces de dominio) de los casos de uso de `elentari` (home, post, service, catalog, category, notification).
3. CI mínimo (GitHub Actions): `go build`, `go vet`, `go test`, `govulncheck`.
4. Registrar la línea base de cobertura.

Salida: el código del BFF compila en verde y tiene tests; es la red de seguridad de la fusión.

**Estado: hecha en `elentari`, rama `chore/fase-0-compilar-y-tests` (commit `4b8c806`).** `go build`, `go vet` y `go test -race` en verde; cobertura base 19,1 % del total (casos de uso y middleware al 100 %, cliente de auth 85,5 %, repositorios GraphQL y handlers sin cobertura propia; el contrato de rutas se cubre desde `handlers/rest/v1/routes_test.go`). `govulncheck`: 16 vulnerabilidades alcanzables (Echo 8, `x/net` 6, `golang-jwt/jwt` 1, biblioteca estándar 1, esta última se corrige con Go 1.27.3). Se corrigió además un pánico en el middleware JWT.

## Fase 1 — Go 1.27

1. **Fusión:** traer el código de `elentari` al repo `iluvatar` (módulo `iluvatar`), conservando la estructura por capas. Archivar el repo `elentari` con un aviso que apunte a `iluvatar`.
   - **Ajuste respecto al plan original:** el login Mi UTEM, el JWT propio, GORM y el push legacy **se conservan** en la Fase 1 y se eliminan en la Fase 3, cuando Firebase los reemplaza. Quitarlos antes dejaría el servicio sin autenticación ni notificaciones durante el intervalo. El BFF valida el JWT en el mismo proceso (se elimina la llamada HTTP `validate-token` entre servicios).
2. `go.mod` a `go 1.27` y actualizar dependencias:
   - Echo, validator, GORM y swag a versiones vigentes.
   - `dgrijalva/jwt-go` → `golang-jwt/jwt/v5`, con tests que fijan el comportamiento antes de migrar.
   - `ioutil` → `io`; cliente HTTP propio sobre `net/http` (`shared/httpclient`).
   - El cliente GraphQL se mantiene fijo en `v0.2.0` (la versión `v0.3.0` cambia la API) porque se elimina en la Fase 2.
3. `go vet`, tests y regeneración de Swagger.
4. Unificar los `CLAUDE.md` (hoy hay uno por servicio) en el de `iluvatar`.

**Estado: hecha en `iluvatar`, rama `feat/fase-1-fusion-bff-go-1-27` (sin commitear).** `go 1.27`; `go build`, `go vet` y `go test -race` en verde. Cobertura total 13,9 % (baja respecto a la Fase 0 porque se sumó código sin tests: login, base de datos, push). `govulncheck` pasó de 16 a 3 vulnerabilidades alcanzables: 2 de Echo sin parche publicado y 1 de la biblioteca estándar (se corrige con Go 1.27.3; aquí hay 1.27.2). Pendiente: archivar el repo `elentari` en GitHub (acción tuya, no la hice).

## Fase 2 — Strapi 5 + PostgreSQL + REST (en paralelo con la Fase 3)

1. Proyecto nuevo de Strapi 5 con PostgreSQL; Node LTS soportado por Strapi 5 (verificar la matriz vigente).
2. Recrear los tipos de contenido desde los `*.settings.json` actuales: `servicio`, `persona`, `contacto`, `categoria`, `post`, `aviso`. Script de siembra para datos de prueba.
3. Reescribir los repositorios de `elentari` que hoy usan GraphQL (≈660 líneas en `infrastructure/graphql/repository/iluvatar/*`) como clientes REST de Strapi (`populate`, filtros, paginación), y quitar la dependencia del cliente GraphQL.
4. Webhook Strapi → `iluvatar`: nuevo formato, filtrado por evento de publicación, firmado con un secreto compartido.
5. Hardening: secretos por variable de entorno (sin valores por defecto), permisos públicos mínimos (solo lectura de lo publicado), almacenamiento de imágenes fuera del disco del contenedor (S3 o similar), roles para los equipos de servicios estudiantiles.

**Estado: hecha, salvo lo indicado.** `iluvatar-cms` rama `feat/fase-2-strapi-5` (commit `a565e52`) e `iluvatar` rama `feat/fase-2-rest-strapi` (commit `00672e5`), sin push.
- Hecho y verificado: Strapi 5.61 con los 6 tipos y datos de prueba; BFF por REST probado contra un Strapi 5 real (prueba e2e, con las 8 rutas); webhook probado con la carga real que envía Strapi; GraphQL eliminado; secretos sin valores por defecto. Cobertura total de `iluvatar`: 41,5 %. `govulncheck`: 3 alcanzables (sin cambio).
- ~~Sin verificar: PostgreSQL~~ Verificado en la Fase 4 (Strapi 5.61 sobre PostgreSQL 17 con Podman).
- Pendiente: almacenamiento de imágenes fuera del disco (decisión abierta), roles de los equipos de servicios en el panel, CORS del CMS y CI del CMS (Fase 4).
- Cambios de comportamiento a tener en cuenta: los identificadores de la API son ahora `documentId` (cadenas); el webhook responde 200 y se ignora si no es la publicación de un post (antes 400); `/v1/hooks/post` exige `WEBHOOK_SECRET`.

## Fase 3 — Firebase Auth y notificaciones (en paralelo con la Fase 2)

1. Proyecto Firebase de producción nuevo.
2. `iluvatar` (ya unido con el BFF) verifica ID tokens de Firebase (SDK Admin de Go) en un middleware, sin restringir dominio; el identificador es el `uid`. Desaparecen la contraseña UTEM, el MD5, la cuenta admin fija, `ADMIN_ACCOUNT_*` y `MI_UTEM_API_HOST`.
3. Notificaciones con **FCM HTTP v1** y **topics** (uno general y uno por servicio) en vez de un envío por dispositivo; esto elimina la necesidad de guardar tokens de dispositivo en el servidor.
4. Con lo anterior, `iluvatar` no necesita base de datos propia: no guarda estudiantes ni tokens de dispositivo.
5. El proveedor de acceso (Google, Apple, correo) se elige en la PWA; el backend solo verifica el token y no depende del proveedor.
6. Endpoint `POST /v1/push/subscribe` para suscribir el token web de la PWA a los topics (el cliente web no puede hacerlo por sí solo; ver Fase F).
7. Cuenta de prueba de Firebase para pruebas. No hay revisión de tiendas: la PWA solo existe como web (D10).

**Estado: hecha en código, con dos cosas sin verificar.** `iluvatar`, rama `feat/fase-3-firebase` (commit `d53979c`), sin push.
- Hecho: verificación de ID tokens con el SDK Admin (middleware `RequireAuth`); eliminados Mi UTEM, JWT propio, MD5, cuenta admin y GORM/PostgreSQL (`iluvatar` ya no tiene base de datos); push por FCM v1 con **un solo mensaje por condición** (`'all' in topics || 'service.<id>' in topics`); `POST /v1/push/subscribe` y `/unsubscribe` con validación de topics. Cobertura total 62,8 %; `govulncheck` sigue en 3 alcanzables (Echo ×2 sin parche y biblioteca estándar).
- Probado con el emulador de Firebase Auth (SDK real) y con un e2e completo: Strapi real + emulador + las 8 rutas del BFF.
- **Sin verificar: la verificación de firma de tokens y el envío FCM reales.** El emulador no firma los tokens y no existe emulador de FCM; hace falta un proyecto de Firebase de prueba con su cuenta de servicio.
- Pendiente (no es código): crear el proyecto de Firebase de producción (punto 1), definir `FIREBASE_PROJECT_ID` y las credenciales, y la cuenta de prueba.
- Cambios de contrato para los clientes: ya no existen `/v1/auth/*`; el cliente envía `Authorization: Bearer <ID token de Firebase>` (el prefijo `Bearer ` es obligatorio) y todo fallo de autenticación es 401. El webhook responde 502 si falla el envío (antes 400).

## Fase 4 — Producción

- **Contenedores:** Dockerfile multi-etapa por servicio, usuario no root, `HEALTHCHECK`.
- **CI/CD:** build, tests, escaneo de vulnerabilidades y despliegue automático.
- **Hosting y base de datos gestionada con respaldos** (decisión pendiente, ver abajo).
- **Seguridad:** TLS, CORS restringido a los dominios de la PWA (hoy `middleware.CORS()` abierto), rate limiting, Swagger desactivado o protegido en producción, secretos fuera del repo.
- **Observabilidad:** logs estructurados sin datos personales, métricas, alertas y una verificación de salud real (`/health` hoy no comprueba dependencias).
- **Resiliencia:** el informe promete un BFF que responde aunque la API o el CMS caigan; hoy no está implementado (se puede resolver con caché).
- **Carga (fuera del monorepo, por decisión de nov 2026; las pruebas siguen en el repo `kumelen-load-tests`):** repetir las pruebas de Artillery sobre el entorno nuevo, con la meta del informe de 8.000 usuarios y menos de 3 segundos.
- **Legal:** actualizar Términos y Condiciones y política de privacidad (ya no hay credenciales UTEM; el informe clasifica los datos como sensibles).

**Estado: hecha en código y verificada en local; la salida a un entorno real sigue pendiente.** `iluvatar` rama `feat/fase-4-produccion` (commit `8bcc61b`), `iluvatar-cms` rama `feat/fase-4-produccion` (`ce74b2c`) y `kumelen-load-tests` rama `feat/fase-4-cargas` (`6c97ddc`), sin push.
- Hecho y verificado:
  - **PostgreSQL** (cierra la deuda de la Fase 2): Strapi 5.61 sobre PostgreSQL 17, con seed, API y el e2e de `iluvatar`.
  - **Imágenes**: `iluvatar` (≈50 MB, `distroless`, usuario no root, apagado ordenado) y `iluvatar-cms` (usuario `node`, panel compilado); ambas construidas y ejecutadas con Podman.
  - **Seguridad** en `iluvatar`: CORS por lista de orígenes (cerrado en producción sin configuración), cabeceras de seguridad y HSTS, límite de cuerpo de 1 MB, rate limit por IP que no se evade con un `X-Forwarded-For` falso, timeouts, Swagger cerrado en producción, `PUSH_DISABLED` rechazado en producción. CORS cerrado también en el CMS.
  - **Observabilidad**: log JSON sin tokens ni correos ni términos de búsqueda (con test), ID de petición, `/health` y `/ready` (comprueba el CMS).
  - **Resiliencia**: caché del CMS que sirve la última respuesta buena hasta 24 h si el CMS cae; probado con el CMS detenido (`/ready` 503, `/v1/home` 200).
  - **CI** (en los repos antiguos; el monorepo no lo incluye por decisión de nov 2026, se agregará más adelante): build, tests, escaneo y construcción de imagen en ambos repos; en `iluvatar`, publicación de la imagen en `ghcr.io` solo al crear una etiqueta `v*`. Los workflows están escritos pero no se han ejecutado en GitHub.
  - **Carga** (local, una sola máquina, 30 usuarios de prueba): recorrido mixto, 25 usuarios nuevos/s durante 30 s con caché: 9.120 peticiones, 0 errores, p95 de 2 ms; sin caché, 10/s: 3.720 peticiones, 0 errores, p95 de 21 ms. **No valida los 8.000 usuarios del informe**: falta probar en el entorno real, con red, varias IP y el CMS en su hosting.
  - **Legal**: `iluvatar/PRIVACIDAD.md` con el inventario de datos (no es asesoría legal).
- Sin hacer / pendiente:
  - Hosting, dominio y TLS (**decisión abierta**): sin destino no hay despliegue automático; el workflow de publicación solo sube la imagen.
  - PostgreSQL gestionado con respaldos y almacenamiento de imágenes del CMS fuera del disco (decisión abierta).
  - Métricas y alertas (solo hay logs y sondas); la caché y el rate limit son por réplica.
  - Escribir los Términos y Condiciones y la política de privacidad (necesitan decisiones y revisión legal).
  - Proyecto de Firebase de producción con cuenta de servicio, y probar el envío FCM real.
  - No pude confirmar que el runtime marque el contenedor como "healthy" (el Podman de macOS no ejecuta los healthchecks); el comando sí funciona dentro del contenedor.

**Compose local (añadido tras la Fase 4):** `compose.yaml`, `compose-env.sh` y `COMPOSE.md` en la carpeta raíz levantan PostgreSQL, CMS, API/BFF y PWA. **Validado con `docker compose config` pero nunca ejecutado**; la imagen de la PWA no se construyó y el CMS cambió después de su última construcción. Para armarlo hubo que añadir `CMS_PUBLIC_URL` e `ILUVATAR_TOKEN_FILE` a `iluvatar` (ramas `feat/compose-local` de `iluvatar` y `iluvatar-cms`). `manwe-pwa` ya tiene su `Dockerfile` (construido y servido con Podman) y soporte del emulador de Firebase Auth (13 tests); falta probar el inicio de sesión con el emulador de punta a punta en un navegador, y que `manwe-pwa` haga su primer commit.

## Fase F — PWA que reemplaza a `manwe` (aparte del plan de corrección)

**Decisión D8:** se deja de depender de Flutter. `manwe` se reescribe como una PWA con un framework de JavaScript, como copia visual 1:1 de la app actual. Se reemplaza la decisión anterior de actualizar Flutter y publicar en tiendas.

**Principio: paridad, sin features nuevas.** El alcance es solo la base que existe hoy en `manwe`. El único cambio funcional es el inicio de sesión, que pasa a Firebase Auth (D5).

### Alcance: lo que existe hoy

Pantallas actuales (según `lib/src/ui`): onboarding (comunicación y contenido), login, navegación principal, catálogo, publicaciones por categoría, publicaciones por servicio, búsqueda, publicación, servicio con pestañas, notificaciones, perfil, y las pantallas de error, sin contenido e imagen a pantalla completa.

Consume del BFF: `/v1/home`, `/v1/catalog`, `/v1/category/{id}/posts`, `/v1/service/{id}` y `/v1/service/{id}/posts`, `/v1/post/{id}`, `/v1/post/search/{term}` y `/v1/notification`. Los endpoints de `auth` desaparecen (D5).

Material de diseño que ya existe: tipografía Nunito (`fonts/`) e imágenes en `assets/images/` (logo, íconos, onboarding, marca UTEM).

### Qué puede hacer una PWA (evaluado)

| Capacidad | Veredicto | Condiciones |
|---|---|---|
| Contenido sin conexión | Sí | Service worker (Workbox): armazón precargado, respuestas del BFF con *stale-while-revalidate*, imágenes con tope, lo guardado por el usuario en IndexedDB. Solo está disponible lo ya abierto; Safari puede borrar datos de sitios sin uso, salvo que estén instalados |
| Notificaciones push | Sí, con límite en iPhone | Android: funciona. iPhone/iPad (iOS 16.4 o superior): solo con la PWA agregada a la pantalla de inicio y permiso aceptado. Requiere HTTPS y `firebase-messaging-sw.js` |
| Varios navegadores móviles | Sí | En iPhone todos usan el motor de Safari. Probar Safari iOS, Chrome, Samsung Internet y Firefox en Android. Instalación distinta (Safari no ofrece aviso). Cuidar zonas seguras, `dvh` y tipografías propias |
| Accesibilidad | Sí, es la mayor ventaja | Meta WCAG 2.2 AA (axe, Lighthouse, prueba manual con VoiceOver y TalkBack). Una copia 1:1 no hereda la accesibilidad: hay que replicar el aspecto, no la estructura |

Detalles a verificar con la documentación vigente antes de construir: política de borrado de datos de Safari, requisitos de push en iOS y soporte de alternativas de motor en la UE.

### Ajustes necesarios en el backend (Fases 3 y 4)

- **Suscripción a topics en web:** el cliente web no puede suscribirse a un topic de FCM por sí solo. `iluvatar` necesita un endpoint pequeño (`POST /v1/push/subscribe`, autenticado con el token de Firebase) que reciba el token de la PWA y lo suscriba con el SDK Admin, reenviado cada vez que el token cambie. El servidor sigue sin guardar tokens.
- **CORS:** restringir a los dominios de la PWA.
- **HTTPS** en todos los servicios.

### Framework (decidido: SvelteKit, D9)

Se eligió **SvelteKit**. Análisis que respalda la decisión:

Opción recomendada: **SvelteKit** con TypeScript y `@vite-pwa/sveltekit` (Workbox). Razones: bundle muy liviano para celulares modestos, advertencias de accesibilidad en el compilador, soporte nativo de service worker y renderizado en servidor opcional para vistas previas de enlaces. Alternativa segura: **React + Vite** (más ecosistema; el modelo de componentes y estado es el más cercano a Flutter con BLoC). Se descartan Next.js/Nuxt (más peso del necesario), Angular y Astro (no encaja con una app con sesión y modo offline). La decisión final depende de qué framework domine quien lo mantenga.

Piezas comunes: Firebase JS SDK modular (Auth y Messaging), caché de datos con TanStack Query, CSS con variables para los tokens de diseño, Vitest + Testing Library, Playwright con Chromium, Firefox y WebKit para pruebas entre navegadores, `eslint` con reglas de accesibilidad.

### Estado

**Primera versión creada en `manwe-pwa/`** (SvelteKit 3, TypeScript, sin commitear): todas las pantallas de `manwe` con paridad visual, cliente del BFF, login con Firebase Auth (más un modo simulado), Markdown saneado, service worker con caché de páginas, API e imágenes, BFF simulado y 24 pruebas unitarias. Verificada en el navegador integrado con el BFF simulado y auditada con axe-core. Detalles, diferencias deliberadas y lo no verificado en `manwe-pwa/README.md`.

Pendiente: notificaciones push (falta el proyecto de Firebase y el endpoint `POST /v1/push/subscribe`), probar el modo sin conexión y Firebase real en un navegador completo, y decidir el contraste del verde primario.

**Contenedor y emulador de Auth (añadidos para el compose local):** `manwe-pwa/Dockerfile` (multietapa sobre `node:24-slim`, usuario `node`, `HEALTHCHECK`) y `.dockerignore`; se construyó con Podman y sirve la página, el service worker y el manifiesto. Nueva variable `PUBLIC_FIREBASE_AUTH_EMULATOR_HOST` (vacía por defecto): con ella definida la app se conecta al emulador de Firebase Auth antes de usar Auth; con ella vacía el comportamiento no cambia. Cubierto por pruebas unitarias con el SDK simulado. **Sin probar:** el inicio de sesión de la PWA contra el emulador real (requiere ejecutar `firebase-tools`, que no se autorizó), el registro del service worker dentro de la imagen y que Docker ejecute el `HEALTHCHECK`.

### Look and feel en Penpot (D11)

**Decisión D11:** el diseño visual se gestiona en **Penpot** (código abierto), autoalojado con `docker-compose`. Los tokens siguen viviendo en el código (`manwe-pwa/src/app.css`) como fuente de verdad; Penpot es la referencia visual y el lugar donde se diseñan los cambios antes de tocar código. No hay una sincronización automática entre ambos: los tokens se copian de `app.css` a Penpot y no al revés.

**Se puede guardar en Penpot el look actual de `manwe`.** Es el siguiente trabajo, aún no hecho. Enfoque propuesto:

1. **Plugin de Penpot** (API de plugins, JavaScript) que lee una especificación en JSON y crea, en un archivo de Penpot: la biblioteca de colores y tipografías (verde `#459951`, azul `#00417C`, texto `#3C4046`, fondo `#F1F5F0`, Nunito 300/400/600/700), los componentes (cabecera, barra inferior, tarjeta de publicación, tarjeta de servicio, fila de lista, tarjeta de notificación, chips, botones de contacto, diálogo) y un tablero de 390×844 por pantalla. Se instala desde el gestor de plugins de Penpot indicando la URL del manifiesto.
2. **La especificación sale del código ya existente:** tokens de `app.css`, medidas de `manwe/lib/src/ui` y datos de ejemplo de `mock-bff`. Las imágenes del diseño se cargan desde esa misma fuente.
3. **Alternativa más simple:** generar un SVG por pantalla e importarlo en Penpot. Es más rápido, pero queda como formas sueltas sin componentes ni tokens reutilizables.

**Lo que no queda guardado:** una captura fiel de la app Flutter corriendo (nunca se ejecutó). La referencia es el código Dart, las capturas del informe y la PWA ya construida.

**Instalación de Penpot (documentada, no ejecutada):**

- Requiere Docker con Compose. En esta máquina los binarios de Docker existen pero no están en el `PATH`, y hay Podman Desktop instalado.
- Descargar el `docker-compose.yaml` oficial del repositorio de Penpot (carpeta `docker/images`), fijar una versión en lugar de `latest` y guardarlo en una carpeta propia, por ejemplo `kumelen/penpot/`.
- Fijar `PENPOT_PUBLIC_URI=http://localhost:9001` y los flags de registro y verificación de correo según la documentación vigente. Con el correo de verificación desactivado, basta registrarse una vez desde la interfaz.
- Levantar con `docker compose up -d` y abrir `http://localhost:9001`. Los datos (PostgreSQL y activos) quedan en volúmenes de Docker: respaldarlos si el diseño pasa a ser un activo del proyecto.
- Verificar los nombres exactos de servicios, variables y flags contra la documentación oficial antes de usarlos, porque cambian entre versiones.

**Pendiente:** instalar Penpot, escribir y probar el plugin, y decidir dónde se versiona la especificación JSON (propuesta: en `manwe-pwa/design/`).

#### Penpot con agentes de IA (D12, para implementar más adelante)

**Decisión D12:** aprovechar la integración de Penpot con agentes de IA, sin depender de ella hasta comprobarla. Estado verificado en fuentes públicas en nov 2026 (no probado en esta máquina):

- **Servidor MCP oficial** ([documentación](https://help.penpot.app/mcp/)): permite que un cliente MCP (por ejemplo Claude Code) lea y trabaje sobre archivos de Penpot. Se conecta con una URL y, dentro del archivo, con `File → MCP Server → Connect`. La documentación lo describe como alojado por Penpot y solo activo en una pestaña del navegador a la vez.
- **Servidores de la comunidad** para instancias autoalojadas: [ancrz/penpot-mcp-server](https://github.com/ancrz/penpot-mcp-server) (Apache 2.0) y [montevive/penpot-mcp](https://github.com/montevive/penpot-mcp). Al ser de terceros, revisar licencia y seguridad antes de usarlos; si piden credenciales, usar un token de acceso con permisos mínimos y no la contraseña principal.
- **API de plugins** abierta, con acceso a tokens desde la versión 2.14 (dato de segunda mano, sin verificar en las notas oficiales), y tokens en formato abierto W3C DTCG. Existe el plugin [DTCG Token Import](https://penpot.app/penpothub/plugins/dtcg-toke-import), que crea los conjuntos de tokens desde un JSON y los resincroniza sin duplicar.

**Uso previsto:**

1. **Tokens:** generar un JSON DTCG desde `manwe-pwa/src/app.css` y cargarlo con el plugin de importación. Reduce el plugin propio de D11 a los componentes y tableros.
2. **Diseño a código:** un agente con MCP lee un diseño aprobado en Penpot y lo implementa en `manwe-pwa`.
3. **Código a diseño:** un agente crea o actualiza tableros y componentes en Penpot desde la PWA.

**Prueba previa (antes de depender de esto):** con la instancia local, conectar un MCP, leer un tablero y crear un componente simple. Comprobar en particular si el servidor oficial funciona con una instancia local o si hay que usar uno de la comunidad, y cuántas herramientas expone (una fuente indica que son pocas).

**Riesgos:** ecosistema joven y documentación con inconsistencias entre alojado y local; credenciales de Penpot en servidores de terceros; el servidor oficial podría enviar datos del diseño a un servicio externo, lo que importa si algún diseño contuviera datos reales de estudiantes (no deberían contenerlos: usar datos de ejemplo).

### Pasos previos

1. Extraer de `manwe` una especificación visual (colores, tipografías, espaciados, componentes de las pantallas) que sirva de contrato para el 1:1.
2. Auditar la accesibilidad del diseño actual (contraste, tamaño de los botones, texto de tamaño fijo) antes de copiarlo.
3. ~~Elegir el framework~~ → hecho: SvelteKit.

### Riesgos

- El push en iPhone depende de que el usuario instale la PWA.
- Los enlaces compartidos no muestran vista previa sin renderizado en servidor o prerenderizado.
- La app actual en Flutter no se ha podido ejecutar (Flutter no está instalado): la referencia visual será el código y las capturas del informe.

### Decisiones abiertas de la PWA

1. ~~Framework final~~ → resuelta: SvelteKit (D9).
2. ~~¿Versión en tiendas?~~ → resuelta: no, solo web (D10).
3. ~~SPA o renderizado en servidor~~ → resuelta: SSR (D10).
4. Dominio y hosting de la PWA — se deciden más adelante. Con SSR el hosting debe ejecutar un servidor Node (por ejemplo con `adapter-node`) o una plataforma compatible con el adaptador elegido, no basta con archivos estáticos.
5. **(Diferida: se mitiga más adelante)** Vistas previas de enlaces: el SSR solo las logra si el servidor puede leer los metadatos de una publicación (título, resumen, imagen). Hoy todas las rutas `/v1/post/*` del BFF exigen token, y el servidor de la PWA no tiene el token del usuario. Hay que decidir entre un endpoint público de solo metadatos o una credencial de servicio para el SSR.

## Fase L — Laurelin (después del plan de corrección)

**Por qué existe:** al quitar Mi UTEM y la restricción de dominio, Kümelen pierde la fuente de datos de estudiantes y la verificación de que alguien es estudiante. Para salir a producción de forma autónoma se necesita una aplicación propia que cumpla ese rol. Se llamará **Laurelin**.

**Qué es:** un CRUD de estudiantes universitarios chilenos, de pregrado y postgrado. Es el reemplazo de lo que Mi UTEM aportaba a Kümelen, pero no limitado a una sola universidad.

**Cuándo:** se inicia una vez terminadas las fases 0 a 4 (el plan de corrección). No se diseña ni construye antes.

### Alcance inicial (borrador, a refinar)

- Alta, consulta, edición y baja de estudiantes.
- Distinción de nivel: pregrado y postgrado (magíster, doctorado, otros).
- Datos de institución y carrera o programa.
- Autenticación de quienes administran el registro.

### Puntos de contacto con Kümelen (a definir)

- Laurelin sería la fuente que verifique que una cuenta de Firebase corresponde a un estudiante (por ejemplo, como dato en los tokens de Firebase), recuperando la verificación que se pierde al quitar la restricción de dominio.
- Podría devolver los datos que antes venían de Mi UTEM (carrera, nivel) para personalizar contenido.

### Consideraciones desde ya

- **Datos personales de estudiantes chilenos:** exige minimización de datos, cifrado, control de acceso y trazabilidad. Revisar el marco legal vigente en Chile (Ley 19.628 y la nueva ley de protección de datos personales, 21.719) antes de decidir qué se almacena, en particular si se guarda el RUT.
- **Reutilizar lo aprendido:** misma base técnica que Kümelen (Go 1.27, PostgreSQL, Firebase Auth, contenedores, CI/CD) para no abrir otro frente tecnológico.
- Este plan solo registra la intención; el diseño detallado de Laurelin va en un documento aparte cuando llegue su turno.

### Decisiones abiertas de Laurelin

1. ¿Quién carga y mantiene los datos: una institución, un administrador central o los propios estudiantes?
2. ¿Es multi-institución desde el inicio o parte con una sola?
3. ¿Qué identifica a un estudiante (RUT, correo, ambos) y cómo se verifica su condición?
4. ¿Laurelin y Kümelen se comunican por API entre servicios o por reclamos (claims) en los tokens de Firebase?
5. Stack y hosting: ¿mismo de Kümelen?

## Orden y dependencias

```
Fase 0 → Fase 1 → ┬ Fase 2 ┬→ Fase 4 → Fase L (Laurelin) → (Fase F, aparte)
                  └ Fase 3 ┘
```

Las fases 2 y 3 tocan `iluvatar`/`elentari` en archivos distintos (consultas y webhook frente a login y push), pero ambas modifican `go.mod`.

## Decisiones abiertas

1. **Hosting y PostgreSQL gestionado:** dónde y con qué presupuesto.
2. **Proveedores de acceso** (Google, Apple, correo) — se decide en la PWA.
3. **Almacenamiento de imágenes** de Strapi.
4. **Quién administra el contenido** y con qué roles.
