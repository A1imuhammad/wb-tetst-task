package validate

import (
	"demoserv/internal/models"
	"errors"
	"fmt"
	"time"
)


// Проверяет валидность заказа 
func ValidateOrder(order models.Order) error {
	// Проверка идентификаторов заказа и транзакции
	if order.OrderUID == "" {
		return errors.New("order_uid is required")
	}
	if order.TrackNumber == "" {
		return errors.New("track_number is required")
	}
	if order.Payment.Transaction == "" {
		return errors.New("payment.transaction is required")
	}

	// Проверка временных меток
	if order.DateCreated.IsZero() {
		return errors.New("date_created is required and must be valid")
	}
	if order.DateCreated.After(time.Now()) {
		return errors.New("date_created cannot be in the future")
	}
	if order.Payment.PaymentDT <= 0 {
		return errors.New("payment.payment_dt must be positive")
	}

	// Проверка связанных сущностей: Delivery
	if order.Delivery.Name == "" {
		return errors.New("delivery.name is required")
	}
	if order.Delivery.Phone == "" {
		return errors.New("delivery.phone is required")
	}
	if order.Delivery.Address == "" {
		return errors.New("delivery.address is required")
	}

	// Проверка связанных сущностей: Items
	if len(order.Items) == 0 {
		return errors.New("items must not be empty")
	}
	seenChrtID := make(map[int64]bool)
	for i, item := range order.Items {
		if err := validateItem(item, order.TrackNumber); err != nil {
			return fmt.Errorf("invalid item at index %d: %v", i, err)
		}
		// Проверка уникальности ChrtID в рамках заказа
		if seenChrtID[item.ChrtID] {
			return fmt.Errorf("duplicate chrt_id %d at index %d", item.ChrtID, i)
		}
		seenChrtID[item.ChrtID] = true
	}

	// Проверка финансовых данных
	if order.Payment.Amount <= 0 {
		return errors.New("payment.amount must be positive")
	}
	if order.Payment.Currency == "" {
		return errors.New("payment.currency is required")
	}
	if order.Payment.Provider == "" {
		return errors.New("payment.provider is required")
	}


	// Проверка системных полей
	if order.CustomerID == "" {
		return errors.New("customer_id is required")
	}
	if order.DeliveryService == "" {
		return errors.New("delivery_service is required")
	}

	return nil
}

// Проверяет валидность элемента заказа
func validateItem(item models.Item, orderTrackNumber string) error {
	// Проверка track_number
	if item.TrackNumber == "" {
		return errors.New("item.track_number is required")
	}
	// Проверка совпадения item.track_number с order.track_number
	if item.TrackNumber != orderTrackNumber {
		return errors.New("item.track_number must match order.track_number")
	}
	// Проверка chrt_id
	if item.ChrtID == 0 {
		return errors.New("item.chrt_id is required and must be non-zero")
	}
	
	return nil
}