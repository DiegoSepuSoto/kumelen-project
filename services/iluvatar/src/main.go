package main

import (
	"context"
	"errors"
	firebaseSDK "firebase.google.com/go/v4"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "iluvatar/docs"
	catalogUseCase "iluvatar/src/application/usecase/catalog"
	categoryUseCase "iluvatar/src/application/usecase/category"
	homeUseCase "iluvatar/src/application/usecase/home"
	notificationUseCase "iluvatar/src/application/usecase/notification"
	postUseCase "iluvatar/src/application/usecase/post"
	pushUseCase "iluvatar/src/application/usecase/push"
	serviceUseCase "iluvatar/src/application/usecase/service"
	"iluvatar/src/domain/repository"
	firebaseAdapter "iluvatar/src/infrastructure/firebase"
	"iluvatar/src/infrastructure/http/handlers/rest/health"
	customMiddlewares "iluvatar/src/infrastructure/http/handlers/rest/middleware"
	catalogHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/catalog"
	categoryHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/category"
	homeHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/home"
	hooksHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/hooks"
	notificationHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/notification"
	postHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/post"
	pushHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/push"
	serviceHandler "iluvatar/src/infrastructure/http/handlers/rest/v1/service"
	"iluvatar/src/infrastructure/http/server"
	"iluvatar/src/infrastructure/laurelin"
	"iluvatar/src/infrastructure/strapi"
	"iluvatar/src/shared/validations"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// @title Documentación Artefacto API
// @version 1.0
// @description En esta documentación se encuentran los detalles de los endpoints presentes en el artefacto API del proyecto Kümelen
// @contact.name Diego Sepúlveda
// @contact.email diego.sepulvedas@utem.cl
func main() {
	port := listenPort()
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck(port))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	config := server.ConfigFromEnv()

	if err := run(context.Background(), config, logger, port); err != nil {
		logger.Error("the service can't start", "error", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, config server.Config, logger *slog.Logger, port string) error {
	e := echo.New()
	e.Validator = validations.NewCustomValidator(validator.New())
	e.HideBanner = true
	e.HidePort = true
	server.Apply(e, config, logger)

	if config.SwaggerEnabled {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// La configuración insegura se rechaza antes de conectar con nada.
	if parseBool(os.Getenv("PUSH_DISABLED")) && config.Production() {
		return errors.New("PUSH_DISABLED can't be used in production")
	}

	firebaseApp, err := firebaseAdapter.NewApp(ctx, os.Getenv("FIREBASE_PROJECT_ID"))
	if err != nil {
		return err
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		return errors.New("can't create the firebase auth client: " + err.Error())
	}
	pushRepository, err := newPushRepository(ctx, firebaseApp)
	if err != nil {
		return err
	}

	cmsClient := newCMSClient()
	health.NewHealthHandler(e, health.Check{Name: "cms", Run: cmsClient.Ping})

	authMiddleware := customMiddlewares.RequireAuth(firebaseAdapter.NewAuthVerifier(authClient))
	studentGate, err := newStudentGate(config)
	if err != nil {
		return err
	}
	if studentGate != nil {
		authMiddleware = chain(authMiddleware, studentGate)
	}
	pushUseCaseImpl := pushUseCase.NewPushNotificationUseCase(pushRepository)

	_ = hooksHandler.NewHooksHandler(e, pushUseCaseImpl)
	_ = pushHandler.NewPushHandler(e, pushUseCaseImpl, authMiddleware)
	registerBFFHandlers(e, authMiddleware, cmsClient)

	quit := make(chan os.Signal, 1)
	go startServer(e, port, quit)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	gracefulShutdown(e)
	return nil
}

// newPushRepository crea el envío por FCM. PUSH_DISABLED=true lo reemplaza por uno que no envía,
// solo para desarrollo sin credenciales (run() lo rechaza en producción).
func newPushRepository(ctx context.Context, app *firebaseSDK.App) (repository.PushNotificationRepository, error) {
	if parseBool(os.Getenv("PUSH_DISABLED")) {
		slog.Warn("push notifications are disabled")
		return firebaseAdapter.NewDisabledPushRepository(), nil
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, errors.New("can't create the firebase messaging client (check GOOGLE_APPLICATION_CREDENTIALS): " + err.Error())
	}
	return firebaseAdapter.NewPushRepository(messagingClient), nil
}

// newCMSClient crea el cliente del CMS con su caché. CMS_CACHE_TTL (segundos, 30 por defecto; 0 la desactiva)
// define cuánto dura una respuesta fresca; ante una caída del CMS se sirve la última respuesta buena hasta 24 horas.
// CMS_PUBLIC_URL es la dirección del CMS que ve el navegador (imágenes) cuando difiere de ILUVATAR_CMS_HOST.
func newCMSClient() *strapi.Client {
	ttl := 30 * time.Second
	if seconds, err := strconv.Atoi(os.Getenv("CMS_CACHE_TTL")); err == nil && seconds >= 0 {
		ttl = time.Duration(seconds) * time.Second
	}
	return strapi.NewClient(
		os.Getenv("ILUVATAR_CMS_HOST"),
		cmsToken(),
		&http.Client{Timeout: 10 * time.Second},
		strapi.WithCache(ttl, 24*time.Hour, 500),
		strapi.WithPublicURL(os.Getenv("CMS_PUBLIC_URL")),
	)
}

// cmsToken devuelve el API token del CMS: ILUVATAR_TOKEN o, si no está, el contenido del archivo
// ILUVATAR_TOKEN_FILE (secretos montados, o el token que escribe el seed del CMS en desarrollo).
func cmsToken() string { return secretFromEnv("ILUVATAR_TOKEN") }

// secretFromEnv devuelve el valor de la variable name o, si está vacía, el contenido del archivo que indica name+"_FILE".
func secretFromEnv(name string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	path := os.Getenv(name + "_FILE")
	if path == "" {
		return ""
	}
	content, err := os.ReadFile(path)
	if err != nil {
		slog.Error("can't read the secret file", "variable", name+"_FILE", "error", err.Error())
		return ""
	}
	return strings.TrimSpace(string(content))
}

// newStudentGate crea la verificación de que quien entra es un estudiante activo (Laurelin).
// Con LAURELIN_URL la activa; LAURELIN_TOKEN (o LAURELIN_TOKEN_FILE) es el token de servicio.
// Sin LAURELIN_URL cualquier cuenta de Firebase con token válido entra, así que en producción eso exige
// STUDENT_CHECK_DISABLED=true de forma explícita. Devuelve nil cuando la verificación está desactivada.
func newStudentGate(config server.Config) (echo.MiddlewareFunc, error) {
	url := strings.TrimSpace(os.Getenv("LAURELIN_URL"))
	if url == "" {
		if config.Production() && !parseBool(os.Getenv("STUDENT_CHECK_DISABLED")) {
			return nil, errors.New("LAURELIN_URL is required in production (or set STUDENT_CHECK_DISABLED=true to accept any Firebase account)")
		}
		slog.Warn("student verification is disabled: any valid Firebase account can use the API")
		return nil, nil
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, errors.New("LAURELIN_URL must start with http:// or https://")
	}
	token := secretFromEnv("LAURELIN_TOKEN")
	if token == "" {
		return nil, errors.New("LAURELIN_TOKEN (or LAURELIN_TOKEN_FILE) is required when LAURELIN_URL is set")
	}

	// Un estudiante verificado vale 10 minutos; uno desconocido, 1 minuto (un alta nueva se nota pronto);
	// si Laurelin cae, quien ya fue verificado sigue entrando hasta 24 horas.
	directory := laurelin.NewClient(url, token, &http.Client{Timeout: 10 * time.Second},
		laurelin.WithCache(10*time.Minute, time.Minute, 24*time.Hour, 20000))
	return customMiddlewares.RequireStudent(directory), nil
}

// chain aplica los middlewares en orden: el primero se ejecuta antes.
func chain(first echo.MiddlewareFunc, rest ...echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		for i := len(rest) - 1; i >= 0; i-- {
			next = rest[i](next)
		}
		return first(next)
	}
}

func listenPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

func parseBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	return err == nil && parsed
}

// registerBFFHandlers monta las pantallas de la app (antes el servicio "elentari").
// Todas exigen un ID token de Firebase válido.
func registerBFFHandlers(e *echo.Echo, authMiddleware echo.MiddlewareFunc, cmsClient *strapi.Client) {
	serviceRepositoryImpl := strapi.NewServiceRepository(cmsClient)
	postRepositoryImpl := strapi.NewPostRepository(cmsClient)
	categoryRepositoryImpl := strapi.NewCategoryRepository(cmsClient)
	notificationRepositoryImpl := strapi.NewNotificationRepository(cmsClient)

	homeHandler.NewHomeHandler(e, homeUseCase.NewHomeUseCase(serviceRepositoryImpl), authMiddleware)
	postHandler.NewPostHandler(e, postUseCase.NewPostUseCase(postRepositoryImpl), authMiddleware)
	serviceHandler.NewServiceHandler(e, serviceUseCase.NewServiceUseCase(serviceRepositoryImpl), authMiddleware)
	catalogHandler.NewCatalogHandler(e, catalogUseCase.NewCatalogUseCase(serviceRepositoryImpl, categoryRepositoryImpl), authMiddleware)
	categoryHandler.NewCategoryHandler(e, categoryUseCase.NewCategoryUseCase(categoryRepositoryImpl), authMiddleware)
	notificationHandler.NewNotificationHandler(e, notificationUseCase.NewNotificationUseCase(notificationRepositoryImpl), authMiddleware)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err.Error())
	}
}
