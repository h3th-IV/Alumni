package handlers

import (
	"net/http"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

var _ http.Handler = &profileHandler{}
var profileTTL = 3600

type profileHandler struct {
	logger      *zap.Logger
	mysqlclient mysql.Database
}

func NewProfileHandler(logger *zap.Logger, mysqlclient mysql.Database) *profileHandler {
	return &profileHandler{
		logger:      logger,
		mysqlclient: mysqlclient,
	}
}

func (handler *profileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	profileres := map[string]interface{}{}
	userInfo, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.mysqlclient)
	if err != nil {
		profileres["err"] = "please sign in to access this page"
		handler.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(profileres["err"], profileTTL, nil), http.StatusUnauthorized)
		return
	}
	profileres["id"] = userInfo.Id
	profileres["username"] = userInfo.Username
	profileres["email"] = userInfo.Email
	profileres["phone"] = userInfo.Phone
	profileres["degree"] = userInfo.Degree
	profileres["current_job"] = userInfo.CurrentJob
	profileres["linkedin_profile"] = userInfo.LinkedinProfile
	profileres["twitter_profile"] = userInfo.TwitterProfile
	apiResponse(w, GetSuccessResponse(profileres, profileTTL), http.StatusOK)
}
