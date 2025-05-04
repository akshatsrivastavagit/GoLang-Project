package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddOrUpdateStock(t *testing.T) {
	err := AddOrUpdateStock(context.Background(), "test", 1, 10)
	assert.Nil(t, err)
}

func TestGetConsolidatedStock(t *testing.T) {
	stock, err := GetConsolidatedStock(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, 0, stock)
}

func TestSimulateOrder(t *testing.T) {
	err := SimulateOrder(context.Background(), "test", "amazon", 1)
	assert.Nil(t, err)
}

func TestGetInventoryHistory(t *testing.T) {
	history, err := GetInventoryHistory(context.Background(), "test")
	assert.Nil(t, err)
	assert.Nil(t, history)
} 