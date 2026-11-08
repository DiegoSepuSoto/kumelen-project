# syntax=docker/dockerfile:1

FROM docker.io/library/node:24-slim AS build
WORKDIR /app
# Herramientas por si alguna dependencia nativa (better-sqlite3) no trae binario precompilado.
RUN apt-get update && apt-get install -y --no-install-recommends python3 make g++ && rm -rf /var/lib/apt/lists/*
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
ENV NODE_ENV=production
# El panel lee esta opción al construirse; ver config/admin.js.
ARG ADMIN_COOKIE_SECURE
ENV ADMIN_COOKIE_SECURE=${ADMIN_COOKIE_SECURE}
RUN npm run build && npm prune --omit=dev

FROM docker.io/library/node:24-slim
WORKDIR /app
ENV NODE_ENV=production PORT=1337
COPY --from=build --chown=node:node /app /app
# Las imágenes subidas van a public/uploads: monta ahí un volumen o configura un proveedor de almacenamiento.
# /shared: directorio donde el compose de desarrollo entrega el API token a `iluvatar` (lo escribe `npm run seed`).
RUN mkdir -p /app/public/uploads /app/.tmp /shared && chown -R node:node /app/public/uploads /app/.tmp /shared
USER node
EXPOSE 1337
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD node -e "fetch('http://127.0.0.1:'+(process.env.PORT||1337)+'/_health').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"
CMD ["npm", "run", "start"]
