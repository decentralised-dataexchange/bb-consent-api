package main

import (
	"log"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/orgtype"
	"github.com/bb-consent/api/src/user"
	"go.mongodb.org/mongo-driver/mongo"
)

func createOrganisationAdmin(config *config.Configuration) user.User {
	u, err := user.GetByEmail(config.User.Username)
	if err != nil {
		log.Println("Failed to get user, creating new user.")
		u, err = user.RegisterUser(config.User, config.Iam)
		if err != nil {
			log.Println("failed to create user")
			panic(err)
		}
	}

	return u
}

func createOrganisationType(config *config.Configuration) orgtype.OrgType {
	orgType, err := orgtype.GetFirstType()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Organization type doesn't exist, creating organization type.")
			orgType, err = orgtype.AddOrganizationType(config.Type)
			if err != nil {
				log.Println("failed to add organization")
				panic(err)
			}
		} else {
			log.Println("failed to find organization")
			panic(err)
		}
	}

	return orgType
}

func addOrganisationAdminRole(organisationAdminId string, organisationId string) {
	_, err := user.AddRole(organisationAdminId, user.Role{RoleID: common.GetRoleID("Admin"), OrgID: organisationId})
	if err != nil {
		log.Printf("Failed to update user : %v roles for org: %v", organisationAdminId, organisationId)
		panic(err)
	}
}

func createOrganisation(config *config.Configuration, orgType orgtype.OrgType, organisationAdminId string) org.Organization {
	organization, err := org.GetFirstOrganization()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Organization doesn't exist, creating organization.")
			organization, err = org.AddOrganization(config.Organization, orgType.ID.Hex(), organisationAdminId)
			if err != nil {
				log.Println("failed to add organization")
				panic(err)
			}
			// Add roles to organisation admin user
			addOrganisationAdminRole(organisationAdminId, organization.ID.Hex())

		} else {
			log.Println("failed to find organization")
			panic(err)
		}
	}

	return organization
}

func deleteAllOrganisationTypes() {
	typesCount, err := orgtype.GetTypesCount()
	if err != nil {
		log.Println("failed to count types")
		panic(err)
	}

	if typesCount > 1 {
		_, err := orgtype.DeleteAllTypes()
		if err != nil {
			log.Println("failed to delete organizations")
			panic(err)
		}
	}
}

func deleteAllOrganisations() {
	count, err := org.GetOrganizationsCount()
	if err != nil {
		log.Println("failed to count organization")
		panic(err)
	}
	if count > 1 {
		_, err := org.DeleteAllOrganizations()
		if err != nil {
			log.Println("failed to delete organizations")
			panic(err)
		}
	}
}

// SingleTenantConfiguration If the application starts in single tenant mode then create/update organisation, type, admin logic
func SingleTenantConfiguration(config *config.Configuration) {

	// Following is not allowed:
	// 1. Updation of organisation is not allowed
	// 2. Updation of organistaion type is not allowed
	// 3. Updation of organisation admin is not allowed
	// Note: Database has to be cleared if new organisation, type or admin has to be added

	// If there is more than 1 organisation or type, delete all (this is a temporary and will be removed later)
	deleteAllOrganisationTypes()
	deleteAllOrganisations()

	// Create an organisation admin
	organisationAdmin := createOrganisationAdmin(config)
	organisationAdminId := organisationAdmin.ID.Hex()

	// TODO: If wrong password is provided, the application panics
	user.GetOrganisationAdminToken(config.User, config.Iam)

	// Create organisation type
	orgType := createOrganisationType(config)

	// Create organisation
	createOrganisation(config, orgType, organisationAdminId)

}
