package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"vk-test-assignment-mattermost-polls/internal/api"
	"vk-test-assignment-mattermost-polls/internal/repository"
	"vk-test-assignment-mattermost-polls/internal/service"
	"vk-test-assignment-mattermost-polls/pkg/config"
	"vk-test-assignment-mattermost-polls/pkg/logger"
)

// @title Mattermost Voting Bot API
// @version 1.0
// @description API для бота голосований в Mattermost, поддерживающее создание и управление голосованиями в чатах.
// @contact.name t.me/mpstrkv
// @host localhost:8080
// @BasePath /

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Setup(cfg.Logger)

	log.Info().
		Str("app_env", cfg.Server.AppEnv).
		Str("port", cfg.Server.Port).
		Msg("Starting application")

	repo, err := repository.NewTarantoolRepository(cfg.Tarantool)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize repository")
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing repository connection")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pollService := service.NewPollService(repo, cfg.Poll)

	pollService.StartPollWatcher(ctx)
	pollService.StartPollCleaner(ctx)

	handler := api.NewHandler(pollService, cfg.Mattermost)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(cfg.Server.RequestTimeout))

	handler.RegisterRoutes(router)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	go func() {
		log.Info().
			Str("address", server.Addr).
			Msg("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Info().Msg("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown error")
	}

	cancel()

	log.Info().Msg("Server stopped successfully")
}
