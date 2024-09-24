package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

var _ http.Handler = &UpdateMembershipHandler{}

type UpdateMembershipHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewUpdateMembershipHandler(logger *zap.Logger, db mysql.Database) *UpdateMembershipHandler {
	return &UpdateMembershipHandler{
		logger: logger,
		db:     db,
	}
}

func (handler *UpdateMembershipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}

	userInfo, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.db)
	if err != nil {
		resp["err"] = "please sign in"
		handler.logger.Error("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	groupIdStr := r.URL.Query().Get("group_id")
	groupID, err := strconv.Atoi(groupIdStr)
	if err != nil {
		resp["err"] = "invalid group ID"
		handler.logger.Error("error converting groupID to int", zap.String("group_id", groupIdStr))
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusBadRequest)
		return
	}

	var status struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		resp["err"] = "unable to process request"
		handler.logger.Error("error decoding JSON request", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusBadRequest)
		return
	}

	success, err := handler.db.UpdateGroupMembershipRequest(r.Context(), status.Status, groupID, userInfo.Id)
	if err != nil || !success {
		resp["err"] = "unable to update membership request"
		handler.logger.Error("failed to update membership", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}

	resp["message"] = "membership request updated successfully"
	apiResponse(w, GetSuccessResponse(resp, 30), http.StatusOK)
}
