package ds

import "time"

// LogisticRequest - модель логистической заявки
type LogisticRequest struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	SessionID string `json:"session_id"`
	IsDraft   bool   `json:"is_draft" gorm:"not null;default:true"`
	FromCity  string `json:"from_city"`
	ToCity    string `json:"to_city"`
	// Параметры груза
	Weight    float64                  `json:"weight" gorm:"not null;default:0"`
	Length    float64                  `json:"length" gorm:"not null;default:0"`
	Width     float64                  `json:"width" gorm:"not null;default:0"`
	Height    float64                  `json:"height" gorm:"not null;default:0"`
	Services  []LogisticRequestService `json:"services" gorm:"foreignKey:LogisticRequestID"`
	TotalCost float64                  `json:"total_cost"`
	TotalDays int                      `json:"total_days"`
	Status    string                   `json:"status" gorm:"type:varchar(32);not null;default:'draft'"`

	// Количество записей м-м (услуги), где рассчитанное поле результата НЕ пустое.
	// Требование лаб8: отображается в GET списка заявок.
	AsyncResultsFilledCount int `json:"async_results_filled_count" gorm:"-"`

	// Системные поля
	CreatorID   int        `json:"creator_id" gorm:"not null"`
	ModeratorID *int       `json:"moderator_id"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	FormedAt    *time.Time `json:"formed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"-" gorm:"index"`

	// Связи
	Creator   User  `json:"creator" gorm:"foreignKey:CreatorID"`
	Moderator *User `json:"moderator" gorm:"foreignKey:ModeratorID"`
}

func (LogisticRequest) TableName() string {
	return "logistic_requests"
}

// LogisticRequest statuses
const (
	StatusDraft     = "draft"     // черновик
	StatusFormed    = "formed"    // сформирован
	StatusCompleted = "completed" // завершён
	StatusRejected  = "rejected"  // отклонён
	StatusDeleted   = "deleted"   // удалён
)

// LogisticRequestService - услуга в логистической заявке (м-м)
type LogisticRequestService struct {
	ID                 int    `json:"id" gorm:"primaryKey"`
	// Требование: составной уникальный ключ по (logistic_request_id, transport_service_id).
	// Также это нужно для ON CONFLICT (logistic_request_id, transport_service_id) в репозитории.
	LogisticRequestID  int    `json:"logistic_request_id" gorm:"not null;uniqueIndex:ux_logistic_request_service,priority:1"`
	TransportServiceID int    `json:"transport_service_id" gorm:"not null;uniqueIndex:ux_logistic_request_service,priority:2"`
	Quantity           int    `json:"quantity" gorm:"not null;default:1"`
	Comment            string `json:"comment" gorm:"type:text"`
	SortOrder          int    `json:"sort_order" gorm:"column:order;not null;default:0"`
	// Результат отложенного действия (заполняется async-сервисом).
	AsyncResult string `json:"async_result" gorm:"type:varchar(128)"`

	// Связи
	TransportService TransportService `json:"service" gorm:"foreignKey:TransportServiceID"`
}

func (LogisticRequestService) TableName() string {
	return "logistic_request_services"
}
