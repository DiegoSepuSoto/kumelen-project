# Imagen de la PWA (SvelteKit con adaptador Node).
#
# Las variables PUBLIC_* se piden como ARG porque PUBLIC_API_BASE se usa al construir (la regla de
# caché del service worker apunta a ese origen). En ejecución el servidor vuelve a leerlas del entorno.
# Deben ser las direcciones que ve el NAVEGADOR, no nombres internos de Docker.
#
#   docker build -t kumelen/manwe-pwa \
#     --build-arg PUBLIC_API_BASE=http://localhost:8080 .

FROM docker.io/library/node:24-slim AS build
WORKDIR /app

COPY package.json package-lock.json .npmrc ./
RUN npm ci

COPY tsconfig.json vite.config.ts ./
COPY src ./src
COPY static ./static

ARG PUBLIC_API_BASE
ARG PUBLIC_CMS_BASE
ARG PUBLIC_AUTH_MODE
ARG PUBLIC_FIREBASE_API_KEY
ARG PUBLIC_FIREBASE_AUTH_DOMAIN
ARG PUBLIC_FIREBASE_PROJECT_ID
ARG PUBLIC_FIREBASE_APP_ID
ARG PUBLIC_FIREBASE_MESSAGING_SENDER_ID
ARG PUBLIC_FIREBASE_AUTH_EMULATOR_HOST
ENV PUBLIC_API_BASE=$PUBLIC_API_BASE \
    PUBLIC_CMS_BASE=$PUBLIC_CMS_BASE \
    PUBLIC_AUTH_MODE=$PUBLIC_AUTH_MODE \
    PUBLIC_FIREBASE_API_KEY=$PUBLIC_FIREBASE_API_KEY \
    PUBLIC_FIREBASE_AUTH_DOMAIN=$PUBLIC_FIREBASE_AUTH_DOMAIN \
    PUBLIC_FIREBASE_PROJECT_ID=$PUBLIC_FIREBASE_PROJECT_ID \
    PUBLIC_FIREBASE_APP_ID=$PUBLIC_FIREBASE_APP_ID \
    PUBLIC_FIREBASE_MESSAGING_SENDER_ID=$PUBLIC_FIREBASE_MESSAGING_SENDER_ID \
    PUBLIC_FIREBASE_AUTH_EMULATOR_HOST=$PUBLIC_FIREBASE_AUTH_EMULATOR_HOST

RUN npm run build && npm prune --omit=dev

FROM docker.io/library/node:24-slim
WORKDIR /app
ENV NODE_ENV=production PORT=3000

COPY --from=build --chown=node:node /app/build ./build
COPY --from=build --chown=node:node /app/node_modules ./node_modules
COPY --from=build --chown=node:node /app/package.json ./package.json

USER node
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD node -e "fetch('http://127.0.0.1:3000/').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"

CMD ["node", "build"]
