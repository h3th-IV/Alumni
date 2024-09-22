package handlers

import (
	"net/http"
	"strconv"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

type sendGroupMessageHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewSendGroupMessageHandler(logger *zap.Logger, db mysql.Database) *sendGroupMessageHandler {
	return &sendGroupMessageHandler{
		logger: logger,
		db:     db,
	}
}

func (sgm *sendGroupMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sgm_resp := map[string]interface{}{}
	userInfo, err := utils.AuthenticateUser(r.Context(), sgm.logger, sgm.db)
	if err != nil {
		sgm_resp["err"] = "please sign in to access this page"
		sgm.logger.Error("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(sgm_resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}
	groupID, err := strconv.Atoi(r.FormValue("group_id"))
	if err != nil {
		sgm_resp["err"] = "unable to process request"
		sgm.logger.Error("err, invalid group ID")
		apiResponse(w, GetErrorResponseBytes(sgm_resp["err"], 30, nil), http.StatusBadRequest)
		return
	}

	exist, err := sgm.db.CheckGroupMembership(r.Context(), groupID, userInfo.Id)
	if err != nil {
		sgm_resp["err"] = "unable to confirm membership"
		sgm.logger.Error("err checking membership", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(sgm_resp["err"], 30, nil), http.StatusInternalServerError)
		return
	}

	if !exist {
		sgm_resp["warning"] = "you are not a member of this group"
		sgm.logger.Warn("user is not a mmeber")
		apiResponse(w, GetErrorResponseBytes(sgm_resp["err"], 30, nil), http.StatusForbidden)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		sgm_resp["err"] = "message body cannot be empty"
		sgm.logger.Warn("empty message body")
		apiResponse(w, GetErrorResponseBytes(sgm_resp, 30, nil), http.StatusBadRequest)
		return
	}

	if len(message) > 200 {
		sgm_resp["err"] = "message exceed the maximum lenght"
		sgm.logger.Warn("message exceed the maximum lenght")
		apiResponse(w, GetErrorResponseBytes(sgm_resp, 30, nil), http.StatusBadRequest)
		return
	}

	success, err := sgm.db.SendGroupMessage(r.Context(), groupID, userInfo.Id, message)
	if err != nil || !success {
		sgm_resp["err"] = "unable to send message"
		sgm.logger.Error("err sending message to group")
		apiResponse(w, GetErrorResponseBytes(sgm_resp["err"], 30, nil), http.StatusInternalServerError)
		return
	}

	sgm_resp["message"] = "message sent!"
	apiResponse(w, GetSuccessResponse(sgm_resp, 30), http.StatusOK)
}
