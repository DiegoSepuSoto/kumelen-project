# Laurelin

Registro de estudiantes universitarios chilenos (pregrado y postgrado) para Kümelen. Ver `CLAUDE.md` para el modelo, las rutas, las variables de entorno y cómo probarlo.

```bash
export DATABASE_URL=postgres://postgres:test@localhost:5432/laurelin FIREBASE_PROJECT_ID=demo-laurelin \
  FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 BOOTSTRAP_ADMIN_EMAIL=tu@correo.cl SERVICE_TOKEN=un-secreto-largo
go run ./src      # escucha en :8090
```
