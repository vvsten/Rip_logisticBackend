// @title RIP Go API
// @version 1.0
// @description API for cargo transportation service
// @host localhost:8083
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token for authentication. Format: 'Bearer <token>'
package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"rip-go-app/internal/app/auth"
	"rip-go-app/internal/app/config"
	"rip-go-app/internal/app/dsn"
	"rip-go-app/internal/app/handler"
	"rip-go-app/internal/app/middleware"
	"rip-go-app/internal/app/repository"
	"rip-go-app/internal/app/service"

	// Swagger imports
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "rip-go-app/docs"
)

func main() {
	logrus.Info("Application start up")

	// Загружаем конфигурацию
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Получаем строку подключения к БД
	postgresString := dsn.FromEnv()
	fmt.Println("Connecting to database with DSN:", postgresString)

	// Инициализируем репозиторий
	repo, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	// Инициализируем Redis + session-id авторизацию (без JWT)
	redisService := auth.NewRedisService(conf.RedisHost, conf.RedisPort, conf.RedisPassword, conf.RedisDB)
	if err := redisService.Ping(); err != nil {
		logrus.Fatalf("error connecting to redis: %v", err)
	}
	defer func() { _ = redisService.Close() }()

	accessTTL := time.Duration(conf.JWTAccessTokenExpire) * time.Minute
	refreshTTL := time.Duration(conf.JWTRefreshTokenExpire) * 24 * time.Hour
	sessionService := auth.NewSessionService(redisService, accessTTL, refreshTTL)

	// Инициализируем сервис авторизации
	authService := service.NewAuthService(repo, sessionService)

	// Инициализируем middleware авторизации
	authMiddleware := middleware.NewAuthMiddleware(sessionService)

	// Создаем хендлер
	handler := handler.NewHandler(repo, authService, authMiddleware, conf.AsyncServiceURL, conf.AsyncServiceToken, conf.PublicBaseURL)

	// Создаем роутер
	r := gin.Default()

	// Настраиваем CORS для работы с Tauri и веб-версией
	r.Use(cors.New(cors.Config{
		// ВАЖНО: нельзя сочетать AllowCredentials=true с Allow-Origin="*".
		// Для Tauri/Web/PWA разрешаем любые origin, но отражаем конкретный origin в ответе.
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Регистрируем статические файлы и шаблоны
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "static")

	// Прокси для MinIO изображений
	r.Any("/lab1/*path", func(c *gin.Context) {
		path := c.Param("path")
		minioURL := fmt.Sprintf("http://localhost:9003/lab1%s", path)

		// Создаем запрос к MinIO
		req, err := http.NewRequest(c.Request.Method, minioURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Копируем заголовки
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Выполняем запрос
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to MinIO"})
			return
		}
		defer resp.Body.Close()

		// Копируем заголовки ответа
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Устанавливаем статус код
		c.Status(resp.StatusCode)

		// Копируем тело ответа
		io.Copy(c.Writer, resp.Body)
	})

	// Регистрируем маршруты
	registerRoutes(r, handler)

	// Обработчик для неизвестных маршрутов (SPA fallback)
	// Игнорируем запросы к фронтенд маршрутам, которые должны обрабатываться React Router
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Если это API, статика, swagger или известные бэкенд роуты - возвращаем 404
		if strings.HasPrefix(path, "/api") ||
			strings.HasPrefix(path, "/static") ||
			strings.HasPrefix(path, "/lab1") ||
			strings.HasPrefix(path, "/swagger") ||
			path == "/logistic-request" ||
			path == "/logistic-request/quote" ||
			path == "/delivery-quote" ||
			strings.HasPrefix(path, "/transport-services/") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Route not found",
				"message": "This route should be handled",
			})
			return
		}

		// Для фронтенд роутов (/, /transport-services, /about) - не возвращаем ошибку
		// React Router обработает их на клиенте
		// Просто возвращаем пустой ответ, чтобы не мешать SPA роутингу
		c.Status(http.StatusOK)
		c.Writer.WriteHeaderNow()
	})

	// Запускаем сервер
	serverAddress := fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort)

	if conf.EnableHTTPS {
		logrus.Infof("Starting HTTPS server on %s", serverAddress)

		// Загружаем сертификат для проверки
		cert, err := tls.LoadX509KeyPair(conf.CertFile, conf.KeyFile)
		if err != nil {
			logrus.Fatalf("Failed to load certificate: %v", err)
		}
		logrus.Infof("Certificate loaded successfully from %s", conf.CertFile)

		// Создаем HTTP сервер с упрощенной TLS конфигурацией
		// Используем минимальную конфигурацию для максимальной совместимости
		srv := &http.Server{
			Addr:    serverAddress,
			Handler: r,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS12,
				MaxVersion:   tls.VersionTLS13,
				// Не ограничиваем cipher suites - пусть Go выберет автоматически
				// Это обеспечит лучшую совместимость с разными клиентами
			},
			// Увеличиваем таймауты для TLS handshake
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		// Запускаем HTTPS сервер
		logrus.Info("HTTPS server is ready to accept connections")
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start HTTPS server: %v", err)
		}
	} else {
		logrus.Infof("Starting HTTP server on %s", serverAddress)
		// Создаем HTTP сервер с увеличенными таймаутами для long polling
		srv := &http.Server{
			Addr:         serverAddress,
			Handler:      r,
			ReadTimeout:  60 * time.Second,  // Увеличиваем для long polling
			WriteTimeout: 60 * time.Second,  // Увеличиваем для long polling
			IdleTimeout:  120 * time.Second,
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal(err)
		}
	}
	logrus.Info("Application terminated")
}

func registerRoutes(r *gin.Engine, handler *handler.Handler) {
	const apiVersion = "v1"
	const apiV1Prefix = "/api/" + apiVersion

	// HTML страницы (доменные)
	r.GET("/", handler.GetTransportServicesPage)                      // Каталог транспортных услуг
	r.GET("/transport-services/:id", handler.GetTransportServicePage) // Страница транспортной услуги
	r.GET("/logistic-request", handler.GetLogisticRequestDetailsPage) // Демо-страница деталей заявки
	// Страница расчёта грузоперевозки (quote)
	r.GET("/logistic-request/quote", handler.GetDeliveryQuotePage)
	r.POST("/logistic-request/quote", handler.PostDeliveryQuote)
	// Алиас для совместимости
	r.GET("/delivery-quote", handler.GetDeliveryQuotePage)
	r.POST("/delivery-quote", handler.PostDeliveryQuote)

	// ===================== Версионированный REST API =====================
	apiV1 := r.Group(apiV1Prefix)
	{
		// Auth (оставляем ваши названия эндпоинтов)
		apiV1.POST("/sign_up", handler.RegisterUser)
		apiV1.POST("/login", handler.LoginUser)
		apiV1.POST("/logout", handler.AuthMiddleware.RequireAuth(), handler.LogoutUser)
		apiV1.POST("/refresh", handler.RefreshToken)

		users := apiV1.Group("/users")
		users.Use(handler.AuthMiddleware.RequireAuth())
		{
			users.GET("/profile", handler.GetUserProfile)
			users.PUT("/profile", handler.UpdateUserProfile)
		}

		// Transport services (resources)
		apiV1.GET("/transport-services", handler.GetTransportServices)
		apiV1.GET("/transport-services/:id", handler.GetTransportService)
		apiV1.POST("/transport-services", handler.CreateTransportService)
		apiV1.PUT("/transport-services/:id", handler.UpdateTransportService)
		apiV1.DELETE("/transport-services/:id", handler.DeleteTransportService)
		// Поиск оставляем отдельным endpoint (исторически POST)
		apiV1.POST("/transport-services/search", handler.SearchTransportServices)

		// Logistic requests (resources)
		// Guest draft (без авторизации)
		apiV1.POST("/logistic-requests/draft/services/:service_id", handler.AddTransportServiceToDraftLogisticRequest)
		apiV1.DELETE("/logistic-requests/draft", handler.ClearDraftLogisticRequest)
		apiV1.GET("/logistic-requests/draft", handler.GetDraftLogisticRequest)
		apiV1.GET("/logistic-requests/draft/count", handler.GetDraftLogisticRequestServiceCount)
		apiV1.GET("/logistic-requests/draft/icon", handler.GetDraftLogisticRequestIcon)

		// Quote
		apiV1.POST("/logistic-requests/quote", handler.CalculateLogisticRequestQuote)

		logistic := apiV1.Group("/logistic-requests")
		logistic.Use(handler.AuthMiddleware.RequireAuth())
		{
			// Черновик заявок авторизованного пользователя (для React UI)
			logistic.GET("/user-draft/icon", handler.GetUserDraftIcon)
			logistic.POST("/user-draft/services/:service_id", handler.AddTransportServiceToUserDraft)
			logistic.DELETE("/user-draft", handler.ClearUserDraftLogisticRequest)

			logistic.POST("", handler.CreateCargoLogisticRequest)
			logistic.GET("", handler.GetLogisticRequests)
			logistic.GET("/long-poll", handler.GetLogisticRequestsLongPoll) // Long polling для списка
			logistic.GET("/:id", handler.GetLogisticRequest)
			logistic.GET("/:id/long-poll", handler.GetLogisticRequestLongPoll) // Long polling для одной заявки
			logistic.DELETE("/:id", handler.DeleteLogisticRequest)
			logistic.PUT("/:id/form", handler.FormLogisticRequest)
			logistic.PUT("/:id/update", handler.UpdateLogisticRequest)
			logistic.DELETE("/:id/services/:service_id", handler.RemoveServiceFromLogisticRequest)
			logistic.PUT("/:id/services/:service_id", handler.UpdateLogisticRequestService)
			logistic.PUT("/:id/status", handler.UpdateLogisticRequestStatus)
		}

		// Moderator actions
		moderator := apiV1.Group("/logistic-requests/:id")
		moderator.Use(handler.AuthMiddleware.RequireModerator())
		{
			moderator.PUT("/complete", handler.CompleteLogisticRequest)
			moderator.POST("/async/start", handler.StartAsyncProcessingForRequest)
		}

		// Internal (async-service callbacks)
		internal := apiV1.Group("/internal")
		{
			internal.POST("/logistic-requests/:id/services/:service_id/result", handler.SetAsyncServiceResultIfEmpty)
			internal.PUT("/logistic-requests/:id/services/:service_id/result", handler.ForceSetAsyncServiceResult)
		}
	}

	// ===================== Legacy aliases (совместимость) =====================
	// Черновик логистической заявки (guest) — бывшая "корзина"
	r.POST("/api/logistic-requests/draft/services/:service_id", handler.AddTransportServiceToDraftLogisticRequest)
	r.DELETE("/api/logistic-requests/draft", handler.ClearDraftLogisticRequest)
	r.GET("/api/logistic-requests/draft", handler.GetDraftLogisticRequest)
	r.GET("/api/logistic-requests/draft/count", handler.GetDraftLogisticRequestServiceCount)
	r.GET("/api/logistic-requests/draft/icon", handler.GetDraftLogisticRequestIcon)

	// Доменные API операции под грузоперевозки
	r.POST("/api/transport-services/search", handler.SearchTransportServices)
	r.POST("/api/logistic-requests/quote", handler.CalculateLogisticRequestQuote)

	// CRUD JSON для транспортных услуг
	r.GET("/api/transport-services", handler.GetTransportServices)
	r.GET("/api/transport-services/:id", handler.GetTransportService)
	r.POST("/api/transport-services", handler.CreateTransportService)
	r.PUT("/api/transport-services/:id", handler.UpdateTransportService)
	r.DELETE("/api/transport-services/:id", handler.DeleteTransportService)

	// Авторизация (старые URL)
	r.POST("/sign_up", handler.RegisterUser)
	r.POST("/login", handler.LoginUser)
	r.POST("/logout", handler.AuthMiddleware.RequireAuth(), handler.LogoutUser)
	r.POST("/refresh", handler.RefreshToken)

	// Пользователи (требуют авторизации) — старые URL
	authGroup := r.Group("/api/users")
	authGroup.Use(handler.AuthMiddleware.RequireAuth())
	{
		authGroup.GET("/profile", handler.GetUserProfile)
		authGroup.PUT("/profile", handler.UpdateUserProfile)
	}

	// Логистические заявки (требуют авторизации) — старые URL
	logisticGroup := r.Group("/api/logistic-requests")
	logisticGroup.Use(handler.AuthMiddleware.RequireAuth())
	{
		// Черновик заявок авторизованного пользователя (для React UI)
		logisticGroup.GET("/user-draft/icon", handler.GetUserDraftIcon)
		logisticGroup.POST("/user-draft/services/:service_id", handler.AddTransportServiceToUserDraft)
		logisticGroup.DELETE("/user-draft", handler.ClearUserDraftLogisticRequest)

		logisticGroup.POST("", handler.CreateCargoLogisticRequest)
		logisticGroup.GET("", handler.GetLogisticRequests)
		logisticGroup.GET("/:id", handler.GetLogisticRequest)
		logisticGroup.DELETE("/:id", handler.DeleteLogisticRequest)
		logisticGroup.PUT("/:id/form", handler.FormLogisticRequest)
		logisticGroup.PUT("/:id/update", handler.UpdateLogisticRequest)
		logisticGroup.DELETE("/:id/services/:service_id", handler.RemoveServiceFromLogisticRequest)
		logisticGroup.PUT("/:id/services/:service_id", handler.UpdateLogisticRequestService)
	}
	// Завершение логистической заявки (модератор) — старые URL
	moderatorLR := r.Group("/api/logistic-requests/:id")
	moderatorLR.Use(handler.AuthMiddleware.RequireModerator())
	{
		moderatorLR.PUT("/complete", handler.CompleteLogisticRequest)
	}

	// Lab8: внутренние endpoints для async сервиса (псевдо-авторизация по токену) — старые URL
	r.POST("/api/internal/logistic-requests/:id/services/:service_id/result", handler.SetAsyncServiceResultIfEmpty)
	r.PUT("/api/internal/logistic-requests/:id/services/:service_id/result", handler.ForceSetAsyncServiceResult)
	// Также возможность дернуть async-сервис из основного (удобно для UI)
	r.POST("/api/logistic-requests/:id/async/start", handler.AuthMiddleware.RequireModerator(), handler.StartAsyncProcessingForRequest)

	// Статус логистической заявки через курсор — старые URL
	r.PUT("/api/logistic-requests/:id/status", handler.UpdateLogisticRequestStatus)

	// Swagger документация
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
