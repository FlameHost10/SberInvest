package main

import (
	"AIChallengeNewsAPI/internal/config"
	"AIChallengeNewsAPI/internal/httpServer"
	"AIChallengeNewsAPI/internal/logger"
	"AIChallengeNewsAPI/internal/repository"
	newsUsecase "AIChallengeNewsAPI/internal/usecase/news"
	"fmt"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

func main() {

	cfg := config.MustLoad()

	log := logger.NewLogger(cfg.Env)

	log.Info("starting application", slog.Any("config", cfg))

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)

	repository, err := repository.NewRepository(connectionString, log)
	if err != nil {
		log.Warn("failed to create repository", slog.String("error", err.Error()))
		return
	}

	newsUsecase, err := newsUsecase.NewNewsUseCase(log, repository, 10, 10*time.Minute)
	if err != nil {
		log.Warn("failed to create newsUsecase", slog.String("error", err.Error()))
		return
	}
	go newsUsecase.Start()
	defer newsUsecase.Stop()

	httpHandler := httpServer.NewHTTPHandler(newsUsecase)
	router := mux.NewRouter()
	httpHandler.RegisterRoutes(router)

	log.Info(fmt.Sprintf("Server is running at https://%s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port))
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port), router)
	if err != nil {
		log.Warn("failed to start server", slog.String("error", err.Error()))
		return
	}
	return
}
