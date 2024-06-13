package repository

import "time"

type OTPRepository interface {
	SaveOTP(phoneNumber string, otp string) error
	GetOTP(phoneNumber string) (string, error)
	GetOTPTimestamp(phoneNumber string) (time.Time, error)
}
