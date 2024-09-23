package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var _ http.Handler = &registerHandler{}

var registerTTL = 60

type registerHandler struct {
	logger      *zap.Logger
	mysqlclient mysql.Database
}

func NewRegisterHandler(logger *zap.Logger, mysqlclient mysql.Database) *registerHandler {
	return &registerHandler{
		logger:      logger,
		mysqlclient: mysqlclient,
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return password, err
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func validateEmail(email string) (bool, error) {
	if len(email) > 50 {
		return false, fmt.Errorf("email exceeds required length")
	}
	parts := strings.Split(strings.ToLower(email), "@")
	if len(parts) != 2 {
		return false, fmt.Errorf("email must contain @ symbol")
	}
	local, domain := parts[0], parts[1]
	if len(local) == 0 ||
		len(domain) == 0 {
		return false, fmt.Errorf("local or domain cannot be empty")
	}
	prev_char := rune(0)
	for _, char := range local {
		if strings.ContainsRune("!#$%&'*+-/=?^_`{|}~.", char) {
			if char == prev_char && char != '-' {
				return false, fmt.Errorf("cannot contain special chars before domain")
			}
		}
		prev_char = char
	}
	if strings.ContainsAny(email, " ") {
		return false, fmt.Errorf("email cannot contain spaces")
	}
	if len(local) > 64 || len(domain) > 255 {
		return false, fmt.Errorf("local part or domain part length exceeds the limit in the email")
	}
	return true, nil
}

func (handler *registerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		RegisterUser *model.User
		dataresp     = map[string]interface{}{}
	)

	if err := json.NewDecoder(r.Body).Decode(&RegisterUser); err != nil {
		dataresp["err"] = "unable to process request"
		handler.logger.Error("err decoding JSON object", zap.Error(err))
		apiResponse(w, GetErrorResponseBytes(dataresp, loginTTL, nil), http.StatusNotFound)
		return
	}
	// there should be a frontend validation for all fields
	// the backend would assist to catch empty fields if the
	// frontend validation is compromised.
	if RegisterUser.Username == "" || RegisterUser.Password == "" || RegisterUser.Degree == "" || RegisterUser.Phone == "" || RegisterUser.Email == "" {
		handler.logger.Error("some fields are empty")
		dataresp["err"] = "some fields are empty"
		apiResponse(w, GetSuccessResponse(dataresp, registerTTL), http.StatusBadRequest)
		return
	}

	// ensure the password is greater than 7 values
	// also ensure that it has special characters
	specialchars := strings.ContainsAny(RegisterUser.Password, "$ % @ !")
	passwdcount := len(RegisterUser.Password)
	if !specialchars {
		handler.logger.Error("password must contain special characters")
		dataresp["err"] = "password must contain special characters"
		apiResponse(w, GetSuccessResponse(dataresp, registerTTL), http.StatusBadRequest)
		return
	}
	if passwdcount <= 7 {
		handler.logger.Error("password must contain at least 8 characters")
		dataresp["err"] = "password must contain at least 8 characters"
		apiResponse(w, GetSuccessResponse(dataresp, registerTTL), http.StatusBadRequest)
		return
	}

	// we need to hash the password to avoid security issues
	// and then re-hash it when the user wants to login
	hashed_password, err := hashPassword(RegisterUser.Password)
	if err != nil {
		handler.logger.Error("cannot hash password", zap.String("hashed password error", err.Error()))
		fmt.Printf("cannot hash password(%s)", RegisterUser.Password)
		return
	}
	newsessionkey := createSessionKey(RegisterUser.Email, time.Now())
	sanitize_email, err := validateEmail(RegisterUser.Email)
	if err != nil || !sanitize_email {
		handler.logger.Error("email was malformed!", zap.Error(err))
		return
	}
	createUser, err := handler.mysqlclient.CreateUser(r.Context(), RegisterUser.Username, hashed_password, RegisterUser.Email, RegisterUser.Degree, RegisterUser.GradYear, RegisterUser.CurrentJob, RegisterUser.Phone, newsessionkey, "", RegisterUser.LinkedinProfile, RegisterUser.TwitterProfile)
	if err != nil || !createUser {
		dataresp["err"] = "cannot register user, try again"
		handler.logger.Error("could not create user", zap.Any("error", err))
		apiResponse(w, GetSuccessResponse(dataresp, registerTTL), http.StatusInternalServerError)
		return
	}
	dataresp["username"] = RegisterUser.Username
	dataresp["email"] = RegisterUser.Email
	dataresp["degree"] = RegisterUser.Degree
	dataresp["phone"] = RegisterUser.Phone
	dataresp["session_key"] = newsessionkey
	apiResponse(w, GetSuccessResponse(dataresp, registerTTL), http.StatusOK)
}
