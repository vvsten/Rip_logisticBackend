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
	"rip-go-app/internal/app/config"
	"rip-go-app/internal/app/dsn"
	"rip-go-app/internal/app/handler"
	"rip-go-app/internal/app/repository"
	"rip-go-app/internal/app/auth"
	"rip-go-app/internal/app/service"
	"rip-go-app/internal/app/middleware"
	
	// Swagger imports
	_ "rip-go-app/docs"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
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

	// Инициализируем JWT сервис
	jwtService := auth.NewJWTService(
		conf.JWTSecret,
		conf.JWTAccessTokenExpire,
		conf.JWTRefreshTokenExpire,
	)

	// Инициализируем сервис авторизации
	authService := service.NewAuthService(repo, jwtService)

	// Инициализируем middleware авторизации
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Создаем хендлер
	handler := handler.NewHandler(repo, authService, authMiddleware)

	// Создаем роутер
	r := gin.Default()

	// Настраиваем CORS для работы с Tauri и веб-версией
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Разрешаем все источники (для Tauri и веб)
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
				"error": "Route not found",
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
				MinVersion:    tls.VersionTLS12,
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
		if err := r.Run(serverAddress); err != nil {
			logrus.Fatal(err)
		}
	}
	logrus.Info("Application terminated")
}

func registerRoutes(r *gin.Engine, handler *handler.Handler) {
	// HTML страницы (доменные)
	r.GET("/", handler.GetTransportServicesPage)                                 // Каталог транспортных услуг
	r.GET("/transport-services/:id", handler.GetTransportServicePage)            // Страница транспортной услуги
	r.GET("/logistic-request", handler.GetLogisticRequestDetailsPage)            // Демо-страница деталей заявки
	// Страница расчёта грузоперевозки (quote)
	r.GET("/logistic-request/quote", handler.GetDeliveryQuotePage)
	r.POST("/logistic-request/quote", handler.PostDeliveryQuote)
	// Алиас для совместимости
	r.GET("/delivery-quote", handler.GetDeliveryQuotePage)
	r.POST("/delivery-quote", handler.PostDeliveryQuote)

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

    // Авторизация
    r.POST("/sign_up", handler.RegisterUser)
    r.POST("/login", handler.LoginUser)
    r.POST("/logout", handler.AuthMiddleware.RequireAuth(), handler.LogoutUser)
    r.POST("/refresh", handler.RefreshToken)

    // Пользователи (требуют авторизации)›
    authGroup := r.Group("/api/users")
    authGroup.Use(handler.AuthMiddleware.RequireAuth())
    {
        authGroup.GET("/profile", handler.GetUserProfile)
        authGroup.PUT("/profile", handler.UpdateUserProfile)
    }

    // Логистические заявки (требуют авторизации)
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
    // Завершение логистической заявки (модератор)
    moderatorLR := r.Group("/api/logistic-requests/:id")
    moderatorLR.Use(handler.AuthMiddleware.RequireModerator())
    {
        moderatorLR.PUT("/complete", handler.CompleteLogisticRequest)
    }

    // Статус логистической заявки через курсор
    r.PUT("/api/logistic-requests/:id/status", handler.UpdateLogisticRequestStatus)

    // Swagger документация
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}