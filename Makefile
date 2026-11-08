# Comandos de trabajo del monorepo Kümelen. Ejecuta `make` (o `make help`) para ver la lista.
#
# Requisitos: Docker (por ejemplo Colima) y Go 1.27. Para la PWA y el CMS, Node 24 (se activa solo con nvm).
# Los targets de compose se ejecutan en `deploy/`, que es donde viven `compose.yaml` y `.env`.
# El CMS (Strapi) NO está en el compose: se ejecuta aparte, en el equipo (targets `cms-*`).

SHELL := /bin/bash

# Usa `docker compose` si existe el plugin; si no, el binario independiente `docker-compose`.
COMPOSE ?= $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo "docker-compose")

# Perfil de Compose: `emulator` levanta el emulador de Firebase Auth. Vacío = proyecto real de Firebase.
PROFILE ?= emulator
PROFILE_FLAG := $(if $(PROFILE),--profile $(PROFILE),)
DC := cd deploy && $(COMPOSE) $(PROFILE_FLAG)

# Recursos de la VM de Colima. El build del panel de Strapi necesita memoria (con 2 GiB se cae).
COLIMA_CPU ?= 4
COLIMA_MEM ?= 6

# Activa Node 24 con nvm si está instalado (el Node por defecto no sirve para Strapi 5 ni la PWA).
NODE24 := if [ -s "$$HOME/.nvm/nvm.sh" ]; then . "$$HOME/.nvm/nvm.sh" && nvm use 24 >/dev/null; fi;

# Servicio para `logs`, `restart`, `shell`: make logs S=iluvatar
S ?=

GO_SERVICES := services/iluvatar services/laurelin

.DEFAULT_GOAL := help

.PHONY: help
help: ## Muestra esta ayuda
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Variables: PROFILE=$(PROFILE) (vacío = sin emulador), S=<servicio>, COLIMA_CPU=$(COLIMA_CPU), COLIMA_MEM=$(COLIMA_MEM)"

# --- Entorno (Colima) ---------------------------------------------------------------------------

.PHONY: colima-up colima-down colima-status
colima-up: ## Inicia Colima con los recursos configurados
	colima start --cpu $(COLIMA_CPU) --memory $(COLIMA_MEM)

colima-down: ## Detiene Colima (cierra la VM de Docker)
	colima stop

colima-status: ## Estado de Colima
	colima list

# --- Stack completo (Docker Compose) ------------------------------------------------------------

.PHONY: env up up-fg build down clean restart ps logs shell config
env: ## Genera deploy/.env con secretos aleatorios (no sobrescribe uno existente)
	cd deploy && ./compose-env.sh

up: ## Construye si hace falta y levanta el stack en segundo plano (el CMS va aparte: make cms-dev)
	@test -f deploy/.env || $(MAKE) env
	@test -f deploy/secrets/cms_token || echo "Aviso: falta deploy/secrets/cms_token; iluvatar no podrá leer el CMS hasta que ejecutes make cms-seed."
	$(DC) up --build -d

up-fg: ## Levanta todo en primer plano, con los logs en pantalla (Ctrl+C detiene)
	@test -f deploy/.env || $(MAKE) env
	$(DC) up --build

build: ## Construye las imágenes sin levantar (S=<servicio> para una sola)
	$(DC) build $(S)

down: ## Detiene y elimina los contenedores (conserva los datos)
	$(DC) down

clean: ## Detiene y BORRA los datos del compose (la base de Laurelin)
	@read -p "Esto borra la base de datos de Laurelin. ¿Continuar? [s/N] " r && [ "$$r" = "s" ]
	$(DC) down -v

restart: ## Reinicia un servicio: make restart S=iluvatar
	@test -n "$(S)" || { echo "Indica el servicio: make restart S=<servicio>"; exit 1; }
	$(DC) restart $(S)

ps: ## Estado y salud de los contenedores
	$(DC) ps -a

logs: ## Sigue los logs (todos, o uno con S=<servicio>)
	$(DC) logs -f --tail=100 $(S)

shell: ## Abre una shell dentro de un contenedor: make shell S=iluvatar
	@test -n "$(S)" || { echo "Indica el servicio: make shell S=<servicio>"; exit 1; }
	$(DC) exec $(S) sh

config: ## Valida y muestra la configuración resuelta del compose
	$(DC) config

# --- Atajos ------------------------------------------------------------------------------------

.PHONY: urls health stop-all
urls: ## Direcciones locales del stack
	@echo "PWA            http://localhost:3000"
	@echo "iluvatar       http://localhost:8080  (Swagger: /swagger/index.html)"
	@echo "Laurelin       http://localhost:8090"
	@echo "CMS (panel)    http://localhost:1337/admin  (aparte: make cms-dev)"
	@echo "Emulador Auth  http://localhost:9099  (solo con PROFILE=emulator)"

health: ## Comprueba las sondas de salud de los servicios
	@for u in 8080/ready 8090/ready 3000/ 1337/_health; do \
		printf "%-16s" "$$u"; curl -s -o /dev/null -w "%{http_code}\n" http://localhost:$$u || echo "sin respuesta"; \
	done

stop-all: down colima-down ## Detiene el stack y cierra Colima (el CMS se detiene con Ctrl+C donde corre)

# --- CMS (Strapi, fuera del compose) -----------------------------------------------------------
# Primera vez:  make cms-install && make cms-env && make cms-seed   (después, make up y make cms-dev)

.PHONY: cms-install cms-env cms-seed cms-dev cms-build cms-start
cms-install: ## Instala las dependencias del CMS (npm ci)
	@$(NODE24) cd services/iluvatar-cms && npm ci

cms-env: ## Crea services/iluvatar-cms/.env (SQLite, secretos aleatorios, webhook hacia iluvatar)
	@test -f deploy/.env || $(MAKE) env
	cd deploy && ./cms-env.sh

cms-seed: ## Carga datos de prueba y deja el API token en deploy/secrets/cms_token (el CMS debe estar detenido)
	@$(NODE24) cd services/iluvatar-cms && SEED_TOKEN_FILE="$(CURDIR)/deploy/secrets/cms_token" npm run seed

cms-dev: ## Ejecuta el CMS en modo desarrollo (http://localhost:1337/admin, recarga en caliente)
	@$(NODE24) cd services/iluvatar-cms && npm run develop

cms-build: ## Compila el panel del CMS (necesita memoria: es lo que pesaba en Docker)
	@$(NODE24) cd services/iluvatar-cms && npm run build

cms-start: ## Ejecuta el CMS compilado, sin recarga (antes: make cms-build)
	@$(NODE24) cd services/iluvatar-cms && npm run start

# --- Documentación ------------------------------------------------------------------------

.PHONY: diagrams
diagrams: ## Regenera los diagramas de docs/architecture desde fuentes/*.json (requiere la skill Archify)
	@$(NODE24) for f in docs/architecture/fuentes/*.json; do \
		b=$$(basename $$f .json); t=$${b##*.}; n=$${b%.*}; \
		echo "==> $$n ($$t)"; \
		node "$$HOME/.claude/skills/archify/bin/archify.mjs" finalize $$t $$f docs/architecture/$$n.html --quality showcase --out-dir .archify/$$n --json >/dev/null || { echo "Falló $$n; revisa .archify/$$n"; exit 1; }; \
	done

# --- Código: Go, PWA ----------------------------------------------------------------------

.PHONY: test test-go test-pwa vet fmt lint-go install-pwa
test: test-go test-pwa ## Corre todos los tests unitarios (Go y PWA)

test-go: ## go build, go vet y go test -race en iluvatar y laurelin
	@for d in $(GO_SERVICES); do \
		echo "==> $$d"; \
		(cd $$d && go build ./... && go vet ./... && go test -race ./...) || exit 1; \
	done

test-pwa: ## Tests de la PWA (vitest)
	@$(NODE24) cd apps/manwe-pwa && npm test

vet: ## go vet en los servicios Go
	@for d in $(GO_SERVICES); do (cd $$d && go vet ./...) || exit 1; done

fmt: ## gofmt en los servicios Go
	@for d in $(GO_SERVICES); do (cd $$d && gofmt -w src); done

lint-go: ## Falla si algún archivo Go no está formateado
	@out=$$(for d in $(GO_SERVICES); do gofmt -l $$d/src; done); \
	if [ -n "$$out" ]; then echo "Sin formatear:"; echo "$$out"; exit 1; fi

install-pwa: ## Instala las dependencias de la PWA (npm ci)
	@$(NODE24) cd apps/manwe-pwa && npm ci
