package handlerv2

import (
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
)

// DeleteAPIKey User forgot the password, need to reset the password
func DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)

	err := user.UpdateAPIKey(userID, "")
	if err != nil {
		m := fmt.Sprintf("Failed to remove apiKey for user:%v err:%v", token.GetUserName(r), err)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
