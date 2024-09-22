package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jim-nnamdi/jinx/pkg/database/mysql"
	"github.com/jim-nnamdi/jinx/pkg/model"
	"go.uber.org/zap"
)

var Logger, _ = zap.NewDevelopment()

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Logger.Info((fmt.Sprintf("%v - %v %v %v", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())))
		next.ServeHTTP(w, r)
	})
}

type mapKey string

const (
	UserIDKey mapKey = "user_id"
)

func AuthenticateUser(ctx context.Context, logger *zap.Logger, mysqlclient mysql.Database) (*model.User, error) {
	sessionKey, ok := ctx.Value(UserIDKey).(string)
	if !ok || sessionKey == "" {
		logger.Error("session key is missing")
		return nil, errors.New("please sign in to access this page")
	}

	user, err := mysqlclient.GetBySessionKey(ctx, sessionKey)
	if err != nil {
		logger.Error("user is not authorized", zap.Error(err))
		return nil, errors.New("please sign in to access this page")
	}

	return user, nil
}

func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				//if panic close connection
				w.Header().Set("Connection", "Close")
				//write internal server error
				ServerError(w, "Connection Closed inabruptly", fmt.Errorf("%v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func ServerError(w http.ResponseWriter, errMsg string, err error) {
	errTrace := fmt.Sprintf("%v\n%v", err.Error(), debug.Stack())
	Logger.Error(errTrace)
	http.Error(w, errMsg, http.StatusInternalServerError)
}

func GenerateToken(user *model.User, expiry time.Duration, issuer, secret string) (string, error) {
	//set token expiry time
	bestBefore := time.Now().Add(expiry)
	//set token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  user.Email,
		"user":   user.SessionKey,
		"exp":    bestBefore.Unix(),
		"issuer": issuer,
	})
	//generate jwt token str and sign with secret key
	JWToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return JWToken, nil
}

var (
	MYSTIC    string
	JWTISSUER string
)
