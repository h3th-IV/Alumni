package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/model"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

var _ http.Handler = &SendConnectionRequestHandler{}

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
	var (
		connectionRequest *model.ConnectionRequestEmail
	)
	resp := map[string]interface{}{}
	userInfo, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.db)
	if err != nil {
		resp["err"] = "please sign in to access this page"
		handler.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&connectionRequest); err != nil {
		resp["err"] = "unable to process request"
		handler.logger.Error("err decoding JSON object", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, loginTTL, nil), http.StatusNotFound)
		return
	}
	email := connectionRequest.Email
	if email == "" {
		resp["err"] = "recipient email is required"
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
	if recv.Id == userInfo.Id {
		resp["err"] = "you can't send a request to yourself"
		handler.logger.Info("user tried sending request to self", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusBadRequest)
		return
	}
	connected, err := handler.db.CheckIfConnected(r.Context(), userInfo.Id, recv.Id)
	if err != nil {
		resp["err"] = "unable to check for previous connection"
		handler.logger.Error("err checking if users are connected", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}
	if connected {
		resp["warning"] = "you are already connected to this user"
		handler.logger.Info("users already connetced")
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusConflict)
		return
	}

	existingRequest, err := handler.db.CheckPendingConnection(r.Context(), userInfo.Id, recv.Id)
	if err != nil {
		resp["err"] = "failed to check pending request"
		handler.logger.Error("failed to check pending request", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}
	if existingRequest {
		resp["err"] = "a connection request is already pending with this user"
		handler.logger.Info("pending connection request exists between users")
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusConflict)
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
