package handlers

import (
	"net/http"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

type createGroupHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewCreateGroupHandler(logger *zap.Logger, db mysql.Database) *createGroupHandler {
	return &createGroupHandler{
		logger: logger,
		db:     db,
	}
}

func (cgh *createGroupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cg_resp := map[string]interface{}{}
	userInfo, err := utils.AuthenticateUser(r.Context(), cgh.logger, cgh.db)
	if err != nil {
		cg_resp["err"] = "please sign in to access this page"
		cgh.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(cg_resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	groupName := r.FormValue("name")
	if groupName == "" {
		cg_resp["err"] = "group name cannot be empty"
		cgh.logger.Error("group name is empty")
		apiResponse(w, GetErrorResponseBytes(cg_resp, 30, nil), http.StatusBadRequest)
		return
	}

	groupID, err := cgh.db.CreateGroup(r.Context(), groupName, userInfo.Id)
	if err != nil {
		cg_resp["err"] = "unable to create group"
		cgh.logger.Error("err creating new group", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(cg_resp, 30, nil), http.StatusInternalServerError)
		return
	}

	success, err := cgh.db.AddGroupMember(r.Context(), int(groupID), userInfo.Id)
	if err != nil || !success {
		cg_resp["err"] = "unable to add group creator as a member"
		cgh.logger.Error("failed to add group creator as member", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(cg_resp, 30, nil), http.StatusInternalServerError)
		return
	}

	cg_resp["message"] = "group created successfully"
	apiResponse(w, GetSuccessResponse(cg_resp, 30), http.StatusCreated)
}
