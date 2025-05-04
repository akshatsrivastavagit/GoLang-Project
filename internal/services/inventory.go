package services

import (
	"context"
	"errors"
	"omnichannel_inventory/internal/db"
	"omnichannel_inventory/internal/models"
	"time"
)

type InventoryService struct {
	db    DB
	redis Redis
}

type DB interface {
	Exec(ctx context.Context, sql string, args ...interface{}) error
	Query(ctx context.Context, sql string, args ...interface{}) (db.Rows, error)
}

type Redis interface {
	Publish(ctx context.Context, channel string, message interface{}) error
}

func NewInventoryService(db DB, redis Redis) *InventoryService {
	return &InventoryService{
		db:    db,
		redis: redis,
	}
}

func (s *InventoryService) AddOrUpdateStock(ctx context.Context, update models.StockUpdate) error {
	// Update stock in database
	sql := `
		INSERT INTO stock_levels (sku, warehouse_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (sku, warehouse_id) DO UPDATE
		SET quantity = stock_levels.quantity + $3
	`
	if err := s.db.Exec(ctx, sql, update.SKU, update.WarehouseID, update.Quantity); err != nil {
		return err
	}

	// Record transaction
	sql = `
		INSERT INTO inventory_transactions (sku, warehouse_id, change, type, timestamp)
		VALUES ($1, $2, $3, 'stock_update', $4)
	`
	if err := s.db.Exec(ctx, sql, update.SKU, update.WarehouseID, update.Quantity, time.Now()); err != nil {
		return err
	}

	// Publish event
	return s.redis.Publish(ctx, "inventory_updates", update)
}

func (s *InventoryService) GetConsolidatedStock(ctx context.Context, sku string) ([]models.StockLevel, error) {
	sql := `
		SELECT sku, warehouse_id, quantity
		FROM stock_levels
		WHERE sku = $1
	`
	rows, err := s.db.Query(ctx, sql, sku)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []models.StockLevel
	for rows.Next() {
		var level models.StockLevel
		if err := rows.Scan(&level.SKU, &level.WarehouseID, &level.Quantity); err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}
	return levels, nil
}

func (s *InventoryService) SimulateOrder(ctx context.Context, order models.Order) error {
	// Get available stock
	sql := `
		SELECT warehouse_id, quantity
		FROM stock_levels
		WHERE sku = $1 AND quantity > 0
		ORDER BY quantity DESC
	`
	rows, err := s.db.Query(ctx, sql, order.SKU)
	if err != nil {
		return err
	}
	defer rows.Close()

	remaining := order.Quantity
	for rows.Next() {
		var warehouseID, quantity int
		if err := rows.Scan(&warehouseID, &quantity); err != nil {
			return err
		}

		toDeduct := min(remaining, quantity)
		if toDeduct > 0 {
			// Update stock
			sql = `
				UPDATE stock_levels
				SET quantity = quantity - $1
				WHERE sku = $2 AND warehouse_id = $3
			`
			if err := s.db.Exec(ctx, sql, toDeduct, order.SKU, warehouseID); err != nil {
				return err
			}

			// Record transaction
			sql = `
				INSERT INTO inventory_transactions (sku, warehouse_id, change, type, channel, timestamp)
				VALUES ($1, $2, $3, 'order', $4, $5)
			`
			if err := s.db.Exec(ctx, sql, order.SKU, warehouseID, -toDeduct, order.Channel, time.Now()); err != nil {
				return err
			}

			remaining -= toDeduct
			if remaining == 0 {
				break
			}
		}
	}

	if remaining > 0 {
		return errors.New("insufficient stock")
	}

	// Publish event
	return s.redis.Publish(ctx, "inventory_updates", order)
}

func (s *InventoryService) GetInventoryHistory(ctx context.Context, sku string) ([]models.InventoryTransaction, error) {
	sql := `
		SELECT id, sku, warehouse_id, change, type, channel, timestamp
		FROM inventory_transactions
		WHERE sku = $1
		ORDER BY timestamp DESC
	`
	rows, err := s.db.Query(ctx, sql, sku)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.InventoryTransaction
	for rows.Next() {
		var t models.InventoryTransaction
		if err := rows.Scan(&t.ID, &t.SKU, &t.WarehouseID, &t.Change, &t.Type, &t.Channel, &t.Timestamp); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 