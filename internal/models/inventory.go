package models

import (
	"encoding/json"
	"time"
)

type StockUpdate struct {
	SKU         string    `json:"sku"`
	WarehouseID int       `json:"warehouse_id"`
	Quantity    int       `json:"quantity"`
	Timestamp   time.Time `json:"timestamp"`
}

func (s *StockUpdate) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *StockUpdate) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

type Order struct {
	SKU      string `json:"sku"`
	Channel  string `json:"channel"`
	Quantity int    `json:"quantity"`
}

func (o *Order) MarshalBinary() ([]byte, error) {
	return json.Marshal(o)
}

func (o *Order) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, o)
}

type InventoryTransaction struct {
	ID          int       `json:"id"`
	SKU         string    `json:"sku"`
	WarehouseID int       `json:"warehouse_id"`
	Change      int       `json:"change"`
	Type        string    `json:"type"`
	Channel     string    `json:"channel"`
	Timestamp   time.Time `json:"timestamp"`
}

type StockLevel struct {
	SKU         string `json:"sku"`
	WarehouseID int    `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
}

func (s *StockLevel) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *StockLevel) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
} 