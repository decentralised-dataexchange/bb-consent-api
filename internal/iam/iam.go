package iam

import (
	"context"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bb-consent/api/internal/config"
)

var IamConfig config.Iam
var Timeout time.Duration

// IamInit Initialize the IAM
func Init(config *config.Configuration) {
	IamConfig = config.Iam
	Timeout = time.Duration(time.Duration(IamConfig.Timeout) * time.Second)
}

func GetAdminToken(username string, password string, realm string, client *gocloak.GoCloak) (*gocloak.JWT, error) {
	ctx := context.Background()
	token, err := client.LoginAdmin(ctx, username, password, realm)
	if err != nil {
		return token, err
	}

	return token, err
}
func GetToken(username string, password string, realm string, client *gocloak.GoCloak) (*gocloak.JWT, error) {
	ctx := context.Background()
	clientId := IamConfig.ClientId
	grantType := "password"
	token, err := client.GetToken(ctx, realm, gocloak.TokenOptions{Username: &username, Password: &password, ClientID: &clientId, GrantType: &grantType})
	if err != nil {
		return token, err
	}

	return token, err
}

func RefreshToken(clientId string, refreshToken string, realm string, client *gocloak.GoCloak) (*gocloak.JWT, error) {
	ctx := context.Background()
	grantType := "refresh_token"
	token, err := client.GetToken(ctx, realm, gocloak.TokenOptions{RefreshToken: &refreshToken, ClientID: &clientId, GrantType: &grantType})
	if err != nil {
		return token, err
	}

	return token, err
}

func ForgotPassword(iamId string) error {
	client := GetClient()
	token, err := GetAdminToken(IamConfig.AdminUser, IamConfig.AdminPassword, "master", client)
	if err != nil {
		return err
	}
	params := gocloak.ExecuteActionsEmail{
		UserID: &iamId,
		Actions: &[]string{
			"UPDATE_PASSWORD",
		},
	}

	err = client.ExecuteActionsEmail(context.Background(), token.AccessToken, IamConfig.Realm, params)
	return err
}

func GetClient() *gocloak.GoCloak {
	client := gocloak.NewClient(IamConfig.URL)
	return client
}

// UpdateIamUser Update user info on IAM server end.
func UpdateIamUser(Name string, iamID string) error {
	user := gocloak.User{
		ID:        &iamID,
		FirstName: &Name,
	}

	client := GetClient()

	token, err := GetAdminToken(IamConfig.AdminUser, IamConfig.AdminPassword, "master", client)
	if err != nil {
		return err
	}

	err = client.UpdateUser(context.Background(), token.AccessToken, IamConfig.Realm, user)
	if err != nil {
		return err
	}
	return nil
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

func ResetPassword(userId string, username string, currentPassword string, newPassword string) error {
	client := GetClient()

	ctx := context.Background()
	clientId := IamConfig.ClientId
	grantType := "password"
	_, err := client.GetToken(ctx, IamConfig.Realm, gocloak.TokenOptions{Username: &username, Password: &currentPassword, ClientID: &clientId, GrantType: &grantType})
	if err != nil {
		return err
	}

	token, err := GetAdminToken(IamConfig.AdminUser, IamConfig.AdminPassword, "master", client)
	if err != nil {
		return err
	}

	err = client.SetPassword(context.Background(), token.AccessToken, userId, IamConfig.Realm, newPassword, false)
	return err
}

func RegisterUser(email string, name string) (string, error) {
	user := gocloak.User{
		FirstName: &name,
		Email:     &email,
		Username:  &email,
		Enabled:   gocloak.BoolP(true),
	}

	client := GetClient()

	token, err := GetAdminToken(IamConfig.AdminUser, IamConfig.AdminPassword, "master", client)
	if err != nil {
		return "", err
	}

	iamId, err := client.CreateUser(context.Background(), token.AccessToken, IamConfig.Realm, user)
	if err != nil {
		return "", err
	}

	return iamId, nil
}

// UpdateIamIndividual Update individual info on IAM server end.
func UpdateIamIndividual(name string, iamID string, email string) error {
	user := gocloak.User{
		ID:        &iamID,
		FirstName: &name,
		Username:  &email,
		Email:     &email,
	}

	client := GetClient()

	token, err := GetAdminToken(IamConfig.AdminUser, IamConfig.AdminPassword, "master", client)
	if err != nil {
		return err
	}

	err = client.UpdateUser(context.Background(), token.AccessToken, IamConfig.Realm, user)
	if err != nil {
		return err
	}
	return nil
}

// UnregisterIndividual Unregisters an existing individual
func UnregisterIndividual(iamUserID string) error {
	client := GetClient()

	token, err := GetAdminToken(IamConfig.AdminUser, IamConfig.AdminPassword, "master", client)
	if err != nil {
		return err
	}

	err = client.DeleteUser(context.Background(), token.AccessToken, IamConfig.Realm, iamUserID)
	return err
}

// LogoutUser logout user
func LogoutUser(accessToken string, refreshToken string) error {
	client := GetClient()

	err := client.LogoutPublicClient(context.Background(), IamConfig.ClientId, IamConfig.Realm, accessToken, refreshToken)

	return err
}
