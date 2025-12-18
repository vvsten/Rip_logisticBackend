package repository

import (
    "database/sql"
    "fmt"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "rip-go-app/internal/app/ds"
    "rip-go-app/internal/app/calculator"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // подключаемся к БД
	if err != nil {
		return nil, err
	}

	// Возвращаем объект Repository с подключенной базой данных
	return &Repository{
		db: db,
	}, nil
}

// GetTransportServices - получение всех транспортных услуг с возможностью фильтрации (исключая удалённые)
func (r *Repository) GetTransportServices(search string) ([]ds.TransportService, error) {
	var services []ds.TransportService
	
	query := r.db.Where("deleted_at IS NULL")
	
	if search != "" {
		searchLower := strings.ToLower(search)
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", 
			"%"+searchLower+"%", "%"+searchLower+"%")
	}
	
	err := query.Find(&services).Error
	if err != nil {
		return nil, err
	}

	return services, nil
}

// GetTransportServicesWithFilters - получение транспортных услуг с расширенными фильтрами для API
func (r *Repository) GetTransportServicesWithFilters(search string, minPrice, maxPrice *float64, dateFrom, dateTo *time.Time) ([]ds.TransportService, error) {
	var services []ds.TransportService
	
	query := r.db.Where("deleted_at IS NULL")
	
	// Поиск по названию и описанию
	if search != "" {
		searchLower := strings.ToLower(search)
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", 
			"%"+searchLower+"%", "%"+searchLower+"%")
	}
	
	// Фильтр по минимальной цене
	if minPrice != nil {
		query = query.Where("price >= ?", *minPrice)
	}
	
	// Фильтр по максимальной цене
	if maxPrice != nil {
		query = query.Where("price <= ?", *maxPrice)
	}
	
	// Фильтр по дате создания (от)
	if dateFrom != nil {
		query = query.Where("created_at >= ?", *dateFrom)
	}
	
	// Фильтр по дате создания (до)
	if dateTo != nil {
		query = query.Where("created_at <= ?", *dateTo)
	}
	
	err := query.Find(&services).Error
	if err != nil {
		return nil, err
	}

	return services, nil
}

// GetTransportService - получение транспортной услуги по ID
func (r *Repository) GetTransportService(id int) (ds.TransportService, error) {
	var service ds.TransportService
	err := r.db.Where("id = ?", id).First(&service).Error
	if err != nil {
		return ds.TransportService{}, fmt.Errorf("услуга не найдена")
	}
	return service, nil
}

// GetTransportServiceByDeliveryType - получение транспортной услуги по типу доставки
func (r *Repository) GetTransportServiceByDeliveryType(deliveryType string) (ds.TransportService, error) {
	typeMap := map[string]int{
		"fura":           1,
		"malotonnazhnyi": 2,
		"avia":           3,
		"poezd":          4,
		"korabl":         5,
		"multimodal":     6,
	}
	
	if id, exists := typeMap[deliveryType]; exists {
		return r.GetTransportService(id)
	}
	
	return ds.TransportService{}, fmt.Errorf("тип доставки не найден")
}

// CRUD для TransportService
func (r *Repository) CreateTransportService(s *ds.TransportService) error {
    return r.db.Create(s).Error
}

func (r *Repository) UpdateTransportService(s *ds.TransportService) error {
    return r.db.Save(s).Error
}

func (r *Repository) DeleteTransportService(id int) error {
    return r.db.Delete(&ds.TransportService{}, id).Error
}

// CreateCargoLogisticRequest создаёт заказ на основе перечня транспортов и параметров груза
type CargoLogisticRequestItem struct {
    TransportServiceID int
    FromCity  string
    ToCity    string
    Length    float64
    Width     float64
    Height    float64
    Weight    float64
}

func (r *Repository) CreateCargoLogisticRequest(items []CargoLogisticRequestItem, creatorID int) (int, error) {
    if len(items) == 0 {
        return 0, fmt.Errorf("no items provided")
    }

    return r.createCargoLogisticRequestTx(items, creatorID)
}

func (r *Repository) createCargoLogisticRequestTx(items []CargoLogisticRequestItem, creatorID int) (int, error) {
    calc := calculator.NewDeliveryCalculator()

    returnID := 0
    err := r.db.Transaction(func(tx *gorm.DB) error {
        // Используем параметры первого как общие
        first := items[0]

        order := ds.LogisticRequest{
            SessionID: "guest",
            IsDraft:   true,
            FromCity:  first.FromCity,
            ToCity:    first.ToCity,
            Weight:    0,
            Length:    0,
            Width:     0,
            Height:    0,
            TotalCost: 0,
            TotalDays: 0,
            Status:    ds.StatusDraft,
            CreatorID: creatorID, // используем переданный creatorID
        }
        if err := tx.Create(&order).Error; err != nil {
            return err
        }

        // агрегаты
        maxDays := 0
        totalCost := 0.0
        totalWeight := 0.0
        totalLength := 0.0
        totalWidth := 0.0
        totalHeight := 0.0

        for _, it := range items {
            svc, err := r.GetTransportService(it.TransportServiceID)
            if err != nil {
                return fmt.Errorf("service %d not found", it.TransportServiceID)
            }
            res := calc.CalculateDelivery(svc, it.FromCity, it.ToCity, it.Length, it.Width, it.Height, it.Weight)
            if !res.IsValid {
                return fmt.Errorf("%s", res.ErrorMessage)
            }

            // создаём строку заказа
            os := ds.LogisticRequestService{LogisticRequestID: order.ID, TransportServiceID: it.TransportServiceID, Quantity: 1}
            if err := tx.Create(&os).Error; err != nil {
                return err
            }

            if res.DeliveryDays > maxDays { maxDays = res.DeliveryDays }
            totalCost += res.TotalCost
            totalWeight += it.Weight
            totalLength += it.Length
            totalWidth += it.Width
            totalHeight += it.Height
        }

        // итоговые поля заказа
        order.TotalDays = maxDays
        order.TotalCost = totalCost
        order.Weight = totalWeight
        order.Length = totalLength
        order.Width = totalWidth
        order.Height = totalHeight

        if err := tx.Save(&order).Error; err != nil {
            return err
        }

        returnID = order.ID
        return nil
    })

    if err != nil {
        return 0, err
    }
    return returnID, nil
}

// ==================== ПОЛЬЗОВАТЕЛИ ====================

// CreateUser - создание пользователя
func (r *Repository) CreateUser(user *ds.User) error {
    // Генерируем UUID если он не задан
    if user.UUID == "" {
        user.UUID = uuid.New().String()
    }
    return r.db.Create(user).Error
}

// GetUserByLogin - получение пользователя по логину
func (r *Repository) GetUserByLogin(login string) (ds.User, error) {
    var user ds.User
    err := r.db.Where("login = ?", login).First(&user).Error
    if err != nil {
        return ds.User{}, fmt.Errorf("пользователь не найден")
    }
    return user, nil
}

// GetUser - получение пользователя по ID
func (r *Repository) GetUser(id int) (ds.User, error) {
    var user ds.User
    err := r.db.Where("id = ?", id).First(&user).Error
    if err != nil {
        return ds.User{}, fmt.Errorf("пользователь не найден")
    }
    return user, nil
}

// GetUserByUUID - получение пользователя по UUID
func (r *Repository) GetUserByUUID(userUUID string) (ds.User, error) {
    var user ds.User
    err := r.db.Where("uuid = ?", userUUID).First(&user).Error
    if err != nil {
        return ds.User{}, fmt.Errorf("пользователь не найден")
    }
    return user, nil
}

// UpdateUser - обновление пользователя
func (r *Repository) UpdateUser(user *ds.User) error {
    return r.db.Save(user).Error
}

// ==================== ЗАЯВКИ ====================

// GetLogisticRequests - получение списка заявок с фильтрацией (исключая удалённые и черновики)
func (r *Repository) GetLogisticRequests(status string, dateFrom, dateTo *time.Time) ([]ds.LogisticRequest, error) {
    var orders []ds.LogisticRequest
    
    query := r.db.Preload("Creator").Preload("Moderator").
        Where("deleted_at IS NULL AND status != ?", ds.StatusDraft)
    
    if status != "" {
        query = query.Where("status = ?", status)
    }
    
    if dateFrom != nil {
        query = query.Where("formed_at >= ?", *dateFrom)
    }
    
    if dateTo != nil {
        query = query.Where("formed_at <= ?", *dateTo)
    }
    
    err := query.Order("created_at DESC").Find(&orders).Error
    return orders, err
}

// GetLogisticRequest - получение заявки по ID с услугами
func (r *Repository) GetLogisticRequest(id int) (ds.LogisticRequest, error) {
    var order ds.LogisticRequest
    err := r.db.Preload("Services.TransportService").Preload("Creator").Preload("Moderator").
        Where("id = ? AND deleted_at IS NULL", id).First(&order).Error
    if err != nil {
        return ds.LogisticRequest{}, fmt.Errorf("заявка не найдена")
    }
    return order, nil
}

// GetDraftLogisticRequest - получение черновика заявки пользователя
func (r *Repository) GetDraftLogisticRequest(creatorID int) (ds.LogisticRequest, error) {
    var order ds.LogisticRequest
    err := r.db.Preload("Services.TransportService").
        Where("creator_id = ? AND status = ? AND deleted_at IS NULL", creatorID, ds.StatusDraft).
        First(&order).Error
    if err != nil {
        return ds.LogisticRequest{}, fmt.Errorf("черновик не найден")
    }
    return order, nil
}

// CreateDraftLogisticRequest - создание черновика заявки
func (r *Repository) CreateDraftLogisticRequest(creatorID int) (ds.LogisticRequest, error) {
    order := ds.LogisticRequest{
        CreatorID: creatorID,
        Status:    ds.StatusDraft,
        IsDraft:   true,
    }
    err := r.db.Create(&order).Error
    return order, err
}

// UpdateLogisticRequest - обновление заявки
func (r *Repository) UpdateLogisticRequest(order *ds.LogisticRequest) error {
    return r.db.Save(order).Error
}

// FormLogisticRequest - формирование заявки создателем (проверка обязательных полей)
func (r *Repository) FormLogisticRequest(orderID int, fromCity, toCity string, weight, length, width, height float64) error {
    var order ds.LogisticRequest
    err := r.db.Preload("Services").Where("id = ?", orderID).First(&order).Error
    if err != nil {
        return fmt.Errorf("заявка не найдена")
    }
    
    // Проверяем, что заявка в статусе draft
    if order.Status != ds.StatusDraft {
        return fmt.Errorf("можно формировать только черновики")
    }
    
    // Проверяем обязательные поля
    if fromCity == "" || toCity == "" || weight <= 0 || length <= 0 || width <= 0 || height <= 0 {
        return fmt.Errorf("не заполнены обязательные поля: города и параметры груза")
    }
    
    if len(order.Services) == 0 {
        return fmt.Errorf("в заявке нет услуг")
    }
    
    // Обновляем заявку
    now := time.Now()
    order.FromCity = fromCity
    order.ToCity = toCity
    order.Weight = weight
    order.Length = length
    order.Width = width
    order.Height = height
    order.Status = ds.StatusFormed
    order.FormedAt = &now
    order.IsDraft = false
    
    return r.db.Save(&order).Error
}

// CompleteLogisticRequest - завершение/отклонение заявки модератором
func (r *Repository) CompleteLogisticRequest(orderID int, status string, moderatorID int) error {
    if status != ds.StatusCompleted && status != ds.StatusRejected {
        return fmt.Errorf("неверный статус для завершения")
    }
    
    var order ds.LogisticRequest
    err := r.db.Preload("Services.TransportService").Where("id = ?", orderID).First(&order).Error
    if err != nil {
        return fmt.Errorf("заявка не найдена")
    }
    
    if order.Status != ds.StatusFormed {
        return fmt.Errorf("можно завершать только сформированные заявки")
    }
    
    // Рассчитываем стоимость и сроки при завершении
    if status == ds.StatusCompleted {
        calc := calculator.NewDeliveryCalculator()
        totalCost := 0.0
        maxDays := 0
        
        for _, orderService := range order.Services {
            res := calc.CalculateDelivery(orderService.TransportService, order.FromCity, order.ToCity, 
                order.Length, order.Width, order.Height, order.Weight)
            if res.IsValid {
                totalCost += res.TotalCost
                if res.DeliveryDays > maxDays {
                    maxDays = res.DeliveryDays
                }
            }
        }
        
        order.TotalCost = totalCost
        order.TotalDays = maxDays
    }
    
    now := time.Now()
    order.Status = status
    order.ModeratorID = &moderatorID
    order.CompletedAt = &now
    
    return r.db.Save(&order).Error
}

// DeleteLogisticRequest - удаление заявки (мягкое удаление)

func (r *Repository) DeleteLogisticRequest(orderID int) error {
    // Каскадное удаление автоматически удалит связанные записи в logistic_request_services
    return r.db.Where("id = ?", orderID).Delete(&ds.LogisticRequest{}).Error
}

// GetCartIcon - получение иконки корзины (количество услуг в черновике)
func (r *Repository) GetCartIcon(creatorID int) (int, int, error) {
    var order ds.LogisticRequest
    err := r.db.Preload("Services").Where("creator_id = ? AND status = ? AND deleted_at IS NULL", 
        creatorID, ds.StatusDraft).First(&order).Error
    if err != nil {
        // Создаём черновик если нет
        order, err = r.CreateDraftLogisticRequest(creatorID)
        if err != nil {
            return 0, 0, err
        }
    }
    
    count := len(order.Services)
    return order.ID, count, nil
}

// GetLogisticRequestServiceQuantitySum - сумма quantity по услугам заявки
// Используется для счетчика в UI (если одну услугу добавили 3 раза — хотим видеть 3).
func (r *Repository) GetLogisticRequestServiceQuantitySum(orderID int) (int, error) {
	var sum sql.NullInt64
	err := r.db.
		Model(&ds.LogisticRequestService{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("logistic_request_id = ?", orderID).
		Scan(&sum).Error
	if err != nil {
		return 0, err
	}
	if sum.Valid {
		return int(sum.Int64), nil
	}
	return 0, nil
}

// ClearUserDraftLogisticRequest - очистка черновика заявки пользователя (удаляем строки услуг)
func (r *Repository) ClearUserDraftLogisticRequest(creatorID int) error {
	draft, err := r.GetDraftLogisticRequest(creatorID)
	if err != nil {
		// Если черновика нет — считаем, что уже очищено
		return nil
	}
	return r.db.Where("logistic_request_id = ?", draft.ID).Delete(&ds.LogisticRequestService{}).Error
}

// ==================== М-М ЗАЯВКА-УСЛУГА ====================

// AddServiceToLogisticRequest - добавление услуги в заявку-черновик
func (r *Repository) AddServiceToLogisticRequest(orderID, serviceID int) error {
    // Проверяем что заявка - черновик
    var order ds.LogisticRequest
    err := r.db.Where("id = ? AND status = ?", orderID, ds.StatusDraft).First(&order).Error
    if err != nil {
        return fmt.Errorf("заявка не найдена или не является черновиком")
    }
    
    // Проверяем услугу
    _, err = r.GetTransportService(serviceID)
    if err != nil {
        return fmt.Errorf("услуга не найдена")
    }
    
    // Проверяем не добавлена ли уже
    var existing ds.LogisticRequestService
    err = r.db.Where("logistic_request_id = ? AND transport_service_id = ?", orderID, serviceID).First(&existing).Error
    if err == nil {
        // Увеличиваем количество
        existing.Quantity++
        return r.db.Save(&existing).Error
    }
    
    // Добавляем новую
    orderService := ds.LogisticRequestService{
        LogisticRequestID:   orderID,
        TransportServiceID: serviceID,
        Quantity:  1,
    }
    return r.db.Create(&orderService).Error
}

// RemoveServiceFromLogisticRequest - удаление услуги из заявки
func (r *Repository) RemoveServiceFromLogisticRequest(orderID, serviceID int) error {
    var orderService ds.LogisticRequestService
    err := r.db.Where("logistic_request_id = ? AND transport_service_id = ?", orderID, serviceID).First(&orderService).Error
    if err != nil {
        return fmt.Errorf("услуга не найдена в заявке")
    }
    
    return r.db.Delete(&orderService).Error
}

// UpdateLogisticRequestService - обновление количества/порядка в м-м
func (r *Repository) UpdateLogisticRequestService(orderID, serviceID int, quantity, orderNum int, comment string) error {
    var orderService ds.LogisticRequestService
    err := r.db.Where("logistic_request_id = ? AND transport_service_id = ?", orderID, serviceID).First(&orderService).Error
    if err != nil {
        return fmt.Errorf("услуга не найдена в заявке")
    }
    
    orderService.Quantity = quantity
    orderService.SortOrder = orderNum
    orderService.Comment = comment
    
    return r.db.Save(&orderService).Error
}


// ensureGuestDraftLogisticRequest - гарантирует наличие черновика логистической заявки для sessionID (guest/web)
func (r *Repository) ensureGuestDraftLogisticRequest(sessionID string) (int, error) {
    var order ds.LogisticRequest
    if err := r.db.Where("session_id = ? AND is_draft = ? AND deleted_at IS NULL", sessionID, true).First(&order).Error; err != nil {
        // создаём с системным создателем
        order = ds.LogisticRequest{
            SessionID: sessionID, 
            IsDraft: true,
            CreatorID: ds.GetCreatorID(),
            Status: ds.StatusDraft,
        }
        if err := r.db.Create(&order).Error; err != nil {
            return 0, err
        }
    }
    return order.ID, nil
}

// AddTransportServiceToGuestDraftLogisticRequest - добавляет транспортную услугу в черновик заявки (guest)
func (r *Repository) AddTransportServiceToGuestDraftLogisticRequest(serviceID int) error {
    // проверяем услугу
    if _, err := r.GetTransportService(serviceID); err != nil {
        return fmt.Errorf("услуга не найдена")
    }
    // берём черновик логистической заявки
    orderID, err := r.ensureGuestDraftLogisticRequest("guest")
    if err != nil { return err }

    // upsert в logistic_request_services
    // используем нативное подключение для ON CONFLICT
    sqlDB, err := r.db.DB(); if err != nil { return err }
    _, err = sqlDB.Exec(`
        INSERT INTO logistic_request_services(logistic_request_id, transport_service_id, quantity)
        VALUES ($1, $2, 1)
        ON CONFLICT (logistic_request_id, transport_service_id)
        DO UPDATE SET quantity = logistic_request_services.quantity + 1
    `, orderID, serviceID)
    return err
}

// RemoveTransportServiceFromGuestDraftLogisticRequest - уменьшает количество услуги в черновике или удаляет строку
func (r *Repository) RemoveTransportServiceFromGuestDraftLogisticRequest(serviceID int) error {
    orderID, err := r.ensureGuestDraftLogisticRequest("guest")
    if err != nil { return err }

    sqlDB, err := r.db.DB(); if err != nil { return err }
    // уменьшаем qty если >1, иначе удаляем
    var qty int
    err = sqlDB.QueryRow(`SELECT quantity FROM logistic_request_services WHERE logistic_request_id=$1 AND transport_service_id=$2`, orderID, serviceID).Scan(&qty)
    if err == sql.ErrNoRows { return fmt.Errorf("услуга не найдена в черновике заявки") }
    if err != nil { return err }

    if qty > 1 {
        _, err = sqlDB.Exec(`UPDATE logistic_request_services SET quantity = quantity - 1 WHERE logistic_request_id=$1 AND transport_service_id=$2`, orderID, serviceID)
    } else {
        _, err = sqlDB.Exec(`DELETE FROM logistic_request_services WHERE logistic_request_id=$1 AND transport_service_id=$2`, orderID, serviceID)
    }
    return err
}

// GetGuestDraftLogisticRequestView - получение представления черновика заявки (guest)
func (r *Repository) GetGuestDraftLogisticRequestView() (ds.DraftLogisticRequest, error) {
    orderID, err := r.ensureGuestDraftLogisticRequest("guest")
    if err != nil { return ds.DraftLogisticRequest{}, err }
    var items []ds.DraftLogisticRequestService
    if err := r.db.Where("logistic_request_id = ?", orderID).Find(&items).Error; err != nil {
        return ds.DraftLogisticRequest{}, err
    }
    return ds.DraftLogisticRequest{ID: orderID, SessionID: "guest", IsDraft: true, Services: items}, nil
}

// GetGuestDraftLogisticRequestServices - услуги в черновике заявки (guest) с полной информацией
func (r *Repository) GetGuestDraftLogisticRequestServices() ([]ds.TransportService, error) {
    orderID, err := r.ensureGuestDraftLogisticRequest("guest")
    if err != nil { 
        logrus.Errorf("GetGuestDraftLogisticRequestServices: failed to ensure draft request: %v", err)
        return nil, err 
    }
    var items []ds.DraftLogisticRequestService
    if err := r.db.Where("logistic_request_id = ?", orderID).Find(&items).Error; err != nil { 
        logrus.Errorf("GetGuestDraftLogisticRequestServices: failed to find draft items: %v", err)
        return nil, err 
    }
    logrus.Infof("GetGuestDraftLogisticRequestServices: found %d items in draft for requestID %d", len(items), orderID)
    services := make([]ds.TransportService, 0, len(items))
    for _, it := range items {
        s, err := r.GetTransportService(it.TransportServiceID)
        if err != nil {
            logrus.Errorf("GetGuestDraftLogisticRequestServices: failed to get service %d: %v", it.TransportServiceID, err)
        } else {
            services = append(services, s)
        }
    }
    logrus.Infof("GetGuestDraftLogisticRequestServices: returning %d services", len(services))
    return services, nil
}

// GetGuestDraftLogisticRequestServiceCount - общее количество услуг в черновике заявки (guest)
func (r *Repository) GetGuestDraftLogisticRequestServiceCount() int {
    orderID, err := r.ensureGuestDraftLogisticRequest("guest")
    if err != nil { return 0 }
    sqlDB, err := r.db.DB(); if err != nil { return 0 }
    var count sql.NullInt64
    _ = sqlDB.QueryRow(`SELECT COALESCE(SUM(quantity),0) FROM logistic_request_services WHERE logistic_request_id=$1`, orderID).Scan(&count)
    if count.Valid { return int(count.Int64) }
    return 0
}

// ClearGuestDraftLogisticRequest - очистка черновика заявки (guest) (удаление всех строк услуг)
func (r *Repository) ClearGuestDraftLogisticRequest() {
    orderID, err := r.ensureGuestDraftLogisticRequest("guest")
    if err != nil { return }
    r.db.Where("logistic_request_id = ?", orderID).Delete(&ds.DraftLogisticRequestService{})
}

// UpdateLogisticRequestStatusWithCursor - обновление статуса заказа через курсор без ORM
func (r *Repository) UpdateLogisticRequestStatusWithCursor(orderID int, newStatus string) error {
	// Получаем нативное подключение к БД из GORM
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Подготавливаем запрос с курсором
	query := `
		UPDATE logistic_requests 
		SET status = $1 
		WHERE id = $2
		RETURNING id, status, from_city, to_city
	`

	// Выполняем запрос через курсор
	rows, err := sqlDB.Query(query, newStatus, orderID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Обрабатываем результат через курсор
	if rows.Next() {
		var id int
		var status, fromCity, toCity string
		
		err := rows.Scan(&id, &status, &fromCity, &toCity)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		
		// Логируем обновление
		fmt.Printf("LogisticRequest %d status updated to: %s (Route: %s -> %s)\n", id, status, fromCity, toCity)
	} else {
		return fmt.Errorf("order with id %d not found", orderID)
	}

	return nil
}
