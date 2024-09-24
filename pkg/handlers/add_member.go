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

var _ http.Handler = &addGroupMemberHandler{}

type addGroupMemberHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewAddGroupMemberHandler(logger *zap.Logger, db mysql.Database) *addGroupMemberHandler {
	return &addGroupMemberHandler{
		logger: logger,
		db:     db,
	}
}

func (handler *addGroupMemberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	var new_user_email struct {
		Email string `json:"email"`
	}
	userInfo, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.db)
	if err != nil {
		resp["err"] = "please sign in to access this page"
		handler.logger.Error("unauthorized user")
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

	if err := json.NewDecoder(r.Body).Decode(&new_user_email); err != nil {
		resp["err"] = "unable to process request"
		handler.logger.Error("err decoding JSON object", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, loginTTL, nil), http.StatusNotFound)
		return
	}

	email := new_user_email.Email
	newUser, err := handler.db.GetUserByEmail(r.Context(), email)
	if err != nil {
		resp["err"] = "user does not exist"
		handler.logger.Error("use does not exist", zap.Any("checkuser", err))
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusNotFound)
		return
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
		resp["err"] = "you are not allowed to add members to this group"
		handler.logger.Warn("admin, user IDs do not match")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	success, err := handler.db.AddGroupMember(r.Context(), groupID, newUser.Id)
	if err != nil || !success {
		resp["err"] = "unable to add user to group"
		handler.logger.Error("admin, user IDs do not match")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}
	resp["message"] = "user added to group successfully"
	apiResponse(w, GetSuccessResponse(resp, 30), http.StatusOK)
}
