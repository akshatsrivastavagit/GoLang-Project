package webhooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotifyLowStock(t *testing.T) {
	err := NotifyLowStock("test", 0)
	assert.Nil(t, err) // No webhook URL configured
} 