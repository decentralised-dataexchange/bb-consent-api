package individual

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/iam"
	"github.com/bb-consent/api/src/v2/individual"
	"go.mongodb.org/mongo-driver/mongo"
)

// ConfigCreateIndividual
func ConfigCreateIndividualsInBulk(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	file, _, err := r.FormFile("individuals")
	if err != nil {
		m := "Failed to extract csv file for adding individuals"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var individuals []individual.Individual

	// Read and parse the CSV data
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		if len(record) != 3 {
			fmt.Println("Invalid CSV format")
			continue
		}

		individual := individual.Individual{
			Name:  record[0],
			Email: record[1],
			Phone: record[2],
		}
		individuals = append(individuals, individual)

	}
	go createIndividuals(individuals, organisationId)

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}

func createIndividuals(individuals []individual.Individual, organisationId string) {
	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	for _, individual := range individuals {
		// Check if individual exists If exist update individual
		// If individual doesn't exist, create individual
		u, err := individualRepo.GetIndividualByEmail(individual.Email)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// Register individual to keycloak
				iamId, err := iam.RegisterUser(individual.Email, individual.Name)
				if err != nil {
					log.Printf("Unable to register individual to keycloak: %v", u.Name)
				}

				individual.IsDeleted = false
				individual.OrganisationId = organisationId
				individual.IamId = iamId
				// Save individual to db
				_, err = individualRepo.Add(individual)
				if err != nil {
					log.Printf("Unable to save individual in db: %v", individual.Email)
				}
			} else {
				log.Printf("Unable to fetchindividual in db: %v", individual.Name)
			}
		} else {
			// Update individual in keycloak
			err := iam.UpdateIamUser(individual.Name, u.IamId)
			if err != nil {
				log.Printf("Unable to update individual to keycloak: %v", u.IamId)
			}
			u.Name = individual.Name
			u.Phone = individual.Phone
			// Update individual to db
			_, err = individualRepo.Update(u)
			if err != nil {
				log.Printf("Unable to update individual in db: %v", u.Id)
			}
		}

	}
}
