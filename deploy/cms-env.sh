#!/usr/bin/env bash
# Crea services/iluvatar-cms/.env para ejecutar el CMS en el equipo (fuera de Docker), con SQLite y secretos
# aleatorios. Toma WEBHOOK_SECRET de deploy/.env para que coincida con el que valida iluvatar.
# No sobrescribe un .env existente. Uso: ./cms-env.sh   (desde deploy/; antes ejecuta ./compose-env.sh)
set -euo pipefail

cd "$(dirname "$0")"
destination="../services/iluvatar-cms/.env"

if [ -e "$destination" ]; then
  echo "$destination ya existe: no se sobrescribe (bórralo si quieres generar otro)." >&2
  exit 1
fi
if [ ! -f .env ]; then
  echo "Falta deploy/.env: ejecuta ./compose-env.sh primero." >&2
  exit 1
fi

webhook_secret="$(grep -E '^WEBHOOK_SECRET=' .env | head -n1 | cut -d= -f2-)"
if [ -z "$webhook_secret" ]; then
  echo "deploy/.env no define WEBHOOK_SECRET." >&2
  exit 1
fi

random() { openssl rand -base64 24 | tr -d '\n'; }

cat > "$destination" <<ENV
# Generado por deploy/cms-env.sh. No lo versiones: contiene secretos.
HOST=0.0.0.0
PORT=1337
APP_KEYS="$(random),$(random),$(random),$(random)"
API_TOKEN_SALT=$(random)
ADMIN_JWT_SECRET=$(random)
TRANSFER_TOKEN_SALT=$(random)
JWT_SECRET=$(random)
ENCRYPTION_KEY=$(random)

# Desarrollo local: SQLite en .tmp/data.db
DATABASE_CLIENT=sqlite

# Aviso a iluvatar (publicado en el equipo por el compose) cada vez que se publica contenido.
ILUVATAR_WEBHOOK_URL=http://localhost:8080/v1/hooks/post
WEBHOOK_SECRET=${webhook_secret}
ENV
chmod 600 "$destination"
echo "Creado $destination"
