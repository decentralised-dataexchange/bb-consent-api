package tenant

import (
	"log"
	"strings"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/fixture"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/orgtype"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func createOrganisationAdmin(config *config.Configuration) user.User {
	organization, _ := org.GetFirstOrganization()
	if len(organization.Admins) == 0 {
		log.Println("Failed to get user, creating new user.")
		u, err := user.RegisterUser(config.User, config.Iam)
		if err != nil {
			log.Println("failed to create user")
			panic(err)
		}
		return u
	}

	u, err := user.Get(organization.Admins[0].UserID)
	if err != nil {
		log.Println("failed to find organisation admin")
		panic(err)
	}

	return u
}

func createOrganisationType(config *config.Configuration) orgtype.OrgType {
	orgType, err := orgtype.GetFirstType()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Organization type doesn't exist, creating organization type.")
			orgType, err = orgtype.AddOrganizationType(config.Policy)
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

func deleteAllPolicies() {

	// Get first organisation
	o, err := org.GetFirstOrganization()
	if err != nil {
		panic(err)
	}

	// Repository
	prepo := policy.PolicyRepository{}
	prepo.Init(o.ID.Hex())

	count, err := prepo.GetPolicyCountByOrganisation()
	if err != nil {
		log.Println("failed to count policy")
		panic(err)
	}
	if count > 1 {
		err := policy.DeleteAllPolicies(o.ID.Hex())
		if err != nil {
			log.Println("failed to delete policies")
			panic(err)
		}
	}
}

// createDefaultPolicy
func createDefaultPolicy(config *config.Configuration, org org.Organization, orgAdminId string) (policy.Policy, error) {

	var newPolicy policy.Policy
	newPolicy.Id = primitive.NewObjectID()
	newPolicy.Name = config.Policy.Name
	newPolicy.Url = org.PolicyURL
	newPolicy.Jurisdiction = org.Location
	newPolicy.IndustrySector = config.Policy.IndustrySector
	newPolicy.DataRetentionPeriodDays = 1095
	newPolicy.GeographicRestriction = config.Policy.GeographicRestriction
	newPolicy.StorageLocation = config.Policy.StorageLocation
	newPolicy.ThirdPartyDataSharing = true
	newPolicy.OrganisationId = org.ID.Hex()
	newPolicy.IsDeleted = false

	version := common.IntegerToSemver(1)
	newPolicy.Version = version

	if len(strings.TrimSpace(config.Policy.Url)) > 1 {
		newPolicy.Url = config.Policy.Url
		updateOrganisationPolicyUrl(config.Policy.Url, org)
	}

	// Update revision
	_, err := revision.UpdateRevisionForPolicy(newPolicy, orgAdminId)
	if err != nil {
		return newPolicy, err
	}

	return newPolicy, nil
}

func updateOrganisationPolicyUrl(policyUrl string, organisation org.Organization) {
	organisation.PolicyURL = policyUrl
	_, err := org.Update(organisation)
	if err != nil {
		log.Println("failed to update policy url for organisation")
		panic(err)
	}
}

func createGlobalPolicy(config *config.Configuration, orgAdminId string) {
	// Get first organisation
	o, err := org.GetFirstOrganization()
	if err != nil {
		panic(err)
	}

	// Repository
	prepo := policy.PolicyRepository{}
	prepo.Init(o.ID.Hex())

	policyCount, _ := prepo.GetPolicyCountByOrganisation()
	if policyCount == 0 {
		log.Println("Failed to get global policy, creating new global policy.")
		createdPolicy, err := createDefaultPolicy(config, o, orgAdminId)
		if err != nil {
			log.Println("failed to create global policy")
			panic(err)
		}

		_, err = prepo.Add(createdPolicy)
		if err != nil {
			log.Println("failed to create global policy")
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

	// Create organisation type
	orgType := createOrganisationType(config)

	// Create organisation
	createOrganisation(config, orgType, organisationAdminId)

	// delete all policies
	// deleteAllPolicies()

	// Load image assets for organisation
	err := fixture.LoadImageAssetsForSingleTenantConfiguration()
	if err != nil {
		log.Println("Error occured while loading image assets for organisation")
	}

	// Create global policy
	createGlobalPolicy(config, organisationAdminId)

}
