package handlers

import (
	"errors"
	"net/http"

	"omnichannel_inventory/internal/models"
	"omnichannel_inventory/internal/services"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidSKU         = errors.New("invalid SKU")
	ErrInvalidWarehouseID = errors.New("invalid warehouse ID")
	ErrInvalidQuantity    = errors.New("invalid quantity")
	ErrInvalidChannel     = errors.New("invalid channel")
)

var inventoryService *services.InventoryService

func SetInventoryService(service *services.InventoryService) {
	inventoryService = service
}

// @Summary Add or update stock for a product in a warehouse
// @Description Add or update stock for a product in a warehouse
// @Tags inventory
// @Accept json
// @Produce json
// @Param request body models.StockUpdate true "Stock update request"
// @Success 200 {object} map[string]interface{}
// @Router /inventory/add_or_update [post]
func AddOrUpdateStock(c *gin.Context) {
	var update models.StockUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if update.SKU == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidSKU.Error()})
		return
	}
	if update.WarehouseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidWarehouseID.Error()})
		return
	}
	if update.Quantity == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidQuantity.Error()})
		return
	}

	if err := inventoryService.AddOrUpdateStock(c.Request.Context(), update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stock updated successfully"})
}

// @Summary Get consolidated stock for a product
// @Description Get consolidated stock for a product across all warehouses
// @Tags inventory
// @Produce json
// @Param sku path string true "Product SKU"
// @Success 200 {object} []models.StockLevel
// @Router /inventory/stock/{sku} [get]
func GetConsolidatedStock(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidSKU.Error()})
		return
	}

	levels, err := inventoryService.GetConsolidatedStock(c.Request.Context(), sku)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, levels)
}

// @Summary Simulate an order event
// @Description Simulate an order event from a sales channel
// @Tags inventory
// @Accept json
// @Produce json
// @Param request body models.Order true "Order request"
// @Success 200 {object} map[string]interface{}
// @Router /inventory/order [post]
func SimulateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if order.SKU == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidSKU.Error()})
		return
	}
	if order.Channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidChannel.Error()})
		return
	}
	if order.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidQuantity.Error()})
		return
	}

	if err := inventoryService.SimulateOrder(c.Request.Context(), order); err != nil {
		if err.Error() == "insufficient stock" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order processed successfully"})
}

// @Summary Get inventory history for a product
// @Description Get inventory history for a product
// @Tags inventory
// @Produce json
// @Param sku path string true "Product SKU"
// @Success 200 {object} []models.InventoryTransaction
// @Router /inventory/history/{sku} [get]
func GetInventoryHistory(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidSKU.Error()})
		return
	}

	transactions, err := inventoryService.GetInventoryHistory(c.Request.Context(), sku)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
} 