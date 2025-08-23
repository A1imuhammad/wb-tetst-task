package validate_test

import (
	"testing"
	"time"

	"demoserv/internal/models"
	"demoserv/internal/validate"
)

func baseOrder() models.Order {
	return models.Order{
		OrderUID:    "o-1",
		TrackNumber: "T-1",
		DateCreated: time.Now().Add(-time.Hour),
		Payment: models.Payment{
			Transaction: "tx-1",
			PaymentDT:   time.Now().Unix(),
			Amount:      100,
			Currency:    "USD",
			Provider:    "x",
		},
		Delivery: models.Delivery{
			Name:    "Ivan",
			Phone:   "+70000000000",
			Address: "Moscow, 1",
		},
		Items: []models.Item{
			{ChrtID: 1, TrackNumber: "T-1"},
		},
		CustomerID:      "cust",
		DeliveryService: "meest",
	}
}

func TestValidateOrder_Valid(t *testing.T) {
	o := baseOrder()
	if err := validate.ValidateOrder(o); err != nil {
		t.Fatalf("expected valid order, got: %v", err)
	}
}

func TestValidateOrder_MissingFields(t *testing.T) {
	o := baseOrder()
	o.OrderUID = ""
	if err := validate.ValidateOrder(o); err == nil {
		t.Fatalf("expected error for missing order_uid, got nil")
	}
}

func TestValidateOrder_DuplicateChrtID(t *testing.T) {
	o := baseOrder()
	o.Items = append(o.Items, models.Item{ChrtID: 1, TrackNumber: "T-1"})
	if err := validate.ValidateOrder(o); err == nil {
		t.Fatalf("expected duplicate chrt_id error, got nil")
	}
}

func TestValidateOrder_ItemTrackMismatch(t *testing.T) {
	o := baseOrder()
	o.Items[0].TrackNumber = "DIFF"
	if err := validate.ValidateOrder(o); err == nil {
		t.Fatalf("expected item.track_number mismatch error, got nil")
	}
}
