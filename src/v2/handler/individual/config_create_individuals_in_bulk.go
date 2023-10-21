package individual

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/individual"
)

// ConfigCreateIndividual
func ConfigCreateIndividualsInBulk(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	file, _, err := r.FormFile("csvFile")
	if err != nil {
		m := fmt.Sprintf("Failed to extract csv file for adding individuals")
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var users []individual.Individual

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

		user := individual.Individual{
			Name:  record[0],
			Email: record[1],
			Phone: record[2],
		}
		users = append(users, user)

	}

	for _, user := range users {
		createUser(user)
	}
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}

// func createUser(user User) {
// 	// Replace this function with your user creation logic
// 	// For example, you might create user accounts in a database or perform other operations.
// 	fmt.Printf("Creating user: Username=%s, Password=%s, Phone=%s\n", user.Username, user.Password, user.Phone)
// }
