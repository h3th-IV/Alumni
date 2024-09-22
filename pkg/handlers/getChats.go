package handlers

import (
	"net/http"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

type getUserChatsHistoryHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewGetUserChatsHistoryHandler(logger *zap.Logger, db mysql.Database) *getUserChatsHistoryHandler {
	return &getUserChatsHistoryHandler{
		logger: logger,
		db:     db,
	}
}

func (guc *getUserChatsHistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chat_resp := map[string]interface{}{}

	userInfo, err := utils.AuthenticateUser(r.Context(), guc.logger, guc.db)
	if err != nil {
		chat_resp["err"] = "please sign in to access this page"
		guc.logger.Error("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(chat_resp, 30, nil), http.StatusUnauthorized)
		return
	}

	recipientEmail := r.FormValue("recv_email")
	if recipientEmail == "" {
		chat_resp["err"] = "recipient email is required"
		guc.logger.Error("recipient email is missing")
		apiResponse(w, GetErrorResponseBytes(chat_resp, 30, nil), http.StatusBadRequest)
		return
	}

	recipientUser, err := guc.db.GetUserByEmail(r.Context(), recipientEmail)
	if err != nil {
		chat_resp["err"] = "recipient not found"
		guc.logger.Error("failed to fetch recipient", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(chat_resp, 30, err), http.StatusNotFound)
		return
	}

	chats, err := guc.db.FetchUserChats(r.Context(), userInfo.Id, recipientUser.Id)
	if err != nil {
		chat_resp["err"] = "failed to fetch chat history"
		guc.logger.Error("failed to fetch chats", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(chat_resp, 30, err), http.StatusInternalServerError)
		return
	}

	chat_resp["chats"] = chats
	apiResponse(w, GetSuccessResponse(chat_resp, 30), http.StatusOK)
}
