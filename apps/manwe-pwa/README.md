# manwe-pwa

Reescritura de `manwe` (la app Flutter de Kümelen) como PWA con SvelteKit. Es una copia visual 1:1 y **sin funcionalidades nuevas**; el único cambio funcional es el inicio de sesión, que pasa de Pasaporte.UTEM a Firebase Auth (decisiones D5, D8, D9 y D10 de `../PLAN_MIGRACION.md`).

## Cómo ejecutarla

Requiere Node 22.14 o superior (se probó con Node 24).

```bash
npm install
npm run mock        # BFF simulado en http://localhost:8787 (otra terminal)
npm run dev         # http://localhost:5173
```

Con `PUBLIC_AUTH_MODE=mock` (el valor por defecto de `.env.example`) el botón "Ingresar con Google" inicia una sesión local sin contactar a Firebase. Para usar el BFF real basta con apuntar `PUBLIC_API_BASE` a `iluvatar`.

| Comando | Qué hace |
|---|---|
| `npm run check` | Tipos y accesibilidad estática (`svelte-check`) |
| `npm test` | Pruebas unitarias (Vitest) |
| `npm run build` | Build de producción con adaptador Node (SSR) y service worker |
| `npm start` | Sirve la build (`node build`); lee las variables `PUBLIC_*` al arrancar |

Variables (ver `.env.example` y `src/env.ts`): `PUBLIC_API_BASE`, `PUBLIC_CMS_BASE`, `PUBLIC_AUTH_MODE`, las `PUBLIC_FIREBASE_*` y `PUBLIC_FIREBASE_AUTH_EMULATOR_HOST`. **`PUBLIC_API_BASE` también se fija al construir**, porque la regla de caché del service worker usa ese origen: hay que compilar con el valor de producción.

## Emulador de Firebase Auth

Con `PUBLIC_AUTH_MODE=firebase` y `PUBLIC_FIREBASE_AUTH_EMULATOR_HOST=localhost:9099` (la dirección del emulador **tal como la ve el navegador**), la app llama a `connectAuthEmulator(auth, 'http://localhost:9099', { disableWarnings: true })` antes de usar Auth, así que el botón "Ingresar con Google" abre la ventana falsa del emulador para elegir o crear una cuenta. Con la variable vacía (el valor por defecto) el comportamiento es el de Firebase real y no se contacta ningún emulador.

- Acepta `host:puerto` y tolera un `http://` inicial o una barra final.
- Con el emulador, si `PUBLIC_FIREBASE_API_KEY` o `PUBLIC_FIREBASE_PROJECT_ID` están vacías se usan `demo-api-key` y `demo-kumelen`, porque el SDK exige una clave aunque el emulador no la valide. Sin emulador no se inventa ningún valor.
- El emulador **no firma** los tokens: el servidor debe aceptarlos explícitamente (`iluvatar` con `FIREBASE_AUTH_EMULATOR_HOST` y `FIREBASE_PROJECT_ID=demo-kumelen`). Nunca usar esto con datos reales.
- Se levanta con `npx firebase-tools emulators:start --only auth --project demo-kumelen` (Node 24; ver `../firebase/firebase.json`).

## Docker

El `Dockerfile` es multietapa sobre `node:24-slim`. Las nueve variables `PUBLIC_*` se piden como `ARG` y se exportan al construir (`PUBLIC_API_BASE` también se usa en la compilación, por la regla de caché del service worker); en ejecución el servidor las vuelve a leer del entorno. Deben ser las direcciones que ve el navegador. La imagen corre como el usuario `node`, escucha en el puerto 3000 y trae un `HEALTHCHECK`.

```bash
docker build -t kumelen/manwe-pwa \
  --build-arg PUBLIC_API_BASE=http://localhost:8080 \
  --build-arg PUBLIC_AUTH_MODE=firebase \
  --build-arg PUBLIC_FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 .
docker run -p 3000:3000 -e ORIGIN=http://localhost:3000 \
  -e PUBLIC_API_BASE=http://localhost:8080 -e PUBLIC_AUTH_MODE=firebase \
  -e PUBLIC_FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 kumelen/manwe-pwa
```

Con Podman usar `podman build --format docker` para conservar el `HEALTHCHECK`. El `.dockerignore` deja fuera `node_modules`, `build`, `.svelte-kit`, `.env`, las pruebas, `mock-bff` y los `*.md`.

## Estructura

```
src/
  env.ts                    variables públicas (SvelteKit 3: $app/env/public)
  lib/
    api/                    cliente del BFF (token + reintento tras 401), endpoints y modelos
    auth/                   AuthProvider: Firebase (Google) y mock
    components/             piezas de interfaz (una por widget de manwe)
    markdown.ts             Markdown → HTML saneado (DOMPurify)
    resource.svelte.ts      carga de datos (equivale a los cubits de manwe)
    session.svelte.ts       estado de sesión
    offline.ts              borrado del caché al cerrar sesión
  routes/
    (tabs)/                 Publicaciones, Buscar (catálogo) y Notificaciones, con cabecera y barra inferior
    login, bienvenida, perfil, busqueda
    publicaciones/[id], servicios/[id], servicios/[id]/publicaciones, categorias/[id]
mock-bff/server.mjs         BFF simulado con la misma forma de respuesta
tests/                      pruebas unitarias
```

## Paridad con manwe

Todas las pantallas de `manwe/lib/src/ui` existen y consumen los mismos endpoints (`/v1/home`, `/v1/catalog`, `/v1/category/{id}/posts`, `/v1/service/{id}` y `/v1/service/{id}/posts`, `/v1/post/{id}`, `/v1/post/search/{term}`, `/v1/notification`). Colores, tipografía Nunito, márgenes, radios, sombras y las proporciones de altura (`size.height * 0.19`, etc.) salen del código Dart; se expresan con `--vu` y `--wu` en `src/app.css`.

### Diferencias deliberadas

| Diferencia | Motivo |
|---|---|
| El login usa Google (Firebase) y no pide correo ni contraseña | Decisión D5; el backend ya no recibe contraseñas |
| El perfil no muestra "Tu carrera" | El dato venía de Mi UTEM y ya no existe |
| El botón de correo de feedback no está | El código actual de `manwe` no lo muestra (solo aparece en las capturas del informe) |
| Las pantallas de carga y error conservan la barra superior con la flecha | En una PWA instalada no hay botón de atrás del sistema en iOS |
| El texto "Serás ridirigido a" se corrigió a "redirigido" | Errata en el original |
| Los puntos del onboarding son botones y el botón de saltar/salir va bajo ellos | Accesibilidad por teclado; la disposición exacta del paquete `onboarding` no se pudo verificar |
| Layout de columna móvil de 480 px máximo | En escritorio la app se muestra como en un teléfono |

## Estado de verificación

Verificado con Podman: la imagen se construye (472 MB), arranca como `node` y sirve la página, `sw.js`, el manifiesto, los íconos y las fuentes con código 200; las variables `PUBLIC_*` llegan a la página en ejecución y la regla de caché del service worker apunta al `PUBLIC_API_BASE` de construcción. El comando del `HEALTHCHECK` responde 200 ejecutado a mano dentro del contenedor.

Verificado en el navegador integrado (viewport 390×844, BFF simulado): arranque y redirecciones (onboarding → login → inicio), las pantallas de inicio, catálogo, notificaciones, perfil, publicación, servicio con pestañas y búsqueda, y el diálogo de confirmación de enlaces. `npm run check` sin errores, 24 pruebas unitarias en verde y build de producción correcto (service worker con 75 archivos precargados y tres cachés).

**No verificado:**
- El modo sin conexión: el navegador integrado no admite service workers (falla incluso uno trivial). Hay que probarlo en Chrome o Safari: visitar páginas con red, apagar el BFF y recargar.
- Firebase Auth real: solo se probó el proveedor simulado.
- **El emulador de Firebase Auth de punta a punta:** las pruebas unitarias comprueban con el SDK simulado que se llama a `connectAuthEmulator` con la dirección correcta, antes de usar Auth y solo si la variable está definida. No se levantó el emulador ni se inició sesión contra él.
- Que el service worker se registre en la imagen: el navegador integrado no admite service workers; solo se comprobó que `sw.js` y el manifiesto se sirven.
- Que Docker ejecute el `HEALTHCHECK` (se comprobó que la imagen lo conserva y que el comando funciona a mano; Podman en macOS no lo ejecuta).
- Notificaciones push: **no están implementadas**. Faltan el proyecto de Firebase, el endpoint `POST /v1/push/subscribe` en `iluvatar` y el manejador de push en el service worker.
- Comparación píxel a píxel con la app Flutter: nunca se ejecutó (no hay Flutter instalado); la referencia fue el código y las capturas del informe.
- Safari en iPhone y lectores de pantalla (VoiceOver, TalkBack).

## Accesibilidad

Auditoría automática con axe-core (WCAG 2.2 AA) sobre las pantallas. Se corrigieron: nombre accesible del enlace de perfil, falta de landmark en la cabecera, falta de `h1` en cuatro pantallas y contenedor desplazable no enfocable.

**Pendiente de decisión: contraste.** El texto blanco sobre el verde `#459951` (botones "Ver más", nombres de servicio, categorías, pestañas, "Saltar") tiene una relación de contraste de unos 3,5:1, por debajo del 4,5:1 que exige AA para texto normal. Lo mismo ocurre con el rojo del botón "Cerrar Sesión". Se mantuvo el color para respetar el 1:1; oscurecer `--primary` en `src/app.css` (por ejemplo a `#2f7a3b`, cerca de 5,3:1) lo corrige en toda la app con un cambio.

## Notas técnicas

- SvelteKit 3 y Vite 8: el alias `#lib` del plantilla no resuelve módulos sin extensión, por eso el proyecto usa `$lib` (alias en `vite.config.ts` y `tsconfig.json`).
- Las respuestas del BFF se guardan en el caché `manwe-api` con estrategia *network first*; al cerrar sesión se borra para no exponer contenido en dispositivos compartidos.
- Con SSR, las pantallas se renderizan en el servidor en estado de carga y los datos se piden en el navegador con el token del usuario. Las vistas previas de enlaces compartidos quedan pendientes (ver el plan).
