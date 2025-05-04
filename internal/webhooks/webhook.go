package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type LowStockPayload struct {
	SKU         string    `json:"sku"`
	WarehouseID int       `json:"warehouse_id"`
	Stock       int       `json:"stock"`
	Timestamp   time.Time `json:"timestamp"`
}

type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Color     string   `json:"color"`
	Title     string   `json:"title"`
	Text      string   `json:"text"`
	Fields    []Field  `json:"fields"`
	Timestamp int64    `json:"ts"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func NotifyLowStock(sku string, warehouseID int, stock int) error {
	slackURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackURL == "" {
		log.Printf("SLACK_WEBHOOK_URL not set in environment variables")
		return fmt.Errorf("SLACK_WEBHOOK_URL not set")
	}
	log.Printf("Slack webhook URL found: %s", slackURL)

	// Create Slack message
	message := SlackMessage{
		Text: "⚠️ Low Stock Alert",
		Attachments: []Attachment{
			{
				Color: "danger",
				Title: "Low Stock Alert",
				Text:  "Stock level has fallen below threshold",
				Fields: []Field{
					{
						Title: "SKU",
						Value: sku,
						Short: true,
					},
					{
						Title: "Warehouse ID",
						Value: fmt.Sprintf("%d", warehouseID),
						Short: true,
					},
					{
						Title: "Current Stock",
						Value: fmt.Sprintf("%d", stock),
						Short: true,
					},
					{
						Title: "Threshold",
						Value: "10",
						Short: true,
					},
				},
				Timestamp: time.Now().Unix(),
			},
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling Slack message: %v", err)
		return fmt.Errorf("error marshaling message: %v", err)
	}

	log.Printf("Preparing to send Slack notification with payload: %s", string(data))

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	log.Printf("Sending POST request to Slack webhook URL: %s", slackURL)
	req, err := http.NewRequest("POST", slackURL, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending Slack notification: %v", err)
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Slack webhook response - Status: %d, Body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("Slack notification failed with status: %d", resp.StatusCode)
		return fmt.Errorf("slack notification failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Slack notification sent successfully")
	return nil
} 