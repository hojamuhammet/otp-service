package service

import (
	"fmt"
	"math/rand"
	"otp-service/internal/config"
	repository "otp-service/internal/repository/interfaces"
	"otp-service/pkg/lib/logger"
	"strings"
	"time"

	"github.com/tarm/serial"
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

	err = s.sendSMS(phoneNumber, otp)
	if err != nil {
		s.logger.ErrorLogger.Error("Error sending OTP via SMS: %v", err)
		return err
	}

	return nil
}

func (s *OTPService) sendSMS(phoneNumber, otp string) error {
	c := &serial.Config{Name: "COM14", Baud: 115200}
	serialPort, err := serial.OpenPort(c)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}
	defer serialPort.Close()

	if _, err := serialPort.Write([]byte("AT\r")); err != nil {
		return fmt.Errorf("failed to initialize modem: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	if _, err := serialPort.Write([]byte("AT+CMGF=1\r")); err != nil {
		return fmt.Errorf("failed to set SMS text mode: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	if _, err := serialPort.Write([]byte(fmt.Sprintf("AT+CMGS=\"%s\"\r", phoneNumber))); err != nil {
		return fmt.Errorf("failed to specify recipient number: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	message := fmt.Sprintf("Your OTP is: %s", otp)
	if _, err := serialPort.Write([]byte(message + "\x1A")); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	buf := make([]byte, 128)
	n, err := serialPort.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response := string(buf[:n])
	if !strings.Contains(response, "+CMGS:") {
		return fmt.Errorf("failed to send message, response: %s", response)
	}

	s.logger.InfoLogger.Info(fmt.Sprintf("Message 'Your OTP is: %s' sent to %s, status: %s\n", otp, phoneNumber, response))
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
