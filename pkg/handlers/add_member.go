package handlers

import (
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

func (agh *addGroupMemberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	agh_resp := map[string]interface{}{}
	userInfo, err := utils.AuthenticateUser(r.Context(), agh.logger, agh.db)
	if err != nil {
		agh_resp["err"] = "please sign in to access this page"
		agh.logger.Error("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(agh_resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}
	groupID, err := strconv.Atoi(r.FormValue("group_id"))
	if err != nil {
		agh_resp["err"] = "unable to process request"
		agh.logger.Error("err ", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(agh_resp, 30, nil), http.StatusUnauthorized)
		return
	}

	newUser, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		agh_resp["err"] = "unable to process request"
		agh.logger.Error("err ", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(agh_resp, 30, nil), http.StatusUnauthorized)
		return
	}

	admin, err := agh.db.GetGroupCreator(r.Context(), groupID)
	if err != nil {
		agh_resp["err"] = "unable to proceed"
		agh.logger.Error("err fetching group admin")
		apiResponse(w, GetErrorResponseBytes(agh_resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}
	if admin.Id != userInfo.Id {
		log.Printf("%d, %d", admin.Id, userInfo.Id)
		agh_resp["err"] = "you are not allowed to add members to this group"
		agh.logger.Warn("admin, user IDs do not match")
		apiResponse(w, GetErrorResponseBytes(agh_resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	success, err := agh.db.AddGroupMember(r.Context(), groupID, newUser)
	if err != nil || !success {
		agh_resp["err"] = "unable to add user to group"
		agh.logger.Error("admin, user IDs do not match")
		apiResponse(w, GetErrorResponseBytes(agh_resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}
	agh_resp["message"] = "user added to group successfully"
	apiResponse(w, GetSuccessResponse(agh_resp, 30), http.StatusOK)
}
