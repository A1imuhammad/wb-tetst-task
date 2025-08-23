package cache

import (
	// "container/list"
	"context"
	"demoserv/internal/models"
	"demoserv/internal/postgress"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Cache struct {
	mu    sync.Mutex
	data  map[string]models.Order
	queue []string
	limit int
}

// NewCache создает новый кэш
func NewCache(limit int) *Cache {
	return &Cache{
		data:  make(map[string]models.Order),
		queue: make([]string, 0),
		limit: limit,
	}
}

// Add добавляет заказ в кэш
func (c *Cache) Add(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.queue) >= c.limit {
		oldest := c.queue[0]
		c.queue = c.queue[1:]
		delete(c.data, oldest)
	}

	c.queue = append(c.queue, order.OrderUID)
	c.data[order.OrderUID] = order
}

// Get получает заказ из кэша
func (c *Cache) Get(orderUID string) (models.Order, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	order, ok := c.data[orderUID]
	return order, ok
}

// InitCacheFromDB инициализирует кэш из базы
func InitCacheFromDB(ctx context.Context, pool *pgxpool.Pool, cache *Cache) error {
	orders, err := postgress.GetLastOrders(ctx, pool, cache.limit)
	if err != nil {
		return fmt.Errorf("unable to get last orders: %v", err)
	}

	for i := len(orders) - 1; i >= 0; i-- {
		cache.Add(orders[i])
	}
	return nil
}
