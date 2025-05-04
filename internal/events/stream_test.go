package events

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishInventoryEvent(t *testing.T) {
	event := InventoryEvent{
		SKU:         "test",
		WarehouseID: 1,
		Change:      10,
		Channel:     "test",
		Reason:      "test",
	}
	err := PublishInventoryEvent(context.Background(), event)
	assert.NotNil(t, err) // Redis not connected in test
}

func TestStartInventoryEventConsumer(t *testing.T) {
	eventsCh := make(chan InventoryEvent)
	StartInventoryEventConsumer(context.Background(), eventsCh)
	// Test passes if it doesn't panic
} 