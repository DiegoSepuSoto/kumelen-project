# syntax=docker/dockerfile:1

FROM docker.io/library/golang:1.27 AS build
WORKDIR /src
# GOTOOLCHAIN=local: usa la versión de Go de la imagen y no descarga otra durante el build.
ENV GOTOOLCHAIN=local CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/iluvatar ./src

# Imagen mínima sin shell ni gestor de paquetes; trae certificados de CA (para hablar con Firebase) y corre sin root.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/iluvatar /iluvatar
ENV PORT=8080 ENV=production
EXPOSE 8080
USER nonroot:nonroot
# Sin curl en la imagen: el propio binario consulta /health.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/iluvatar", "healthcheck"]
ENTRYPOINT ["/iluvatar"]
