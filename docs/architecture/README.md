# Diagramas de arquitectura

Diagramas de la arquitectura **actual** de Kümelen, hechos con [Archify](https://archify.si) (reemplazan a los C4 en PlantUML). Cada uno es un HTML autónomo e interactivo (temas claro y oscuro, exportable a imagen o SVG); ábrelo directo en el navegador.

| Diagrama | Qué muestra | Fuente |
|---|---|---|
| [`sistema.html`](sistema.html) | Servicios, almacenes y dependencias externas (Firebase Auth y FCM) | [`fuentes/sistema.architecture.json`](fuentes/sistema.architecture.json) |
| [`acceso-estudiante.html`](acceso-estudiante.html) | Ingreso con Firebase y verificación de estudiante activo con Laurelin | [`fuentes/acceso-estudiante.sequence.json`](fuentes/acceso-estudiante.sequence.json) |
| [`publicacion-push.html`](publicacion-push.html) | Publicar un post en el CMS hasta el push al dispositivo | [`fuentes/publicacion-push.sequence.json`](fuentes/publicacion-push.sequence.json) |
| [`laurelin.html`](laurelin.html) | Modelo de datos de Laurelin | [`fuentes/laurelin.erd.json`](fuentes/laurelin.erd.json) |

## Cómo se regeneran

El HTML se genera desde el JSON de `fuentes/`; **edita el JSON, no el HTML**. Los archivos `*.delivery.json` son los recibos de generación de Archify y se regeneran junto con el HTML.

```bash
make diagrams
```

Requiere la skill Archify instalada (`~/.claude/skills/archify`; ver https://archify.si) y Node 24. También puedes pedirle a Claude Code con la skill que actualice un diagrama.

## Qué no incluyen

- Muestran la arquitectura **actual**, no el diseño original (app Flutter, `elentari`/`aule`, MongoDB y Mi UTEM). Los C4 en PlantUML de ese diseño (`*.puml`, `components/` y `out/` en esta carpeta) ya no se mantienen; siguen en el historial de git y el informe de la tesis conserva sus imágenes.
- No hay diagramas de componentes por servicio: la estructura interna de cada uno está en su `CLAUDE.md`.
- La verificación fue automática (validación, comprobación del archivo y navegador); no se hizo una revisión visual manual.
