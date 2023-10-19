package iam

import (
	"context"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bb-consent/api/src/config"
)

var IamConfig config.Iam
var Timeout time.Duration

// IamInit Initialize the IAM
func Init(config *config.Configuration) {
	IamConfig = config.Iam
	Timeout = time.Duration(time.Duration(IamConfig.Timeout) * time.Second)
}

func GetToken(username string, password string, realm string, client *gocloak.GoCloak) (*gocloak.JWT, error) {
	ctx := context.Background()
	token, err := client.LoginAdmin(ctx, username, password, realm)
	if err != nil {
		return token, err
	}

	return token, err
}

func GetClient() *gocloak.GoCloak {
	client := gocloak.NewClient(IamConfig.URL)
	return client
}

type IamToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
}

type IamError struct {
	ErrorType string `json:"error"`
	Error     string `json:"error_description"`
}
