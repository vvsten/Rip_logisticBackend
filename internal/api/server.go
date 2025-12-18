package api

import (
	"fmt"
	"log"

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

	jwtService := auth.NewJWTService(
		conf.JWTSecret,
		conf.JWTAccessTokenExpire,
		conf.JWTRefreshTokenExpire,
	)

	authService := service.NewAuthService(repo, jwtService)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	h := handler.NewHandler(repo, authService, authMiddleware)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
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

	// Статус заявки
	r.PUT("/api/logistic-requests/:id/status", h.UpdateLogisticRequestStatus)

	serverAddress := fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort)
	r.Run(serverAddress)
	log.Println("Server down")
}