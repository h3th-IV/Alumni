package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/model"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"go.uber.org/zap"
)

var _ http.Handler = &ichatStruct{}

type ichatStruct struct {
	logger *zap.Logger
	DB     mysql.Database
}

func NewChat(logger *zap.Logger, Db mysql.Database) *ichatStruct {
	return &ichatStruct{
		logger: logger,
		DB:     Db,
	}
}

func (cs *ichatStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		msg_resp = map[string]interface{}{}
		log      log.Logger
		chat     *model.Chat
	)
	current_user, err := utils.AuthenticateUser(r.Context(), cs.logger, cs.DB)
	if err != nil {
		msg_resp["error"] = "please sign in to access this page"
		msg_resp["usr"] = "failed to get user"
		msg_resp["db_error"] = "authentication failed: unable to get user data"
		cs.logger.Error("unauthorized user")
		apiResponse(w, GetErrorResponseBytes(msg_resp, 30, fmt.Errorf("'%s'", "please try authenticating again")), http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
		msg_resp["err"] = "unable to process request"
		cs.logger.Error("err decoding JSON object", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(msg_resp, loginTTL, nil), http.StatusNotFound)
		return
	}

	recipient := chat.Email
	message := chat.Message
	if recipient == "" || message == "" {
		msg_resp["error"] = "some fields are empty"
		cs.logger.Error("receiver and message are empty")
		apiResponse(w, GetErrorResponseBytes(msg_resp, 30, fmt.Errorf("'%s'", "please provide recipient or message body")), http.StatusBadRequest)
		return
	}
	recv_user, err := cs.DB.GetUserByEmail(r.Context(), recipient)
	if err != nil {
		log.Printf("'%s'\n", err)
		log.Printf("'%s'\n", "check if recipient exists")
		failed_retrieval := map[string]string{}
		failed_retrieval["error"] = "failed to get recipient"
		failed_retrieval["db_error"] = err.Error()
		apiResponse(w, GetErrorResponseBytes(failed_retrieval, 30, fmt.Errorf("'%s'", err.Error())), http.StatusInternalServerError)
		return
	}
	msg_valid := len(message)
	if msg_valid > 100 {
		log.Printf("'%s'\n", "max message threshold")
		log.Printf("'%s'\n", message)
		msg_resp["err"] = "max message threshold"
		apiResponse(w, GetErrorResponseBytes(msg_resp, 30, fmt.Errorf("'%s'", "error sending message")), http.StatusBadRequest)
		return
	}
	send_chat, err := cs.DB.SendMessage(r.Context(), current_user.Id, recv_user.Id, message, time.Now(), time.Now())
	if err != nil {
		log.Printf("'%s'\n", "could not send message to recipient")
		nilc_resp := map[string]string{}
		nilc_resp["error"] = "error sending message"
		nilc_resp["db_error"] = err.Error()
		apiResponse(w, GetErrorResponseBytes(nilc_resp, 30, err), http.StatusInternalServerError)
		return
	}
	if send_chat {
		chatresp := map[string]interface{}{}
		chatresp["sender"] = current_user.Username
		chatresp["receiver"] = recv_user.Username
		chatresp["message"] = message
		chatresp["created_at"] = time.Now()
		chatresp["updated_at"] = time.Now()
		apiResponse(w, GetSuccessResponse(chatresp, 30), http.StatusOK)
		return
	}
}
