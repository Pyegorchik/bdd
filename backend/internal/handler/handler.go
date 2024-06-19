package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Pyegorchik/bdd/backend/internal/config"
	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/Pyegorchik/bdd/backend/internal/service"
	"github.com/Pyegorchik/bdd/backend/pkg/logger"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
)

type Handler interface {
	Init() http.Handler
}

type HandlerFuncWithUser func(http.ResponseWriter, *domain.UserWithTokenNumber, *http.Request)

type handler struct {
	cfg               *config.HandlerConfig
	service           service.Service
	validationFormats strfmt.Registry
	logging           logger.Logger
}

func NewHandler(cfg *config.HandlerConfig, service service.Service, logging logger.Logger) Handler {
	return &handler{
		cfg:               cfg,
		service:           service,
		validationFormats: strfmt.NewFormats(),
		logging:           logging,
	}
}

const handlerIDPattern = "{id:[0-9]+}"

func (h *handler) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", h.cfg.SwaggerHost)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(200)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handler) Init() http.Handler {
	router := mux.NewRouter()

	log.Print(http.Dir("../static/"))
	fs := http.FileServer(http.Dir("../static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	authRouter := router.PathPrefix("/g1/auth").Subrouter()
	authRouter.Handle("/refresh", h.CookieRefreshAuthMiddleware((HandlerFuncWithUser(h.RefreshAuth))))
	authRouter.Handle("/logout", h.CookieAuthMiddleware((HandlerFuncWithUser(h.Logout))))
	authRouter.Handle("/full_logout", h.CookieAuthMiddleware((HandlerFuncWithUser(h.FullLogout))))
	authRouter.HandleFunc("/message", h.AuthMessage)
	authRouter.HandleFunc("/by_signature", h.AuthByMessage)

	rndRouter := router.PathPrefix("/rnd").Subrouter()
	rndRouter.HandleFunc("", h.Rnd)

	dialogsRounter := router.PathPrefix("/g1/dialogs").Subrouter()
	dialogsRounter.Handle("", h.CookieAuthMiddleware((HandlerFuncWithUser(h.GetDialogs))))
	dialogsRounter.Handle("/message", h.CookieAuthMiddleware((HandlerFuncWithUser(h.SendMessage))))
	dialogsRounter.Handle(fmt.Sprintf("/%s/messages", handlerIDPattern), h.CookieAuthMiddleware((HandlerFuncWithUser(h.GetMessages))))

	router.Use(h.corsMiddleware)
	return router
}
