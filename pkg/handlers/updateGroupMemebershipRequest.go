package handlers

import (
	"encoding/json"
	"log"
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

	//admin auth
	userInfo, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.db) //get the admin
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

	var requestMembership struct {
		Status string `json:"status"`
		Email  string `json:"email"` //the new member email
	}
	admin, err := handler.db.GetGroupCreator(r.Context(), groupID)
	if err != nil {
		resp["err"] = "unable to proceed"
		handler.logger.Error("err fetching group admin")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}
	if admin.Id != userInfo.Id {
		log.Printf("%d, %d", admin.Id, userInfo.Id)
		resp["err"] = "you are not allowed to accept or decline members to this group"
		handler.logger.Warn("admin, user IDs do not match")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&requestMembership); err != nil {
		resp["err"] = "unable to process request"
		handler.logger.Error("error decoding JSON request", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusBadRequest)
		return
	}

	email := requestMembership.Email
	if email == "" {
		resp["err"] = "recipient email is required"
		handler.logger.Error("email is missing")
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusBadRequest)
		return
	}
	newMember, err := handler.db.GetUserByEmail(r.Context(), email)
	if err != nil {
		resp["err"] = "user not found"
		handler.logger.Error("err", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusBadRequest)
		return
	}

	success, err := handler.db.UpdateGroupMembershipRequest(r.Context(), requestMembership.Status, groupID, newMember.Id)
	if err != nil || !success {
		resp["err"] = "unable to update membership request"
		handler.logger.Error("failed to update membership", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}

	resp["message"] = "membership request updated successfully"
	apiResponse(w, GetSuccessResponse(resp, 30), http.StatusOK)
}
