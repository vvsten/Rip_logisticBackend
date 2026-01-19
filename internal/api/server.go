package api

import (
	"fmt"
	"log"
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
)

func StartServer() {
	log.Println("Starting server")

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()

	repo, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	redisService := auth.NewRedisService(conf.RedisHost, conf.RedisPort, conf.RedisPassword, conf.RedisDB)
	if err := redisService.Ping(); err != nil {
		logrus.Fatalf("error connecting to redis: %v", err)
	}
	defer func() { _ = redisService.Close() }()

	accessTTL := time.Duration(conf.JWTAccessTokenExpire) * time.Minute
	refreshTTL := time.Duration(conf.JWTRefreshTokenExpire) * 24 * time.Hour
	sessionService := auth.NewSessionService(redisService, accessTTL, refreshTTL)

	authService := service.NewAuthService(repo, sessionService)
	authMiddleware := middleware.NewAuthMiddleware(sessionService)

	h := handler.NewHandler(repo, authService, authMiddleware, conf.AsyncServiceURL, conf.AsyncServiceToken, conf.PublicBaseURL)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		// ВАЖНО: нельзя сочетать AllowCredentials=true с Allow-Origin="*".
		// Разрешаем любые origin, но отражаем конкретный origin в ответе.
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	// добавляем наш html/шаблон
	r.LoadHTMLGlob("templates/*.html")
	// добавляем статические файлы (CSS, JS, изображения)
	r.Static("/static", "static")

	// HTML страницы (доменные)
	r.GET("/", h.GetTransportServicesPage)
	r.GET("/transport-services/:id", h.GetTransportServicePage)
	r.GET("/logistic-request", h.GetLogisticRequestDetailsPage)
	// Страница расчёта грузоперевозки (quote)
	r.GET("/logistic-request/quote", h.GetDeliveryQuotePage)
	r.POST("/logistic-request/quote", h.PostDeliveryQuote)
	// Алиас для совместимости
	r.GET("/delivery-quote", h.GetDeliveryQuotePage)
	r.POST("/delivery-quote", h.PostDeliveryQuote)

	// ===================== Версионированный REST API =====================
	const apiVersion = "v1"
	const apiV1Prefix = "/api/" + apiVersion
	apiV1 := r.Group(apiV1Prefix)
	{
		// Auth (оставляем ваши названия эндпоинтов)
		apiV1.POST("/sign_up", h.RegisterUser)
		apiV1.POST("/login", h.LoginUser)
		apiV1.POST("/logout", h.AuthMiddleware.RequireAuth(), h.LogoutUser)
		apiV1.POST("/refresh", h.RefreshToken)

		users := apiV1.Group("/users")
		users.Use(h.AuthMiddleware.RequireAuth())
		{
			users.GET("/profile", h.GetUserProfile)
			users.PUT("/profile", h.UpdateUserProfile)
		}

		// Transport services
		apiV1.GET("/transport-services", h.GetTransportServices)
		apiV1.GET("/transport-services/:id", h.GetTransportService)
		apiV1.POST("/transport-services", h.CreateTransportService)
		apiV1.PUT("/transport-services/:id", h.UpdateTransportService)
		apiV1.DELETE("/transport-services/:id", h.DeleteTransportService)
		apiV1.POST("/transport-services/search", h.SearchTransportServices)

		// Logistic requests
		apiV1.POST("/logistic-requests/draft/services/:service_id", h.AddTransportServiceToDraftLogisticRequest)
		apiV1.DELETE("/logistic-requests/draft", h.ClearDraftLogisticRequest)
		apiV1.GET("/logistic-requests/draft", h.GetDraftLogisticRequest)
		apiV1.GET("/logistic-requests/draft/count", h.GetDraftLogisticRequestServiceCount)
		apiV1.GET("/logistic-requests/draft/icon", h.GetDraftLogisticRequestIcon)
		apiV1.POST("/logistic-requests/quote", h.CalculateLogisticRequestQuote)

		logistic := apiV1.Group("/logistic-requests")
		logistic.Use(h.AuthMiddleware.RequireAuth())
		{
			logistic.POST("", h.CreateCargoLogisticRequest)
			logistic.GET("", h.GetLogisticRequests)
			logistic.GET("/:id", h.GetLogisticRequest)
			logistic.DELETE("/:id", h.DeleteLogisticRequest)
			logistic.PUT("/:id/form", h.FormLogisticRequest)
			logistic.PUT("/:id/update", h.UpdateLogisticRequest)
			logistic.DELETE("/:id/services/:service_id", h.RemoveServiceFromLogisticRequest)
			logistic.PUT("/:id/services/:service_id", h.UpdateLogisticRequestService)
			logistic.PUT("/:id/status", h.UpdateLogisticRequestStatus)
		}

		// Moderator actions
		moderator := apiV1.Group("/logistic-requests/:id")
		moderator.Use(h.AuthMiddleware.RequireModerator())
		{
			moderator.PUT("/complete", h.CompleteLogisticRequest)
			moderator.POST("/async/start", h.StartAsyncProcessingForRequest)
		}

		// Internal callbacks
		internal := apiV1.Group("/internal")
		{
			internal.POST("/logistic-requests/:id/services/:service_id/result", h.SetAsyncServiceResultIfEmpty)
			internal.PUT("/logistic-requests/:id/services/:service_id/result", h.ForceSetAsyncServiceResult)
		}
	}

	// Черновик логистической заявки (guest)
	r.POST("/api/logistic-requests/draft/services/:service_id", h.AddTransportServiceToDraftLogisticRequest)
	r.DELETE("/api/logistic-requests/draft", h.ClearDraftLogisticRequest)
	r.GET("/api/logistic-requests/draft", h.GetDraftLogisticRequest)
	r.GET("/api/logistic-requests/draft/count", h.GetDraftLogisticRequestServiceCount)
	r.GET("/api/logistic-requests/draft/icon", h.GetDraftLogisticRequestIcon)

	// Доменные операции
	r.POST("/api/transport-services/search", h.SearchTransportServices)
	r.POST("/api/logistic-requests/quote", h.CalculateLogisticRequestQuote)

	// CRUD transport-services
	r.GET("/api/transport-services", h.GetTransportServices)
	r.GET("/api/transport-services/:id", h.GetTransportService)
	r.POST("/api/transport-services", h.CreateTransportService)
	r.PUT("/api/transport-services/:id", h.UpdateTransportService)
	r.DELETE("/api/transport-services/:id", h.DeleteTransportService)

	// Авторизация
	r.POST("/api/users/register", h.RegisterUser)
	r.POST("/api/users/login", h.LoginUser)
	r.POST("/api/users/logout", h.AuthMiddleware.RequireAuth(), h.LogoutUser)
	r.GET("/api/users/profile", h.AuthMiddleware.RequireAuth(), h.GetUserProfile)
	r.PUT("/api/users/profile", h.AuthMiddleware.RequireAuth(), h.UpdateUserProfile)

	// Логистические заявки (auth)
	lr := r.Group("/api/logistic-requests")
	lr.Use(h.AuthMiddleware.RequireAuth())
	{
		lr.POST("", h.CreateCargoLogisticRequest)
		lr.GET("", h.GetLogisticRequests)
		lr.GET("/:id", h.GetLogisticRequest)
		lr.DELETE("/:id", h.DeleteLogisticRequest)
		lr.PUT("/:id/form", h.FormLogisticRequest)
		lr.PUT("/:id/update", h.UpdateLogisticRequest)
		lr.DELETE("/:id/services/:service_id", h.RemoveServiceFromLogisticRequest)
		lr.PUT("/:id/services/:service_id", h.UpdateLogisticRequestService)
	}

	// Lab8: внутренние endpoints для async сервиса (псевдо-авторизация по токену)
	r.POST("/api/internal/logistic-requests/:id/services/:service_id/result", h.SetAsyncServiceResultIfEmpty)
	r.PUT("/api/internal/logistic-requests/:id/services/:service_id/result", h.ForceSetAsyncServiceResult)
	r.POST("/api/logistic-requests/:id/async/start", h.AuthMiddleware.RequireModerator(), h.StartAsyncProcessingForRequest)

	// Статус заявки
	r.PUT("/api/logistic-requests/:id/status", h.UpdateLogisticRequestStatus)

	serverAddress := fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort)
	r.Run(serverAddress)
	log.Println("Server down")
}

