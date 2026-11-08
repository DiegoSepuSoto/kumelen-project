#!/usr/bin/env bash
# Crea .env para compose.yaml con secretos aleatorios. No sobrescribe un .env existente.
# Uso: ./compose-env.sh [ruta]   (por defecto ./.env)
set -euo pipefail

destination="${1:-.env}"
if [ -e "$destination" ]; then
  echo "$destination ya existe: no se sobrescribe (bórralo si quieres generar otro)." >&2
  exit 1
fi

random() { openssl rand -base64 24 | tr -d '\n'; }

cat > "$destination" <<ENV
# Generado por compose-env.sh. No lo versiones: contiene secretos.
POSTGRES_PASSWORD=$(random | tr -d '/+=')
APP_KEYS=$(random),$(random),$(random),$(random)
API_TOKEN_SALT=$(random)
ADMIN_JWT_SECRET=$(random)
TRANSFER_TOKEN_SALT=$(random)
JWT_SECRET=$(random)
ENCRYPTION_KEY=$(random)
WEBHOOK_SECRET=$(random)
LAURELIN_DB_PASSWORD=$(random | tr -d '/+=')
LAURELIN_SERVICE_TOKEN=$(random | tr -d '/+=')
# Primer administrador de Laurelin. Con el emulador, crea una cuenta con este correo y márcala como verificada
# (ver COMPOSE.md); con un proyecto real, pon tu correo.
LAURELIN_BOOTSTRAP_ADMIN_EMAIL=admin@laurelin.test

# --- Firebase: elige UNA de las dos opciones ---
# A) Emulador de Auth (sin proyecto real; el botón de Google de la PWA abre una ventana falsa del emulador):
#    docker compose --profile emulator up --build
FIREBASE_PROJECT_ID=demo-kumelen
FIREBASE_AUTH_EMULATOR_HOST=firebase-emulator:9099
GOOGLE_APPLICATION_CREDENTIALS=
PUBLIC_FIREBASE_AUTH_EMULATOR_HOST=localhost:9099

# B) Proyecto real de Firebase (de pruebas, nunca producción): borra el emulador de arriba y completa:
#    FIREBASE_PROJECT_ID=mi-proyecto
#    FIREBASE_AUTH_EMULATOR_HOST=
#    PUBLIC_FIREBASE_AUTH_EMULATOR_HOST=
#    GOOGLE_APPLICATION_CREDENTIALS=/secrets/firebase.json    (el archivo va en ./secrets/firebase.json)
#    PUBLIC_FIREBASE_API_KEY=...  PUBLIC_FIREBASE_AUTH_DOMAIN=...  PUBLIC_FIREBASE_PROJECT_ID=...
#    PUBLIC_FIREBASE_APP_ID=...   PUBLIC_FIREBASE_MESSAGING_SENDER_ID=...

# Notificaciones push reales (solo con un proyecto real y su cuenta de servicio): PUSH_DISABLED=false
PUSH_DISABLED=true
ENV
chmod 600 "$destination"
echo "Creado $destination"
