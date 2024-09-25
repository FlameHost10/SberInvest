package main

import (
	"AIChallenge/internal/config"
	"AIChallenge/internal/httpServer"
	"AIChallenge/internal/lib/handlers/slogpretty"
	"AIChallenge/internal/storage"
	"AIChallenge/internal/usecase/news"
	"fmt"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application", slog.Any("config", cfg))

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)

	newsRepo, err := storage.NewNewsRepository(connectionString, log)
	if err != nil {
		log.Warn("failed to create repoNews", slog.String("error", err.Error()))
		return
	}

	newsUsecase, err := usecase.NewNewsUseCase(log, newsRepo, 10, 20*time.Minute)
	if err != nil {
		log.Warn("failed to create newsUsecase", slog.String("error", err.Error()))
		return
	}

	httpHandler := httpServer.NewHTTPHandler(newsUsecase)

	router := mux.NewRouter()
	httpHandler.RegisterRoutes(router)

	go newsUsecase.Start()
	defer newsUsecase.Stop()

	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", router)
	log.Info("Started Server")

	return
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
