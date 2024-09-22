package handlers

import (
	"net/http"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

type SendConnectionRequestHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewSendConnectionRequestHandler(logger *zap.Logger, db mysql.Database) *SendConnectionRequestHandler {
	return &SendConnectionRequestHandler{
		logger: logger,
		db:     db,
	}
}

func (handler *SendConnectionRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	userInfo, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.db)
	if err != nil {
		resp["err"] = "please sign in to access this page"
		handler.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		resp["err"] = "recipint email is required"
		handler.logger.Error("email is missing")
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusBadRequest)
		return
	}
	recv, err := handler.db.GetUserByEmail(r.Context(), email)
	if err != nil {
		resp["err"] = "user not found"
		handler.logger.Error("err", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusBadRequest)
		return
	}

	suc, err := handler.db.CreateConnectionRequest(r.Context(), userInfo.Id, recv.Id)
	if err != nil {
		resp["err"] = "unable to send connection request"
		handler.logger.Error("failed to send connection request", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}
	if !suc {
		resp["err"] = "unable to send connection request"
		handler.logger.Error("failed to send connection request without error")
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}

	resp["message"] = "connection request sent"
	apiResponse(w, GetSuccessResponse(resp, 30), http.StatusOK)
}
