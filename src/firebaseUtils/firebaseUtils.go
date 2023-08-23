package firebaseUtils

import (
	"time"

	"github.com/bb-consent/api/src/config"
)

var FirebaseConfig config.Firebase

var timeout time.Duration

func Init(config *config.Configuration) {
	FirebaseConfig = config.Firebase
	timeout = time.Duration(time.Duration(config.Iam.Timeout) * time.Second)
}
