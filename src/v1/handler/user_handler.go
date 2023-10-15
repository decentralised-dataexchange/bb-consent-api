package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
)

// GetUserByIamID Gets a single user by given Iamid
func GetUserByIamID(iamID string) (user.User, error) {
	user, err := user.GetByIamID(iamID)

	if err != nil {
		return user, err
	}
	return user, err
}

// GetUserRoles Get User roles as int array
func GetUserRoles(userRoles []user.Role) (roles []int) {
	for _, role := range userRoles {
		roles = append(roles, role.RoleID)
	}
	return
}

// GetUser Gets a single user by ID
func GetUser(userID string) (user.User, error) {
	user, err := user.Get(userID)

	if err != nil {
		return user, err
	}
	return user, err
}

type userResp struct {
	User user.User
}

// GetCurrentUser Gets the currernt authenticated user details
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)

	u, err := user.Get(userID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch user by id:%v", userID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(userResp{u})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

type updateUserReq struct {
	Phone string
	Name  string
}

// UpdateCurrentUser Updates the currernt authenticated user details
func UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)
	var upReq updateUserReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &upReq)

	u, err := user.Get(userID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch user by id:%v", userID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	if strings.TrimSpace(upReq.Phone) != "" {
		u.Phone = upReq.Phone
	}

	if strings.TrimSpace(upReq.Name) != "" {
		u.Name = upReq.Name
	}

	u, err = user.Update(userID, u)
	if err != nil {
		m := fmt.Sprintf("Failed to update user by id:%v", userID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	if upReq.Name != "" {
		err = UpdateIamUser(u.Name, token.GetIamID(r))
		if err != nil {
			//TODO: revert the changes to the local db as well.
			m := fmt.Sprintf("Failed to update IAM user by id:%v", userID)
			common.HandleError(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	response, _ := json.Marshal(userResp{u})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

type appRegsiterResp struct {
	User user.User
}

type deviceRegisterReq struct {
	DeviceToken string `valid:"required"`
}

// UserClientRegister Registers the user device register token
func UserClientRegister(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)

	var regReq deviceRegisterReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &regReq)

	valid, err := govalidator.ValidateStruct(regReq)
	if !valid {
		log.Printf("Failed to register device token for user:%v", userID)
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	log.Printf("Reg token: %v path: %v", regReq.DeviceToken, r.URL.Path)

	var clientType = common.ClientTypeAndroid
	if strings.Contains(r.URL.Path, "ios") {
		clientType = common.ClientTypeIos
	}

	client := user.ClientInfo{Token: regReq.DeviceToken, Type: clientType}

	u, err := user.UpdateClientDeviceInfo(userID, client)
	if err != nil {
		m := fmt.Sprintf("Failed to update registration token for user id:%v", userID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(appRegsiterResp{u})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}
