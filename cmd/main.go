package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"otp-service/internal/config"
	"otp-service/internal/delivery/routers"
	"otp-service/internal/repository"
	"otp-service/internal/service"
	db "otp-service/pkg/database"
	"otp-service/pkg/lib/logger"
	"syscall"
)

func main() {
	cfg := config.LoadConfig()

	logger, err := logger.SetupLogger(cfg.Env)
	if err != nil {
		slog.Error("failed to set up logger: %v", err)
		os.Exit(1)
	}

	logger.InfoLogger.Info("Server is up and running")
	slog.Info("Server is up and running")

	database, err := db.InitDB(cfg)
	if err != nil {
		logger.ErrorLogger.Error("failed to initialize database: %v", err)
		os.Exit(1)
	}

	repo := repository.NewOTPRepository(database.GetClient(), logger)
	otpService := service.NewOTPService(repo, logger, cfg)

	r := routers.SetupOTPRoutes(otpService, logger)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		logger.InfoLogger.Info("Shutting down the server gracefully...")
		if err := database.Close(); err != nil {
			logger.ErrorLogger.Error("Error closing database:", err)
		}
		os.Exit(0)
	}()

	err = http.ListenAndServe(cfg.HTTPServer.Address, r)
	if err != nil {
		logger.ErrorLogger.Error("Server failed to start:", err)
	}
}
