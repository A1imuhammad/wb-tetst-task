package getorder_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"demoserv/internal/cache"
	getorder "demoserv/internal/http-server/handlers/getOrder"
	"demoserv/internal/models"
	"demoserv/internal/postgress"
	"demoserv/internal/testutils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func createTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	sql := `
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255) NOT NULL,
    entry VARCHAR(50),
    locale VARCHAR(10),
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(10),
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard VARCHAR(10)
);
CREATE TABLE IF NOT EXISTS delivery (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) UNIQUE REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255),
    phone VARCHAR(50),
    zip VARCHAR(20),
    city VARCHAR(255),
    address VARCHAR(255),
    region VARCHAR(255),
    email VARCHAR(255)
);
CREATE TABLE IF NOT EXISTS payment (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) UNIQUE REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction VARCHAR(255),
    request_id VARCHAR(255),
    currency VARCHAR(10),
    provider VARCHAR(50),
    amount INT,
    payment_dt BIGINT,
    bank VARCHAR(50),
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT,
    track_number VARCHAR(255),
    price INT,
    rid VARCHAR(255),
    name VARCHAR(255),
    sale INT,
    size VARCHAR(50),
    total_price INT,
    nm_id BIGINT,
    brand VARCHAR(255),
    status INT,
    CONSTRAINT unique_order_item UNIQUE (order_uid, chrt_id)
);
`
	if _, err := pool.Exec(ctx, sql); err != nil {
		t.Fatalf("createTables exec: %v", err)
	}
}

func sampleOrder(uid string) *models.Order {
	return &models.Order{
		OrderUID:    uid,
		TrackNumber: "T-" + uid,
		DateCreated: time.Now().Add(-time.Hour),
		Delivery: models.Delivery{
			Name:    "N",
			Phone:   "P",
			Address: "A",
		},
		Payment: models.Payment{
			Transaction: "tx",
			PaymentDT:   time.Now().Unix(),
			Amount:      10,
			Currency:    "USD",
			Provider:    "p",
		},
		Items: []models.Item{
			{ChrtID: 1, TrackNumber: "T-" + uid, Price: 10, RID: "r", Name: "n", TotalPrice: 10, NmID: 1, Brand: "b", Status: 1},
		},
		CustomerID:      "cust",
		DeliveryService: "meest",
	}
}

func newChiWithHandler(h http.HandlerFunc) http.Handler {
	r := chi.NewRouter()
	r.Get("/order/{order_uid}", h)
	return r
}

func TestHandler_CacheHit(t *testing.T) {
	c := cache.NewCache(10)
	o := models.Order{OrderUID: "hit-1", TrackNumber: "T", DateCreated: time.Now()}
	c.Add(o)

	h := getorder.New(context.Background(), c, nil)

	req := httptest.NewRequest("GET", "/order/hit-1", nil)
	rr := httptest.NewRecorder()

	r := newChiWithHandler(h)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var got models.Order
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got.OrderUID != "hit-1" {
		t.Fatalf("unexpected uid: %s", got.OrderUID)
	}
}

func TestHandler_CacheMiss_DBFetch(t *testing.T) {
	_, pool := testutils.StartEmbeddedPG(t)

	createTables(t, pool)

	ctx := context.Background()
	order := sampleOrder("db-1")
	if err := postgress.InsertOrder(ctx, pool, order); err != nil {
		t.Fatalf("InsertOrder failed: %v", err)
	}

	c := cache.NewCache(10)
	h := getorder.New(context.Background(), c, pool)

	req := httptest.NewRequest("GET", "/order/db-1", nil)
	rr := httptest.NewRecorder()

	r := newChiWithHandler(h)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d; body: %s", rr.Code, rr.Body.String())
	}

	var got models.Order
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got.OrderUID != "db-1" {
		t.Fatalf("unexpected uid: %s", got.OrderUID)
	}
}
