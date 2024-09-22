package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jim-nnamdi/jinx/pkg/middleware"
	"github.com/jim-nnamdi/jinx/pkg/utils"
	"github.com/justinas/alice"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

type GracefulShutdownServer struct {
	HTTPListenAddr  string
	RegisterHandler http.Handler // register
	LoginHandler    http.Handler // login
	ProfileHandler  http.Handler // profile
	HomeHandler     http.Handler

	AddForumHandler    http.Handler // add forum post
	AllForumHandler    http.Handler // get all posts
	SingleForumHandler http.Handler // get one post
	ChatHandler        http.Handler // chat a user

	httpServer     *http.Server
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
	IdleTimeout    time.Duration
	HandlerTimeout time.Duration
}

func (server *GracefulShutdownServer) getRouter() *mux.Router {
	router := mux.NewRouter()

	mux.CORSMethodMiddleware(router)
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	middleWareChain := alice.New(utils.RequestLogger, utils.RecoverPanic, cors.Handler)
	authRoute := alice.New(middleware.AuthRoute)
	router.Handle("/profile", authRoute.ThenFunc(server.ProfileHandler.ServeHTTP)).Methods(http.MethodGet)
	router.Handle("/chat", authRoute.ThenFunc(server.ChatHandler.ServeHTTP)).Methods(http.MethodPost)
	router.Handle("/add-forum-post", authRoute.ThenFunc(server.AddForumHandler.ServeHTTP)).Methods(http.MethodGet)
	router.Handle("/forums", server.AllForumHandler)
	router.Handle("/forum-post", server.SingleForumHandler)
	router.Handle("/register", server.RegisterHandler)
	router.Handle("/login", server.LoginHandler)
	router.Handle("/", server.HomeHandler)
	router.Use(middleWareChain.Then) //request logging will be handled here
	mux.CORSMethodMiddleware(router)
	router.SkipClean(true)
	return router
}

func (server *GracefulShutdownServer) Start() {
	router := server.getRouter()
	server.httpServer = &http.Server{
		Addr:         server.HTTPListenAddr,
		WriteTimeout: server.WriteTimeout,
		ReadTimeout:  server.ReadTimeout,
		IdleTimeout:  server.IdleTimeout,
		Handler:      router,
	}
	utils.Logger.Info(fmt.Sprintf("listening and serving on %s", server.HTTPListenAddr))
	if err := server.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.Logger.Fatal("server failed to start", zap.Error(err))
	}
}
