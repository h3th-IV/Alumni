package handlers

import (
	"net/http"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"go.uber.org/zap"
)

var _ http.Handler = &aforumStruct{}

type aforumStruct struct {
	logger *zap.Logger
	Db     mysql.Database
}

func NewAForumStruct(logger *zap.Logger, Db mysql.Database) *aforumStruct {
	return &aforumStruct{
		logger: logger,
		Db:     Db,
	}
}

func (fs *aforumStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	all_forum_response := map[string]interface{}{}
	get_all_posts, err := fs.Db.GetAllForums(r.Context())
	if err != nil {
		all_forum_response["err"] = "unable to fetch forum post"
		fs.logger.Error("err fetching forum post", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(all_forum_response, 30, nil), http.StatusInternalServerError)
		return
	}
	apiResponse(w, GetSuccessResponse(get_all_posts, 30), http.StatusOK)

}
