package service

import (
	"fmt"
	"math/rand"
	"otp-service/internal/config"
	repository "otp-service/internal/repository/interfaces"
	"otp-service/pkg/lib/logger"
	"time"
)

type OTPService struct {
	repository repository.OTPRepository
	logger     *logger.Loggers
	cfg        *config.Config
}

func NewOTPService(repo repository.OTPRepository, logger *logger.Loggers, cfg *config.Config) OTPService {
	return OTPService{repository: repo, logger: logger, cfg: cfg}
}

func GenerateOTP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := fmt.Sprintf("%06d", r.Intn(1000000))
	return otp
}

func (s *OTPService) SaveAndSendOTP(phoneNumber string) error {
	otp := GenerateOTP()
	err := s.repository.SaveOTP(phoneNumber, otp)
	if err != nil {
		s.logger.ErrorLogger.Error("Error saving OTP to repository: %v", err)
		return err
	}

	return nil
}

func (s *OTPService) ValidateOTP(phoneNumber string, otp string) error {
	storedOTP, err := s.repository.GetOTP(phoneNumber)
	if err != nil {
		if err.Error() == "redis: nil" {
			return fmt.Errorf("OTP not found or expired")
		}
		s.logger.ErrorLogger.Error("Error retrieving OTP from repository: %v", err)
		return err
	}

	if storedOTP != otp {
		return fmt.Errorf("OTP does not match")
	}

	return nil
}
