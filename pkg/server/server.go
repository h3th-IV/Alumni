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

	AddForumHandler        http.Handler // add forum post
	AllForumHandler        http.Handler // get all posts
	SingleForumHandler     http.Handler // get one post
	ChatHandler            http.Handler // chat a user
	CreateGroup            http.Handler
	AddUserToGroup         http.Handler
	SendGroupMessage       http.Handler
	GetChatHistory         http.Handler
	SendConnectionRequest  http.Handler
	AcceptHandler          http.Handler
	RequestMembership      http.Handler
	UpdateGroupMembership  http.Handler
	CheckPendingMembership http.Handler

	CommentHandler http.Handler //make comments

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
	middleWareChain := alice.New(utils.RequestLogger, cors.Handler)
	authRoute := alice.New(middleware.AuthRoute)
	//authed routes
	router.Handle("/users/profile", authRoute.ThenFunc(server.ProfileHandler.ServeHTTP)).Methods(http.MethodGet)
	router.Handle("/users/chat", authRoute.ThenFunc(server.ChatHandler.ServeHTTP)).Methods(http.MethodPost)
	router.Handle("/users/chat-history", authRoute.ThenFunc(server.GetChatHistory.ServeHTTP)).Methods(http.MethodGet)
	router.Handle("/forums/create/post", authRoute.ThenFunc(server.AddForumHandler.ServeHTTP)).Methods(http.MethodPost)
	router.Handle("/forums/comment", authRoute.ThenFunc(server.CommentHandler.ServeHTTP)).Methods(http.MethodPost)
	router.Handle("/groups/create", authRoute.ThenFunc(server.CreateGroup.ServeHTTP)).Methods(http.MethodPost)
	router.Handle("/groups/add-member", authRoute.ThenFunc(server.AddUserToGroup.ServeHTTP)).Methods(http.MethodPost)
	router.Handle("/groups/send-message", authRoute.ThenFunc(server.SendGroupMessage.ServeHTTP)).Methods(http.MethodPost)
	router.Handle("/groups/request-membership", authRoute.ThenFunc(server.RequestMembership.ServeHTTP)).Methods(http.MethodPost)
	router.Handle("/groups/update-membership", authRoute.ThenFunc(server.UpdateGroupMembership.ServeHTTP)).Methods(http.MethodPatch)
	router.Handle("/groups/check-pending-request", authRoute.ThenFunc(server.CheckPendingMembership.ServeHTTP)).Methods(http.MethodGet)
	router.Handle("/connections/send-new", authRoute.ThenFunc(server.SendConnectionRequest.ServeHTTP)).Methods(http.MethodPost)
	router.Handle("/connections/accept", authRoute.ThenFunc(server.AcceptHandler.ServeHTTP)).Methods(http.MethodPost)

	//no auth routes
	router.Handle("/forums", server.AllForumHandler).Methods(http.MethodGet)
	router.Handle("/forums/post/{slug}", server.SingleForumHandler).Methods(http.MethodGet)
	router.Handle("/register", server.RegisterHandler).Methods(http.MethodPost)
	router.Handle("/login", server.LoginHandler).Methods(http.MethodPost)
	router.Handle("/", server.HomeHandler)
	// cors.Handler(router) jim directive
	router.Use(middleWareChain.Then) //request logging will be handled here
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
