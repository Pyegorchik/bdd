package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Pyegorchik/bdd/backend/internal/config"
	"github.com/Pyegorchik/bdd/backend/internal/handler"
	"github.com/Pyegorchik/bdd/backend/internal/repository"
	"github.com/Pyegorchik/bdd/backend/internal/repository/postgres"
	"github.com/Pyegorchik/bdd/backend/internal/server"
	"github.com/Pyegorchik/bdd/backend/internal/service"
	"github.com/Pyegorchik/bdd/backend/pkg/hash"
	"github.com/Pyegorchik/bdd/backend/pkg/jwtoken"
	"github.com/Pyegorchik/bdd/backend/pkg/logger"
)

func main() {
	logging, err := logger.NewLogger()
	if err != nil {
		log.Panic(err)
	}
	defer logging.Sync()

	var cfgPath string

	flag.StringVar(&cfgPath, "cfg", "", "")
	flag.Parse()

	cfg, err := config.Init(cfgPath)
	if err != nil {
		logging.Panic(err)
	}
	ctx := context.Background()

	pool, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		logging.Panic(err)
	}

	jwtokenManager := jwtoken.NewTokenManager(cfg.TokenManager.SigningKey)

	bddRepos, err := repository.NewRepository(cfg, pool)
	if err != nil {
		logging.Panic(err)
	}

	bddService, err := service.NewService(bddRepos, jwtokenManager, hash.NewHashManager(), cfg.Service, logging)
	if err != nil {
		logging.Panic(err)
	}

	router := handler.NewHandler(cfg.Handler, bddService, logging)

	srv := server.NewServer(cfg.Server, router.Init())

	go func() {
		if err = srv.ListenAndServe(); err != http.ErrServerClosed {
			logging.Panic(err)
		}
	}()
	logging.Infof("server listening on port %d \n", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	if err = srv.Shutdown(ctx); err != nil {
		logging.Panic(err)
	}

	bddService.Shutdown()
}