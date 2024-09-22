package handlers

import "net/http"

var _ http.Handler = &homeHandler{}

type homeHandler struct{}

func NewHomeHandler() *homeHandler {
	return &homeHandler{}
}

func (handler *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	homedata := map[string]interface{}{}
	homedata["success"] = true
	apiResponse(w, GetSuccessResponse(homedata, 30), http.StatusOK)
}
