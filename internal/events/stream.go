package events

import (
	"context"
	"fmt"
	"log"
	"time"

	"omnichannel_inventory/internal/db"
	"omnichannel_inventory/internal/webhooks"

	"github.com/go-redis/redis/v8"
)

const (
	InventoryStream = "inventory_events"
	EventBufferSize = 100
	LowStockThreshold = 10
)

type InventoryEvent struct {
	SKU         string
	WarehouseID int
	Change      int
	Channel     string
	Reason      string
}

type EventProcessor struct {
	eventsCh chan InventoryEvent
	doneCh   chan struct{}
}

func NewEventProcessor() *EventProcessor {
	return &EventProcessor{
		eventsCh: make(chan InventoryEvent, EventBufferSize),
		doneCh:   make(chan struct{}),
	}
}

func (p *EventProcessor) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case event := <-p.eventsCh:
				// Process event asynchronously
				go p.processEvent(ctx, event)
			case <-p.doneCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *EventProcessor) Stop() {
	close(p.doneCh)
}

func (p *EventProcessor) processEvent(ctx context.Context, event InventoryEvent) {
	log.Printf("Received event: SKU=%s, WarehouseID=%d, Change=%d", event.SKU, event.WarehouseID, event.Change)
	
	// Check for low stock and trigger webhook if needed
	if event.Change < 0 {
		log.Printf("Processing negative stock change event for SKU %s in warehouse %d", event.SKU, event.WarehouseID)
		
		// Get current stock level
		sql := `
			SELECT quantity
			FROM stock_levels
			WHERE sku = $1 AND warehouse_id = $2
		`
		rows, err := db.GetDB().Query(ctx, sql, event.SKU, event.WarehouseID)
		if err != nil {
			log.Printf("Error querying stock level: %v", err)
			return
		}
		defer rows.Close()

		if rows.Next() {
			var quantity int
			if err := rows.Scan(&quantity); err != nil {
				log.Printf("Error scanning stock level: %v", err)
				return
			}

			log.Printf("Current stock for SKU %s in warehouse %d: %d (Threshold: %d)", 
				event.SKU, event.WarehouseID, quantity, LowStockThreshold)
			
			if quantity < LowStockThreshold {
				log.Printf("Low stock detected! Triggering webhook for SKU %s (Current: %d, Threshold: %d)", 
					event.SKU, quantity, LowStockThreshold)
				// Trigger low stock webhook
				if err := webhooks.NotifyLowStock(event.SKU, event.WarehouseID, quantity); err != nil {
					log.Printf("Failed to send low stock notification: %v", err)
				} else {
					log.Printf("Low stock notification sent successfully")
				}
			} else {
				log.Printf("Stock level (%d) is above threshold (%d), no notification needed", 
					quantity, LowStockThreshold)
			}
		} else {
			log.Printf("No stock level found for SKU %s in warehouse %d", event.SKU, event.WarehouseID)
		}
	} else {
		log.Printf("Positive stock change, no notification needed")
	}
}

func PublishInventoryEvent(ctx context.Context, event InventoryEvent) error {
	log.Printf("Publishing inventory event: SKU=%s, WarehouseID=%d, Change=%d", event.SKU, event.WarehouseID, event.Change)
	_, err := db.GetRedis().XAdd(ctx, &redis.XAddArgs{
		Stream: InventoryStream,
		Values: map[string]interface{}{
			"sku":          event.SKU,
			"warehouse_id": event.WarehouseID,
			"change":       event.Change,
			"channel":      event.Channel,
			"reason":       event.Reason,
		},
	}).Result()
	if err != nil {
		log.Printf("Error publishing event to Redis: %v", err)
	} else {
		log.Printf("Event published to Redis successfully")
	}
	return err
}

func StartInventoryEventConsumer(ctx context.Context, processor *EventProcessor) {
	go func() {
		lastID := "$"
		for {
			select {
			case <-ctx.Done():
				return
			default:
				streams, err := db.GetRedis().XRead(ctx, &redis.XReadArgs{
					Streams: []string{InventoryStream, lastID},
					Block:   0,
					Count:   10,
				}).Result()
				if err != nil && err != redis.Nil {
					time.Sleep(time.Second)
					continue
				}
				for _, stream := range streams {
					for _, msg := range stream.Messages {
						lastID = msg.ID
						event := InventoryEvent{
							SKU:         fmt.Sprint(msg.Values["sku"]),
							WarehouseID: atoi(msg.Values["warehouse_id"]),
							Change:      atoi(msg.Values["change"]),
							Channel:     fmt.Sprint(msg.Values["channel"]),
							Reason:      fmt.Sprint(msg.Values["reason"]),
						}
						processor.eventsCh <- event
					}
				}
			}
		}
	}()
}

func atoi(v interface{}) int {
	if v == nil {
		return 0
	}
	s, ok := v.(string)
	if !ok {
		return 0
	}
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
} 