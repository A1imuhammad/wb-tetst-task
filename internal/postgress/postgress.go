package postgress

import (
	"context"
	"fmt"

	"demoserv/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertOrder(ctx context.Context, pool *pgxpool.Pool, order *models.Order) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// Вставка в orders
	_, err = tx.Exec(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature, 
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("unable to insert into orders: %v", err)
	}

	// Вставка в delivery
	_, err = tx.Exec(ctx, `
		INSERT INTO delivery (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("unable to insert into delivery: %v", err)
	}

	// Вставка в payment
	_, err = tx.Exec(ctx, `
		INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider, amount, 
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("unable to insert into payment: %v", err)
	}

	// Вставка в items для каждого элемента
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid, name, 
				sale, size, total_price, nm_id, brand, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (order_uid, chrt_id) DO NOTHING`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("unable to insert into items: %v", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("unable to commit transaction: %v", err)
	}

	return nil
}

func GetLastOrders(ctx context.Context, pool *pgxpool.Pool, limit int) ([]models.Order, error) {
	rows, err := pool.Query(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature,
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		ORDER BY date_created DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.OrderUID,
			&o.TrackNumber,
			&o.Entry,
			&o.Locale,
			&o.InternalSignature,
			&o.CustomerID,
			&o.DeliveryService,
			&o.ShardKey,
			&o.SmID,
			&o.DateCreated,
			&o.OofShard,
		); err != nil {
			return nil, err
		}

		// достаем из Delivery
		err = pool.QueryRow(ctx, `
			SELECT name, phone, zip, city, address, region, email
			FROM delivery WHERE order_uid=$1
		`, o.OrderUID).Scan(
			&o.Delivery.Name,
			&o.Delivery.Phone,
			&o.Delivery.Zip,
			&o.Delivery.City,
			&o.Delivery.Address,
			&o.Delivery.Region,
			&o.Delivery.Email,
		)
		if err != nil {
			return nil, err
		}

		// достаем из Payment
		err = pool.QueryRow(ctx, `
			SELECT transaction, request_id, currency, provider, amount,
			       payment_dt, bank, delivery_cost, goods_total, custom_fee
			FROM payment WHERE order_uid=$1
		`, o.OrderUID).Scan(
			&o.Payment.Transaction,
			&o.Payment.RequestID,
			&o.Payment.Currency,
			&o.Payment.Provider,
			&o.Payment.Amount,
			&o.Payment.PaymentDT,
			&o.Payment.Bank,
			&o.Payment.DeliveryCost,
			&o.Payment.GoodsTotal,
			&o.Payment.CustomFee,
		)
		if err != nil {
			return nil, err
		}

		// достаем из Items
		itemRows, err := pool.Query(ctx, `
			SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
			FROM items WHERE order_uid=$1
		`, o.OrderUID)
		if err != nil {
			return nil, err
		}

		var items []models.Item
		for itemRows.Next() {
			var it models.Item
			if err := itemRows.Scan(
				&it.ChrtID,
				&it.TrackNumber,
				&it.Price,
				&it.RID,
				&it.Name,
				&it.Sale,
				&it.Size,
				&it.TotalPrice,
				&it.NmID,
				&it.Brand,
				&it.Status,
			); err != nil {
				itemRows.Close()
				return nil, err
			}
			items = append(items, it)
		}
		itemRows.Close()
		o.Items = items

		orders = append(orders, o)
	}

	return orders, nil
}

func GetOrder(ctx context.Context, orderUID string, pool *pgxpool.Pool) (models.Order, error) {
	var order models.Order

	// получаем из Order
	err := pool.QueryRow(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id,
		       delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1
		`, orderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		return order, fmt.Errorf("get order: %w", err)
	}

	// получаем из Delivery
	err = pool.QueryRow(ctx, `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery
		WHERE order_uid = $1
		`, orderUID).Scan(
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		return order, fmt.Errorf("get delivery: %w", err)
	}

	// получаем из Payment
	err = pool.QueryRow(ctx, `
		SELECT transaction, request_id, currency, provider, amount, payment_dt,
		       bank, delivery_cost, goods_total, custom_fee
		FROM payment
		WHERE order_uid = $1
	`, orderUID).Scan(
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDT,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		return order, fmt.Errorf("get payment: %w", err)
	}

	// получаем из Items
	rows, err := pool.Query(ctx, `
		SELECT chrt_id, track_number, price, rid, name, sale,
		       size, total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = $1
	`, orderUID)
	if err != nil {
		return order, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.RID,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return order, fmt.Errorf("scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}
