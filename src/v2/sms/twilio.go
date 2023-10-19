package sms

import (
	"github.com/bb-consent/api/src/config"
)

var TwilioConfig config.Twilio

// Init Initialize twilio
func Init(config *config.Configuration) {
	TwilioConfig = config.Twilio

}
