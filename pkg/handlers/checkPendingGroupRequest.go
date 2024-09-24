package handlers

import (
	"net/http"
	"strconv"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

var _ http.Handler = &FetchPendingMembershipRequestsHandler{}

type FetchPendingMembershipRequestsHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewFetchPendingMembershipRequestsHandler(logger *zap.Logger, db mysql.Database) *FetchPendingMembershipRequestsHandler {
	return &FetchPendingMembershipRequestsHandler{
		logger: logger,
		db:     db,
	}
}

func (handler *FetchPendingMembershipRequestsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	_, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.db)
	if err != nil {
		resp["err"] = "please sign in to access this page"
		handler.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	groupId := r.URL.Query().Get("group_id")
	groupID, err := strconv.Atoi(groupId)
	if err != nil {
		resp["err"] = "invalid group ID"
		handler.logger.Error("error converting groupID to int", zap.String("group_id", groupId))
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusBadRequest)
		return
	}

	users, err := handler.db.CheckPendingMembershipRequest(r.Context(), groupID)
	if err != nil {
		resp["err"] = "unable to fetch pending requests"
		handler.logger.Error("error fetching pending requests", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}

	resp["users"] = users
	apiResponse(w, GetSuccessResponse(resp, 30), http.StatusOK)
}
