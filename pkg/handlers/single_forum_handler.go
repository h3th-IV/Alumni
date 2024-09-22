package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"go.uber.org/zap"
)

var _ http.Handler = &sforumStruct{}

type sforumStruct struct {
	logger *zap.Logger
	Db     mysql.Database
}

func NewSForumStruct(logger *zap.Logger, Db mysql.Database) *sforumStruct {
	return &sforumStruct{
		logger: logger,
		Db:     Db,
	}
}

func (fs *sforumStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		sfp = map[string]string{}
	)
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		sfp["err"] = "slug is empty"
		fs.logger.Error("slug is empty") //ask jim what is slug
		apiResponse(w, GetErrorResponseBytes(sfp, 30, nil), http.StatusBadRequest)
		return
	}

	get_single_forum_post, err := fs.Db.GetSingleForumPost(r.Context(), slug)
	if err != nil {
		fs.logger.Error("err getting forum post", zap.Error(err))
		sfp["error"] = "unable to get forum post"
		apiResponse(w, GetErrorResponseBytes(sfp, 30, err), http.StatusInternalServerError)
		return
	}
	comments, err := fs.Db.GetCommentsByForumID(r.Context(), get_single_forum_post.Id)
	if err != nil {
		sfp["err"] = "unable to post comments"
		fs.logger.Error("err fetching post comments", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(sfp, 30, nil), http.StatusNotFound)
		return
	}
	if get_single_forum_post != nil {
		forum_resp := map[string]interface{}{}
		forum_resp["id"] = get_single_forum_post.Id
		forum_resp["title"] = get_single_forum_post.Title
		forum_resp["description"] = get_single_forum_post.Description
		forum_resp["slug"] = get_single_forum_post.Slug
		forum_resp["created_at"] = get_single_forum_post.CreatedAt
		forum_resp["updated_at"] = get_single_forum_post.UpdatedAt
		forum_resp["comments"] = comments
		apiResponse(w, GetSuccessResponse(forum_resp, 30), http.StatusOK)
	} else {
		sfp["err"] = "no post data"
		apiResponse(w, GetSuccessResponse(sfp, 30), http.StatusNotFound)
	}

}
