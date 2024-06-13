package routers

import (
	"net/http"
	"otp-service/internal/delivery/handlers"
	"otp-service/internal/service"
	"otp-service/pkg/lib/logger"

	"github.com/go-chi/chi/v5"
)

func SetupOTPRoutes(otpService service.OTPService, logger *logger.Loggers) http.Handler {
	otpRouter := chi.NewRouter()
	otpHandler := handlers.NewOTPHandler(otpService)

	otpRouter.Post("/sendOTP", otpHandler.GenerateAndSaveOTPHandler)
	otpRouter.Post("/validateOTP", otpHandler.ValidateOTPHandler)

	return otpRouter
}
