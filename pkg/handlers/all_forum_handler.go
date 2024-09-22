package handlers

import (
	"log"
	"net/http"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
)

var _ http.Handler = &aforumStruct{}

type aforumStruct struct {
	Log *log.Logger
	Db  mysql.Database
}

func NewAForumStruct(log *log.Logger, Db mysql.Database) *aforumStruct {
	return &aforumStruct{
		Log: log,
		Db:  Db,
	}
}

func (fs *aforumStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	get_all_posts, err := fs.Db.GetAllForums(r.Context())
	if err != nil {
		fs.Log.Printf("'%s'\n", err)
		return
	}

	if get_all_posts != nil {
		apiResponse(w, GetSuccessResponse(get_all_posts, 30), http.StatusOK)
	} else {
		apiResponse(w, GetSuccessResponse([]struct{}{}, 30), http.StatusOK)
	}

}
