package ds

// DraftLogisticRequest — представление черновика логистической заявки (раньше "корзина")
type DraftLogisticRequest struct {
    ID        int          `json:"id" gorm:"primaryKey"`    // это id заявки в таблице logistic_requests
    SessionID string       `json:"session_id"`
    IsDraft   bool         `json:"is_draft" gorm:"not null;default:true"`
    Services  []DraftLogisticRequestService `json:"services" gorm:"foreignKey:LogisticRequestID"`
}

func (DraftLogisticRequest) TableName() string { return "logistic_requests" }

// DraftLogisticRequestService — строка черновика заявки хранится в logistic_request_services
type DraftLogisticRequestService struct {
    ID                 int `json:"id" gorm:"primaryKey"`
    LogisticRequestID  int `json:"logistic_request_id" gorm:"not null"`
    TransportServiceID int `json:"transport_service_id" gorm:"not null"`
    Quantity           int `json:"quantity" gorm:"not null"`
}

func (DraftLogisticRequestService) TableName() string { return "logistic_request_services" }
