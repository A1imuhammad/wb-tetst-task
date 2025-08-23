package postgress_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"demoserv/internal/models"
	"demoserv/internal/postgress"
	"demoserv/internal/testutils"

	embedded "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// getFreePort same helper as in cache_test
func getFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("getFreePort listen: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func startEmbeddedPG(t *testing.T, port int) (*embedded.EmbeddedPostgres, *pgxpool.Pool) {
	t.Helper()
	cfg := embedded.DefaultConfig().
		Username("postgres").
		Password("secret").
		Database("testdb").
		Version(embedded.V13).
		Port(uint32(port))

	epg := embedded.NewDatabase(cfg)
	if err := epg.Start(); err != nil {
		t.Fatalf("embedded start: %v", err)
	}

	uri := fmt.Sprintf("postgres://postgres:secret@127.0.0.1:%d/testdb?sslmode=disable", port)

	var pool *pgxpool.Pool
	deadline := time.Now().Add(8 * time.Second)
	for {
		var err error
		pool, err = pgxpool.New(context.Background(), uri)
		if err == nil {
			if pingErr := pool.Ping(context.Background()); pingErr == nil {
				break
			}
			pool.Close()
		}
		if time.Now().After(deadline) {
			_ = epg.Stop()
			t.Fatalf("cannot connect to embedded postgres: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return epg, pool
}

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

func makeOrder(uid string, created time.Time) *models.Order {
	return &models.Order{
		OrderUID:    uid,
		TrackNumber: "T-" + uid,
		DateCreated: created,
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

func TestInsertAndGetOrder_Embedded(t *testing.T) {
	_, pool := testutils.StartEmbeddedPG(t)

	createTables(t, pool)

	ctx := context.Background()
	o := makeOrder("itest-1", time.Now().Add(-2*time.Hour))

	if err := postgress.InsertOrder(ctx, pool, o); err != nil {
		t.Fatalf("InsertOrder failed: %v", err)
	}

	got, err := postgress.GetOrder(ctx, o.OrderUID, pool)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if got.OrderUID != o.OrderUID {
		t.Fatalf("unexpected uid: got %s want %s", got.OrderUID, o.OrderUID)
	}
}

func TestGetLastOrders_Embedded(t *testing.T) {
	port := getFreePort(t)
	epg, pool := startEmbeddedPG(t, port)
	defer func() {
		pool.Close()
		_ = epg.Stop()
	}()

	createTables(t, pool)

	ctx := context.Background()
	o1 := makeOrder("o-old", time.Now().Add(-2*time.Hour))
	o2 := makeOrder("o-new", time.Now().Add(-1*time.Hour))

	if err := postgress.InsertOrder(ctx, pool, o1); err != nil {
		t.Fatalf("insert o1: %v", err)
	}
	if err := postgress.InsertOrder(ctx, pool, o2); err != nil {
		t.Fatalf("insert o2: %v", err)
	}

	got, err := postgress.GetLastOrders(ctx, pool, 1)
	if err != nil {
		t.Fatalf("GetLastOrders failed: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 order, got %d", len(got))
	}
	if got[0].OrderUID != "o-new" {
		t.Fatalf("expected latest o-new, got %s", got[0].OrderUID)
	}
}
