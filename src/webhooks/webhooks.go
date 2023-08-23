package webhooks

import (
	"github.com/bb-consent/api/src/config"
)

// WebhooksConfiguration Stores webhooks configuration
var WebhooksConfiguration config.WebhooksConfig

// Init Initializes webhooks configuration
func Init(config *config.Configuration) {
	WebhooksConfiguration = config.Webhooks
}
