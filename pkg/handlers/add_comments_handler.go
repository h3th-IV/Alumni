package handlers

import (
	"net/http"
	"strconv"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

type commentHandler struct {
	logger *zap.Logger
	db     mysql.Database
}

func NewCommentHandler(logger *zap.Logger, db mysql.Database) *commentHandler {
	return &commentHandler{
		logger: logger,
		db:     db,
	}
}

func (ch *commentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	make_comment := map[string]interface{}{}
	userInfo, err := utils.AuthenticateUser(r.Context(), ch.logger, ch.db)
	if err != nil {
		make_comment["err"] = "please sign in to access this page"
		ch.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(make_comment["err"], 30, nil), http.StatusUnauthorized)
		return
	}

	comment := r.FormValue("comment")
	forum_id := r.URL.Query().Get("forum_id")
	if comment == "" {
		make_comment["err"] = "comment cannot be empty"
		ch.logger.Error("err, comment is empty")
		apiResponse(w, GetErrorResponseBytes(make_comment, 30, nil), http.StatusBadRequest)
		return
	}

	forumID, err := strconv.Atoi(r.FormValue(forum_id))
	if err != nil {
		make_comment["err"] = "unable to process request"
		ch.logger.Error("err processing request", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(make_comment, 30, nil), http.StatusBadRequest)
		return
	}

	success, err := ch.db.AddComment(r.Context(), userInfo.Id, forumID, comment)
	if err != nil || !success {
		make_comment["err"] = "failed to comment"
		ch.logger.Error("err making comments", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(make_comment, 30, nil), http.StatusBadRequest)
		return
	}

	make_comment["message"] = "comment added succefully"
	apiResponse(w, GetSuccessResponse(make_comment, 30), http.StatusOK)
}
