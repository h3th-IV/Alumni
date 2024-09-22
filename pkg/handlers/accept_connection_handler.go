package handlers

import (
	"net/http"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
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
	userInfo, err := utils.AuthenticateUser(r.Context(), handler.logger, handler.db)
	if err != nil {
		resp["err"] = "please sign in to access this page"
		handler.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(resp["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		resp["err"] = "from_user_id is required"
		handler.logger.Error("from_user_id missing")
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
		resp["err"] = "unable to create connection"
		handler.logger.Error("failed to create connection", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(resp, 30, nil), http.StatusInternalServerError)
		return
	}

	resp["message"] = "connection accepted"
	apiResponse(w, GetSuccessResponse(resp, 30), http.StatusOK)
}
