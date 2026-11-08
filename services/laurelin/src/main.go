package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"laurelin/src/application/usecase"
	firebaseAdapter "laurelin/src/infrastructure/firebase"
	"laurelin/src/infrastructure/http/handlers/rest/health"
	customMiddleware "laurelin/src/infrastructure/http/handlers/rest/middleware"
	serviceHandler "laurelin/src/infrastructure/http/handlers/rest/service"
	v1 "laurelin/src/infrastructure/http/handlers/rest/v1"
	"laurelin/src/infrastructure/http/server"
	"laurelin/src/infrastructure/postgres"

	"github.com/labstack/echo/v4"
)

func main() {
	port := listenPort()
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck(port))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	config := server.ConfigFromEnv()
	// Laurelin no publica Swagger: con esto se aplica siempre la política de contenido restrictiva.
	config.SwaggerEnabled = false

	if err := run(context.Background(), config, logger, port); err != nil {
		logger.Error("the service can't start", "error", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, config server.Config, logger *slog.Logger, port string) error {
	serviceToken := readServiceToken()
	if config.Production() && serviceToken == "" {
		return errors.New("SERVICE_TOKEN (or SERVICE_TOKEN_FILE) is required in production")
	}

	pool, err := postgres.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := postgres.Migrate(ctx, pool); err != nil {
		return err
	}

	firebaseApp, err := firebaseAdapter.NewApp(ctx, os.Getenv("FIREBASE_PROJECT_ID"))
	if err != nil {
		return err
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		return errors.New("can't create the firebase auth client: " + err.Error())
	}

	institutionsRepo := postgres.NewInstitutionRepository(pool)
	programsRepo := postgres.NewProgramRepository(pool)
	studentsRepo := postgres.NewStudentRepository(pool)
	adminsRepo := postgres.NewAdminRepository(pool)
	auditRepo := postgres.NewAuditRepository(pool)

	admins := usecase.NewAdmins(adminsRepo, institutionsRepo, auditRepo)
	students := usecase.NewStudents(studentsRepo, programsRepo, auditRepo)

	// Sin esto nadie podría crear el primer administrador. Es idempotente.
	if email := os.Getenv("BOOTSTRAP_ADMIN_EMAIL"); email != "" {
		if err := admins.Bootstrap(ctx, email); err != nil {
			return errors.New("can't create the bootstrap administrator: " + err.Error())
		}
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	server.Apply(e, config, logger)

	health.NewHealthHandler(e, health.Check{Name: "database", Run: pool.Ping})

	auth := customMiddleware.RequireAdmin(firebaseAdapter.NewAuthVerifier(authClient), admins)
	v1.Register(e, auth, v1.Dependencies{
		Institutions: usecase.NewInstitutions(institutionsRepo, auditRepo),
		Programs:     usecase.NewPrograms(programsRepo, institutionsRepo, auditRepo),
		Students:     students,
		Admins:       admins,
	})
	serviceHandler.Register(e, customMiddleware.RequireServiceToken(serviceToken), students)

	quit := make(chan os.Signal, 1)
	go startServer(e, port, quit)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	gracefulShutdown(e)
	return nil
}

// readServiceToken devuelve el token que usan los otros servicios: SERVICE_TOKEN o el contenido de SERVICE_TOKEN_FILE.
func readServiceToken() string {
	if token := strings.TrimSpace(os.Getenv("SERVICE_TOKEN")); token != "" {
		return token
	}
	path := os.Getenv("SERVICE_TOKEN_FILE")
	if path == "" {
		return ""
	}
	content, err := os.ReadFile(path)
	if err != nil {
		slog.Error("can't read the service token file", "error", err.Error())
		return ""
	}
	return strings.TrimSpace(string(content))
}

func listenPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8090"
}

func startServer(e *echo.Echo, port string, quit chan os.Signal) {
	slog.Info("starting server", "port", port)
	if err := e.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("the server stopped", "error", err.Error())
		close(quit)
	}
}

func gracefulShutdown(e *echo.Echo) {
	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err.Error())
	}
}
