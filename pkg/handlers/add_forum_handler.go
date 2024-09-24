package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/model"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

var _ http.Handler = &forumStruct{}

type forumStruct struct {
	logger *zap.Logger
	Db     mysql.Database
}

func NewForumStruct(logger *zap.Logger, Db mysql.Database) *forumStruct {
	return &forumStruct{
		logger: logger,
		Db:     Db,
	}
}

func (fs *forumStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	new_forum_response := map[string]interface{}{}
	userInfo, err := utils.AuthenticateUser(r.Context(), fs.logger, fs.Db)
	if err != nil {
		new_forum_response["err"] = "please sign in to access this page"
		fs.logger.Debug("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(new_forum_response["err"], 30, nil), http.StatusUnauthorized)
		return
	}
	var (
		author = userInfo.Email
		forum  *model.ForumPost
	)

	if err := json.NewDecoder(r.Body).Decode(&forum); err != nil {
		new_forum_response["err"] = "unable to process request"
		fs.logger.Error("err decoding JSON object", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(new_forum_response, loginTTL, nil), http.StatusNotFound)
		return
	}

	title := forum.Title
	description := forum.Description
	if title == "" || description == "" {
		fs.logger.Error("empty title or description")
		new_forum_response["error"] = "title or description is empty"
		apiResponse(w, GetErrorResponseBytes(new_forum_response, 30, nil), http.StatusBadRequest)
		return
	}

	if len(title) < 5 || len(description) < 50 {
		new_forum_response["error"] = "title or description is too short"
		fs.logger.Error("invalid title or description length")
		apiResponse(w, GetErrorResponseBytes(new_forum_response["err"], 30, nil), http.StatusBadRequest)
		return
	}

	if len(description) > 700 {
		new_forum_response["err"] = "max length of description exceeded"
		fs.logger.Error("max length of description exceeded")
		apiResponse(w, GetErrorResponseBytes(new_forum_response["err"], 30, nil), http.StatusBadRequest)
		return
	}

	slug := strings.Split(title, " ")
	_slug := strings.Join(slug, "")
	add_new_forum_post, err := fs.Db.AddNewForumPost(r.Context(), title, description, author, _slug, time.Now(), time.Now())
	if err != nil {
		fs.logger.Error("err creating new forum Post", zap.Error(err))
		new_forum_response["error"] = err.Error()
		apiResponse(w, GetErrorResponseBytes(new_forum_response, 30, err), http.StatusInternalServerError)
		return
	}
	if add_new_forum_post {
		new_forum_response["title"] = title
		new_forum_response["author"] = author
		new_forum_response["message"] = "forum post added successfully"
		apiResponse(w, GetSuccessResponse(new_forum_response, 30), http.StatusOK)
	}
}
