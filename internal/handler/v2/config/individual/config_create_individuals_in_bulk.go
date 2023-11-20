package individual

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/individual"
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

	// Read the header row
	header, err := reader.Read()
	if err != nil {
		m := "Failed to read CSV header"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	if len(header) == 2 {
		// Validate header columns
		expectedColumns := []string{"Name", "Email"}
		for i, col := range expectedColumns {
			if header[i] != col {
				m := fmt.Sprintf("Invalid column name in CSV header. Expected %s, got %s", col, header[i])
				common.HandleErrorV2(w, http.StatusBadRequest, m, nil)
				return
			}
		}
	} else if len(header) == 3 {
		// Validate header columns
		expectedColumns := []string{"Name", "Email", "Phone"}
		for i, col := range expectedColumns {
			if header[i] != col {
				m := fmt.Sprintf("Invalid column name in CSV header. Expected %s, got %s", col, header[i])
				common.HandleErrorV2(w, http.StatusBadRequest, m, nil)
				return
			}
		}
	} else {
		m := "Invalid number of columns in CSV header"
		common.HandleErrorV2(w, http.StatusBadRequest, m, nil)
		return
	}

	var individuals []individual.Individual

	// Read and parse the CSV data
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			m := "Error reading CSV data"
			common.HandleErrorV2(w, http.StatusBadRequest, m, err)
			return
		}

		if len(record) != 2 && len(record) != 3 {
			fmt.Println("Invalid CSV format")
			continue
		}

		individual := individual.Individual{
			Name:  record[0],
			Email: record[1],
		}

		if len(record) == 3 {
			individual.Phone = record[2]
		}

		individuals = append(individuals, individual)

	}
	if len(individuals) > 0 {
		go createIndividuals(individuals, organisationId)
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}

func createIndividuals(individuals []individual.Individual, organisationId string) {
	// Repository
	individualRepo := individual.IndividualRepository{}

	for _, individual := range individuals {
		// Check if individual exists If exist update individual
		// If individual doesn't exist, create individual
		individualRepo.Init(organisationId)
		u, err := individualRepo.GetIndividualByEmail(individual.Email)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// Register individual to keycloak
				iamId, err := iam.RegisterUser(individual.Email, individual.Name)
				if err != nil {
					log.Printf("Unable to register individual to keycloak: %v", individual.Name)
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
				log.Printf("Unable to fetch individual in db: %v", individual.Name)
			}
		} else {
			// Update individual in keycloak
			err := iam.UpdateIamUser(individual.Name, u.IamId)
			if err != nil {
				log.Printf("Unable to update individual to keycloak: %v", u.IamId)
			}
			u.Name = individual.Name
			if len(strings.TrimSpace(individual.Phone)) > 0 {
				u.Phone = individual.Phone
			}
			// Update individual to db
			_, err = individualRepo.Update(u)
			if err != nil {
				log.Printf("Unable to update individual in db: %v", u.Id)
			}
		}

	}
}
