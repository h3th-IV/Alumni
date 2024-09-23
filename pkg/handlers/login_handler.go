package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/model"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

var _ http.Handler = &loginHandler{}
var loginTTL = 30

type loginHandler struct {
	logger      *zap.Logger
	mysqlclient mysql.Database
}

func NewLoginHandler(logger *zap.Logger, mysqlclient mysql.Database) *loginHandler {
	return &loginHandler{
		logger:      logger,
		mysqlclient: mysqlclient,
	}
}

func (handler *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		loginres = map[string]interface{}{}
		Login    *model.Login
	)
	if err := json.NewDecoder(r.Body).Decode(&Login); err != nil {
		loginres["err"] = "unable to process request"
		handler.logger.Error("err decoding JSON object", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(loginres, loginTTL, nil), http.StatusNotFound)
		return
	}
	email := Login.Email
	password := Login.Password

	if email == "" || password == "" {
		loginres["err"] = "email or password not provided"
		handler.logger.Error("email or password not provided")
		apiResponse(w, GetErrorResponseBytes(loginres["err"], loginTTL, nil), http.StatusNotFound)
		return
	}
	checkuser, err := handler.mysqlclient.GetUserByEmail(r.Context(), email)
	if err != nil {
		loginres["err"] = "user does not exist"
		handler.logger.Error("user does not exist", zap.Any("checkuser", err))
		apiResponse(w, GetErrorResponseBytes(loginres["err"], loginTTL, nil), http.StatusNotFound)
		return
	}
	if checkuser != nil {
		if checkuser.Id > 0 {
			handler.logger.Debug("found user", zap.Bool("user found", true))
			_ = CheckPasswordHash(password, checkuser.Password)
			loginnow, err := handler.mysqlclient.CheckUser(r.Context(), email, checkuser.Password)
			if err != nil {
				loginres["err"] = "email or password incorrect"
				handler.logger.Error("email or password incorrect", zap.Any("login response", "email or password incorrect"))
				apiResponse(w, GetErrorResponseBytes(loginres["err"], loginTTL, nil), http.StatusUnauthorized)
				return
			}
			if loginnow != nil {
				jwt, err := utils.GenerateToken(loginnow, 2*time.Hour, utils.JWTISSUER, utils.MYSTIC)
				if err != nil {
					loginres["err"] = "unable to authenticate user"
					handler.logger.Error("err generating auth token")
					apiResponse(w, GetErrorResponseBytes(loginres["err"], loginTTL, nil), http.StatusInternalServerError)
					return
				}
				loginres["username"] = loginnow.Username
				loginres["email"] = loginnow.Email
				loginres["phone"] = loginnow.Phone
				loginres["degree"] = loginnow.Degree
				loginres["grad_year"] = loginnow.GradYear
				loginres["current_job"] = loginnow.CurrentJob
				loginres["profile_picture"] = loginnow.ProfilePicture
				loginres["linkedin_profile"] = loginnow.LinkedinProfile
				loginres["twitter_profile"] = loginnow.TwitterProfile
				loginres["jwt_token"] = jwt
				apiResponse(w, GetSuccessResponse(loginres, loginTTL), http.StatusOK)
			}
		}
	}
}
