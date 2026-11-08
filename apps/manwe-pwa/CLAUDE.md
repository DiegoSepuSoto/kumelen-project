# manwe-pwa

PWA de Kümelen en SvelteKit 3 (Svelte 5 con runas, TypeScript). Reemplaza a la app Flutter `manwe`; copia visual 1:1 sin features nuevas. Contexto y decisiones en `../PLAN_MIGRACION.md` (Fase F) y `README.md`.

## Comandos

```bash
npm run mock      # BFF simulado en :8787
npm run dev       # :5173
npm run check     # svelte-check
npm test          # vitest
npm run build     # build SSR con service worker
```

Usar Node 22.14 o superior (`source ~/.nvm/nvm.sh && nvm use 24`).

## Reglas

- Paridad: no agregar funcionalidades ni cambiar el diseño sin que se pida. La referencia visual es `../manwe/lib/src/ui` (colores, tamaños, textos) y las capturas del informe.
- Las medidas que en Flutter dependen de `size.height` / `size.width` usan `--vu` y `--wu` (ver `src/app.css`).
- Importar con `$lib/...` (no `#lib`).
- No escribir valores reales de Firebase ni de hosts en archivos versionados; van en `.env` (ignorado). Solo `.env.example` se versiona.
- `PUBLIC_FIREBASE_AUTH_EMULATOR_HOST` (vacía por defecto) conecta la app al emulador de Firebase Auth; solo para desarrollo y pruebas, y el servidor debe aceptar sus tokens sin firma. Con ella vacía el comportamiento no cambia. Detalles en `README.md`.
- Imagen: `Dockerfile` multietapa; las `PUBLIC_*` son `ARG` (la regla de caché del service worker usa `PUBLIC_API_BASE` al construir). Con Podman: `podman build --format docker`.
- El Markdown del CMS siempre pasa por `renderMarkdown` (saneado); nunca insertar HTML crudo con `{@html}` salvo los íconos fijos.
- Cada pantalla pide sus datos con `resource(...)` en el navegador; no usar `load` de servidor, porque el token vive en el cliente.
- Antes de dar algo por hecho: `npm run check && npm test && npm run build` y reportar el resultado real.

## Cosas que no están hechas

- Notificaciones push (FCM web): falta proyecto de Firebase y el endpoint `POST /v1/push/subscribe` en `iluvatar`.
- Modo sin conexión sin probar en un navegador real (el panel integrado no admite service workers).
- Inicio de sesión contra el emulador de Auth sin probar de punta a punta (solo hay pruebas unitarias del cableado).
- Contraste del verde primario (decisión pendiente, ver `README.md`).
