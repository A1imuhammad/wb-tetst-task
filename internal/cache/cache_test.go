package cache_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"demoserv/internal/cache"
	"demoserv/internal/models"
	"demoserv/internal/postgress"
	"demoserv/internal/testutils"

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
		DateCreated: time.Now().Add(-time.Minute),
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

func TestCache_AddGet_Eviction(t *testing.T) {
	c := cache.NewCache(2)
	o1 := models.Order{OrderUID: "o1"}
	o2 := models.Order{OrderUID: "o2"}
	o3 := models.Order{OrderUID: "o3"}

	c.Add(o1)
	c.Add(o2)

	if _, ok := c.Get("o1"); !ok {
		t.Fatalf("expected o1 in cache")
	}

	c.Add(o3) // evict o1
	if _, ok := c.Get("o1"); ok {
		t.Fatalf("expected o1 evicted")
	}
	if _, ok := c.Get("o2"); !ok {
		t.Fatalf("expected o2 to remain")
	}
	if _, ok := c.Get("o3"); !ok {
		t.Fatalf("expected o3 present")
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := cache.NewCache(1000)
	var wg sync.WaitGroup
	total := 500
	wg.Add(total * 2)

	for i := 0; i < total; i++ {
		i := i
		go func() {
			defer wg.Done()
			o := models.Order{OrderUID: fmt.Sprintf("c-%d", i)}
			c.Add(o)
		}()
	}

	for i := 0; i < total; i++ {
		i := i
		go func() {
			defer wg.Done()
			_, _ = c.Get(fmt.Sprintf("c-%d", i))
		}()
	}

	wg.Wait()

	// just ensure cache is non-nil and no panic happened
	if c == nil {
		t.Fatalf("cache is nil")
	}
}

func TestInitCacheFromDB_Embedded(t *testing.T) {
	_, pool := testutils.StartEmbeddedPG(t)

	createTables(t, pool)

	ctx := context.Background()

	// insert 2 orders
	o1 := sampleOrder("init-1")
	o2 := sampleOrder("init-2")
	o2.DateCreated = time.Now() // o2 is newer

	if err := postgress.InsertOrder(ctx, pool, o1); err != nil {
		t.Fatalf("InsertOrder o1: %v", err)
	}
	if err := postgress.InsertOrder(ctx, pool, o2); err != nil {
		t.Fatalf("InsertOrder o2: %v", err)
	}

	c := cache.NewCache(10)
	if err := cache.InitCacheFromDB(ctx, pool, c); err != nil {
		t.Fatalf("InitCacheFromDB failed: %v", err)
	}

	// newest was o2 — it should be present
	if _, ok := c.Get("init-2"); !ok {
		t.Fatalf("expected init-2 in cache after init")
	}
}
