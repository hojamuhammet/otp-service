package repository

import (
	"fmt"
	"math/rand"
	"otp-service/pkg/lib/logger"
	"time"

	"github.com/go-redis/redis"
)

type OTPRepository struct {
	Client *redis.Client
	logger *logger.Loggers
}

func NewOTPRepository(client *redis.Client, logger *logger.Loggers) *OTPRepository {
	return &OTPRepository{
		Client: client,
		logger: logger,
	}
}

func GenerateOTP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := fmt.Sprintf("%06d", r.Intn(1000000))
	return otp
}

func (r *OTPRepository) SaveOTP(phoneNumber string, otp string) error {
	err := r.Client.Set(phoneNumber, otp, 5*time.Minute).Err()
	if err != nil {
		r.logger.ErrorLogger.Error("Error setting up values into redis: %v", err)
		return err
	}
	return nil
}

func (r *OTPRepository) GetOTP(phoneNumber string) (string, error) {
	otp, err := r.Client.Get(phoneNumber).Result()
	if err != nil {
		return "", err
	}
	return otp, nil
}

func (r *OTPRepository) GetOTPTimestamp(phoneNumber string) (time.Time, error) {
	timestamp, err := r.Client.Get(phoneNumber + ":timestamp").Result()
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
