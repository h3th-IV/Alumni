package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/model"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

type AcceptConnectionRequestHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewAcceptConnectionRequestHandler(logger *zap.Logger, db mysql.Database) *AcceptConnectionRequestHandler {
	return &AcceptConnectionRequestHandler{
		logger: logger,
		db:     db,
	}
}

func (handler *AcceptConnectionRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	var (
		request *model.ConnectionRequestEmail
	)
	userInfo, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.db)
	if err != nil {
		resp["err"] = "please sign in to access this page"
		handler.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		resp["err"] = "unable to process request"
		handler.logger.Error("err decoding JSON object", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, loginTTL, nil), http.StatusNotFound)
		return
	}
	email := request.Email
	if email == "" {
		resp["err"] = "connection user is missing"
		handler.logger.Error("connecting email is not provided")
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

	existingConnection, err := handler.db.CheckIfConnected(r.Context(), userInfo.Id, recv.Id)
	if err != nil {
		resp["err"] = "failed to check existing connection"
		handler.logger.Error("failed to check connection", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}
	if existingConnection {
		resp["err"] = "you are already connected with this user"
		handler.logger.Info("connection already exists between users")
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusConflict)
		return
	}

	connectionRequest, err := handler.db.GetConnectionRequest(r.Context(), recv.Id, userInfo.Id)
	if err != nil || connectionRequest.Status != "pending" {
		resp["err"] = "connection request not found or already processed"
		handler.logger.Error("connection request error", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusBadRequest)
		return
	}

	_, err = handler.db.UpdateConnectionRequest(r.Context(), connectionRequest.Id, "accepted")
	if err != nil {
		resp["err"] = "unable to accept connection request"
		handler.logger.Error("failed to accept connection request", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}

	_, err = handler.db.CreateConnection(r.Context(), userInfo.Id, recv.Id)
	if err != nil {
		resp["err"] = "unable to add user to network"
		handler.logger.Error("failed to create connection", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}

	resp["message"] = "connection accepted"
	apiResponse(w, GetSuccessResponse(resp, 30), http.StatusOK)
}
