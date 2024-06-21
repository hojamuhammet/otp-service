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
	serialPortName := s.cfg.SerialPort.Name
	baudRate := s.cfg.SerialPort.Baud
	c := &serial.Config{Name: serialPortName, Baud: baudRate}

	serialPort, err := serial.OpenPort(c)
	if err != nil {
		s.logger.ErrorLogger.Error(fmt.Sprintf("Failed to open serial port: %v", err))
		return err
	}
	defer serialPort.Close()

	if _, err := serialPort.Write([]byte("AT\r")); err != nil {
		s.logger.ErrorLogger.Error(fmt.Sprintf("Failed to initialize modem: %v", err))
		return err
	}
	s.logger.InfoLogger.Info("Modem initialized")
	time.Sleep(500 * time.Millisecond)

	if _, err := serialPort.Write([]byte("AT+CMGF=1\r")); err != nil {
		s.logger.ErrorLogger.Error(fmt.Sprintf("Failed to set SMS text mode: %v", err))
		return err
	}
	s.logger.InfoLogger.Info("SMS text mode set")
	time.Sleep(500 * time.Millisecond)

	if _, err := serialPort.Write([]byte("AT+CMGS=\"" + phoneNumber + "\"\r")); err != nil {
		s.logger.ErrorLogger.Error(fmt.Sprintf("Failed to specify recipient number: %v", err))
		return err
	}
	s.logger.InfoLogger.Info(fmt.Sprintf("Recipient number specified: %s", phoneNumber))
	time.Sleep(500 * time.Millisecond)

	message := "Your OTP is: " + otp
	if _, err := serialPort.Write([]byte(message + "\x1A")); err != nil {
		s.logger.ErrorLogger.Error(fmt.Sprintf("Failed to send message: %v", err))
		return err
	}
	s.logger.InfoLogger.Info("Message sent to modem")
	time.Sleep(500 * time.Millisecond)

	buf := make([]byte, 128)
	n, err := serialPort.Read(buf)
	if err != nil {
		s.logger.ErrorLogger.Error(fmt.Sprintf("Failed to read response: %v", err))
		return err
	}

	response := string(buf[:n])
	cleanedResponse := strings.ReplaceAll(response, "\r", "")
	cleanedResponse = strings.ReplaceAll(cleanedResponse, "\n", " ")

	if !strings.Contains(cleanedResponse, "+CMGS:") {
		s.logger.ErrorLogger.Error(fmt.Sprintf("Failed to send message, response: %s", cleanedResponse))
		return fmt.Errorf("failed to send message, response: %s", cleanedResponse)
	}

	s.logger.InfoLogger.Info(fmt.Sprintf("Message 'Your OTP is: %s' sent to %s, status: %s", otp, phoneNumber, cleanedResponse))
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
