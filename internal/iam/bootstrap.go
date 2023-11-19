package iam

import (
	"context"
	"log"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bb-consent/api/internal/email"
)

// CreateRealmAndClientIfNotExists Creates a realm and a public OIDC client in Keycloak if not exists
func CreateRealmAndClientIfNotExists() {
	ctx := context.Background()

	// Authenticate with Keycloak
	client, token, err := GetAdminTokenAndClient()
	if err != nil {
		log.Fatalf("Failed to authenticate: %v\n", err)
	}

	// Check if the realm already exists
	existingRealm, err := client.GetRealm(ctx, token.AccessToken, IamConfig.Realm)
	if err != nil && existingRealm == nil {
		// Create realm and configure SMTP settings
		CreateRealmAndConfigureSMTP(client, ctx, token)

		log.Println("SMTP settings configured successfully.")

		// Create public OIDC client
		CreateOIDCClient(client, ctx, token)
	}
}

// CreateRealmAndConfigureSMTP Create realm with SMTP server configured
func CreateRealmAndConfigureSMTP(client *gocloak.GoCloak, ctx context.Context, token *gocloak.JWT) {

	newRealm := &gocloak.RealmRepresentation{
		Realm:   gocloak.StringP(IamConfig.Realm),
		Enabled: gocloak.BoolP(true),
		SMTPServer: &map[string]string{
			"replyToDisplayName": "iGrant.io",
			"starttls":           "false",
			"auth":               "true",
			"envelopeFrom":       "",
			"ssl":                "true",
			"password":           email.SMTPConfig.Password,
			"port":               "465",
			"host":               email.SMTPConfig.Host,
			"replyTo":            "support@igrant.io",
			"from":               email.SMTPConfig.AdminEmail,
			"fromDisplayName":    "iGrant.io",
			"user":               email.SMTPConfig.AdminEmail,
		},
		EmailTheme: gocloak.StringP("keycloak"),
	}

	if strings.Contains(IamConfig.URL, "keycloak:8080") {
		newRealm.Attributes = &map[string]string{
			"frontendUrl": "http://localhost:9090",
		}
	}

	realmID, err := client.CreateRealm(ctx, token.AccessToken, *newRealm)
	if err != nil {
		log.Fatalf("Failed to create realm: %v\n", err)
	}

	log.Printf("%s realm created in Keycloak\n", realmID)
}

// CreateOIDCClient Create a public OIDC client
func CreateOIDCClient(client *gocloak.GoCloak, ctx context.Context, token *gocloak.JWT) {
	oidcClient := &gocloak.Client{
		ClientID:     &IamConfig.ClientId,
		PublicClient: gocloak.BoolP(true),
		Enabled:      gocloak.BoolP(true),
		Protocol:     gocloak.StringP("openid-connect"),
	}

	createdClientID, err := client.CreateClient(ctx, token.AccessToken, IamConfig.Realm, *oidcClient)
	if err != nil {
		log.Fatalf("Failed to create OIDC client: %v\n", err)
	}

	log.Printf("OIDC client created with ID: %s\n", createdClientID)
}
