package handlers

import (
	"net/http"
	"strconv"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

var _ http.Handler = &RequestMembershipHandler{}

type RequestMembershipHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewRequestMembershipHandler(logger *zap.Logger, db mysql.Database) *RequestMembershipHandler {
	return &RequestMembershipHandler{
		logger: logger,
		db:     db,
	}
}

func (handler *RequestMembershipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	userInfo, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.db)
	if err != nil {
		resp["err"] = "please sign in to request membership"
		handler.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	groupId := r.URL.Query().Get("forum_id")
	if groupId == "" {
		resp["err"] = "group ID is required"
		handler.logger.Error("group ID missing")
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupId)
	if err != nil {
		resp["err"] = "unable to proceed"
		handler.logger.Error("err converting forumID(string) to forumID(int)", zap.Any("forum_id", groupId))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}

	success, err := handler.db.RequestGroupMembership(r.Context(), groupID, userInfo.Id)
	if err != nil || !success {
		resp["err"] = "unable to request membership"
		handler.logger.Error("err requesting membership", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}

	resp["message"] = "membership request sent"
	apiResponse(w, GetSuccessResponse(resp, 30), http.StatusOK)
}
