package dockerhub

import (
	"sync"
	"time"
)

type authProviderT struct {
	m sync.RWMutex

	Credentials AuthCredentials

	JWTTokenInfo jwtToken
}

type jwtToken struct {
	Token     string        `json:"token"`
	ExpiresIn time.Duration `json:"expires_in"`
	IssuedAt  time.Time     `json:"issued_at"`

	ExpiresAt time.Time
}

type jwtTokenGetterFn func(string) (string, error)

type AuthCredentials struct {
	Login         string
	PasswordToken string
}

func (ac AuthCredentials) isValidCredentials() bool {
	return ac.Login != "" && ac.PasswordToken != ""
}

func (ac AuthCredentials) isBlankCredentials() bool {
	return ac.Login == "" && ac.PasswordToken == ""
}
