package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"rip-go-app/internal/app/ds"
	"rip-go-app/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Совместимость: переименовываем старые таблицы/последовательности, если они еще не переименованы
	sqlDB, _ := db.DB()

	// Проверяем и переименовываем таблицы
	var tableExists int
	sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'services' AND table_schema = 'public'`).Scan(&tableExists)
	if tableExists > 0 {
		sqlDB.Exec(`ALTER TABLE services RENAME TO transport_services`)
	}

	sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'orders' AND table_schema = 'public'`).Scan(&tableExists)
	if tableExists > 0 {
		sqlDB.Exec(`ALTER TABLE orders RENAME TO logistic_requests`)
	}

	sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'order_services' AND table_schema = 'public'`).Scan(&tableExists)
	if tableExists > 0 {
		sqlDB.Exec(`ALTER TABLE order_services RENAME TO logistic_request_services`)
	}

	// Переименовываем последовательности
	sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.sequences WHERE sequence_name = 'services_id_seq' AND sequence_schema = 'public'`).Scan(&tableExists)
	if tableExists > 0 {
		sqlDB.Exec(`ALTER SEQUENCE services_id_seq RENAME TO transport_services_id_seq`)
	}

	sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.sequences WHERE sequence_name = 'orders_id_seq' AND sequence_schema = 'public'`).Scan(&tableExists)
	if tableExists > 0 {
		sqlDB.Exec(`ALTER SEQUENCE orders_id_seq RENAME TO logistic_requests_id_seq`)
	}

	sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.sequences WHERE sequence_name = 'order_services_id_seq' AND sequence_schema = 'public'`).Scan(&tableExists)
	if tableExists > 0 {
		sqlDB.Exec(`ALTER SEQUENCE order_services_id_seq RENAME TO logistic_request_services_id_seq`)
	}

	// Переименовываем колонки, если они существуют
	sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'logistic_request_services' AND column_name = 'service_id' AND table_schema = 'public'`).Scan(&tableExists)
	if tableExists > 0 {
		sqlDB.Exec(`ALTER TABLE logistic_request_services RENAME COLUMN service_id TO transport_service_id`)
	}

	sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'logistic_request_services' AND column_name = 'order_id' AND table_schema = 'public'`).Scan(&tableExists)
	if tableExists > 0 {
		sqlDB.Exec(`ALTER TABLE logistic_request_services RENAME COLUMN order_id TO logistic_request_id`)
	}

	// Обновляем DEFAULT для последовательностей
	sqlDB.Exec(`ALTER TABLE IF EXISTS transport_services ALTER COLUMN id SET DEFAULT nextval('transport_services_id_seq')`)
	sqlDB.Exec(`ALTER TABLE IF EXISTS logistic_requests ALTER COLUMN id SET DEFAULT nextval('logistic_requests_id_seq')`)
	sqlDB.Exec(`ALTER TABLE IF EXISTS logistic_request_services ALTER COLUMN id SET DEFAULT nextval('logistic_request_services_id_seq')`)

	// Обновляем существующие записи с NULL creator_id
	db.Model(&ds.LogisticRequest{}).Where("creator_id IS NULL OR creator_id = 0").Updates(map[string]interface{}{
		"creator_id": 1, // системный создатель
		"status":     ds.StatusDraft,
		"is_draft":   true,
	})

	// Migrate the schema
	err = db.AutoMigrate(
		&ds.User{},
		&ds.TransportService{},
		&ds.LogisticRequest{},
		&ds.LogisticRequestService{},
	)
	if err != nil {
		panic("cant migrate db")
	}

	// Индексы под реально используемые запросы репозитория.
	// Важно: делаем idempotent (IF NOT EXISTS), чтобы миграция была безопасной при повторных запусках.
	indexDDLs := []string{
		// logistic_requests: список заявок (GetLogisticRequests)
		// WHERE deleted_at IS NULL AND status <> 'draft' ORDER BY created_at DESC (+ optional status/date filters)
		`CREATE INDEX IF NOT EXISTS idx_logistic_requests_active_created_at_desc
			ON logistic_requests (created_at DESC)
			WHERE deleted_at IS NULL AND status <> 'draft'`,
		`CREATE INDEX IF NOT EXISTS idx_logistic_requests_active_status_created_at_desc
			ON logistic_requests (status, created_at DESC)
			WHERE deleted_at IS NULL AND status <> 'draft'`,
		`CREATE INDEX IF NOT EXISTS idx_logistic_requests_active_formed_at
			ON logistic_requests (formed_at)
			WHERE deleted_at IS NULL AND status <> 'draft' AND formed_at IS NOT NULL`,

		// logistic_requests: черновик пользователя (GetDraftLogisticRequest / GetCartIcon)
		// WHERE creator_id = ? AND status = 'draft' AND deleted_at IS NULL
		`CREATE INDEX IF NOT EXISTS idx_logistic_requests_draft_creator_id
			ON logistic_requests (creator_id)
			WHERE deleted_at IS NULL AND status = 'draft'`,

		// logistic_requests: гостевой черновик (ensureGuestDraftLogisticRequest)
		// WHERE session_id = ? AND is_draft = true AND deleted_at IS NULL
		`CREATE INDEX IF NOT EXISTS idx_logistic_requests_draft_session_id
			ON logistic_requests (session_id)
			WHERE deleted_at IS NULL AND is_draft = true`,

		// transport_services: фильтры (GetTransportServicesWithFilters)
		// WHERE deleted_at IS NULL (+ optional price/created_at)
		`CREATE INDEX IF NOT EXISTS idx_transport_services_active_created_at_desc
			ON transport_services (created_at DESC)
			WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_transport_services_active_price
			ON transport_services (price)
			WHERE deleted_at IS NULL`,
	}
	for _, ddl := range indexDDLs {
		if _, err := sqlDB.Exec(ddl); err != nil {
			panic(err)
		}
	}
	// Обновляем статистику, чтобы планировщик корректно выбирал индексы.
	sqlDB.Exec(`ANALYZE logistic_requests`)
	sqlDB.Exec(`ANALYZE transport_services`)
	sqlDB.Exec(`ANALYZE logistic_request_services`)

	// Создаем системных пользователей
	users := []ds.User{
		{
			ID:       1,
			Login:    "creator",
			Email:    "creator@example.com",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Name:     "Создатель",
			Role:     ds.RoleBuyer,
		},
		{
			ID:       2,
			Login:    "moderator",
			Email:    "moderator@example.com",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Name:     "Модератор",
			Role:     ds.RoleManager,
		},
	}

	for _, user := range users {
		var existingUser ds.User
		err := db.Where("id = ?", user.ID).First(&existingUser).Error
		if err != nil {
			db.Create(&user)
		}
	}

	// Создаем начальные данные
	services := []ds.TransportService{
		{
			ID:           1,
			Name:         "Фура",
			Description:  "Полуприцеп для перевозки крупногабаритных грузов. Идеально подходит для перевозки мебели, строительных материалов и других тяжелых грузов.",
			Price:        150.0,
			ImageURL:     "http://localhost:9003/lab1/fura.jpg",
			DeliveryDays: 2,
			MaxWeight:    20000.0,
			MaxVolume:    80.0,
		},
		{
			ID:           2,
			Name:         "Малотоннажный грузовик",
			Description:  "Легкий грузовик для перевозки небольших грузов по городу и между городами. Быстрая доставка с возможностью проезда в центр города.",
			Price:        80.0,
			ImageURL:     "http://localhost:9003/lab1/malotonnazhnyi.jpg",
			DeliveryDays: 1,
			MaxWeight:    3000.0,
			MaxVolume:    15.0,
		},
		{
			ID:           3,
			Name:         "Авиаперевозка",
			Description:  "Быстрая доставка грузов авиатранспортом. Подходит для срочных и ценных грузов. Максимальная скорость доставки.",
			Price:        500.0,
			ImageURL:     "http://localhost:9003/lab1/avia.jpg",
			DeliveryDays: 1,
			MaxWeight:    1000.0,
			MaxVolume:    5.0,
		},
		{
			ID:           4,
			Name:         "Поезд",
			Description:  "Железнодорожные перевозки для крупных партий грузов. Экономичный вариант для больших объемов.",
			Price:        120.0,
			ImageURL:     "http://localhost:9003/lab1/poezd.jpg",
			DeliveryDays: 3,
			MaxWeight:    50000.0,
			MaxVolume:    120.0,
		},
		{
			ID:           5,
			Name:         "Корабль",
			Description:  "Морские перевозки для международной доставки. Подходит для крупных партий и контейнерных перевозок.",
			Price:        200.0,
			ImageURL:     "http://localhost:9003/lab1/korabl.jpg",
			DeliveryDays: 7,
			MaxWeight:    100000.0,
			MaxVolume:    500.0,
		},
		{
			ID:           6,
			Name:         "Мультимодальные",
			Description:  "Комбинированные перевозки с использованием нескольких видов транспорта. Оптимальное решение для сложных маршрутов.",
			Price:        300.0,
			ImageURL:     "http://localhost:9003/lab1/multimodal.jpg",
			DeliveryDays: 5,
			MaxWeight:    30000.0,
			MaxVolume:    100.0,
		},
	}

	// Создаем услуги в БД
	for _, service := range services {
		var existingService ds.TransportService
		err := db.Where("id = ?", service.ID).First(&existingService).Error
		if err != nil {
			// Услуга не существует, создаем
			db.Create(&service)
		}
	}

	// Создаем пример заявки
	var existingLogisticRequest ds.LogisticRequest
	err = db.Where("id = ?", 1).First(&existingLogisticRequest).Error
	if err != nil {
		order := ds.LogisticRequest{
			ID:        1,
			CreatorID: 1, // системный создатель
			FromCity:  "Москва",
			ToCity:    "Санкт-Петербург",
			Weight:    500.0,
			Length:    2.0,
			Width:     1.5,
			Height:    1.0,
			TotalCost: 230.0,
			TotalDays: 2,
			Status:    ds.StatusDraft,
			IsDraft:   true,
		}
		db.Create(&order)

		// Создаем услуги в заявке
		orderServices := []ds.LogisticRequestService{
			{LogisticRequestID: 1, TransportServiceID: 1, Quantity: 1, Comment: "Основная доставка", SortOrder: 1},
			{LogisticRequestID: 1, TransportServiceID: 2, Quantity: 1, Comment: "Дополнительная услуга", SortOrder: 2},
		}
		for _, orderService := range orderServices {
			db.Create(&orderService)
		}
	}

	println("Migration completed successfully!")
}
