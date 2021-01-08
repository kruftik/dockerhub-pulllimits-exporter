package dockerhub

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (t *authProviderT) setAuth(req *http.Request) {
	if t.Credentials.Login != "" && t.Credentials.PasswordToken != "" {
		req.SetBasicAuth(t.Credentials.Login, t.Credentials.PasswordToken)
	}
}

func (t *authProviderT) refresh(image string) error {
	sURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", image)

	req, err := http.NewRequest("GET", sURL, nil)
	if err != nil {
		return fmt.Errorf("cannot init request for JWT token: %w", err)
	}

	t.setAuth(req)

	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("cannot obtain JWT token: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("JWT request, respose status code = %d", resp.StatusCode)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("cannot clous resp.Body: %s", err)
		}
	}()

	dec := json.NewDecoder(resp.Body)

	t.m.Lock()
	defer func() {
		t.m.Unlock()
	}()

	err = dec.Decode(&(t.JWTTokenInfo))
	if err != nil {
		return fmt.Errorf("cannot decode JWT token struct: %w", err)
	}

	t.JWTTokenInfo.ExpiresIn *= time.Second

	t.JWTTokenInfo.ExpiresAt = t.JWTTokenInfo.IssuedAt.Add(t.JWTTokenInfo.ExpiresIn)

	return nil
}

func (t *authProviderT) Get(image string) (string, error) {
	var (
		bNewTokenNeeded = false
	)

	t.m.RLock()

	if t.JWTTokenInfo.Token == "" {
		log.Println("There is no obtained token")

		bNewTokenNeeded = true
	} else if time.Now().After(t.JWTTokenInfo.ExpiresAt) {
		log.Println("Present token expired, refresh needed")

		bNewTokenNeeded = true
	}

	if bNewTokenNeeded {
		t.m.RUnlock()

		err := t.refresh(image)
		if err != nil {
			return "", fmt.Errorf("cannot refresh token: %w", err)
		}

		log.Printf("New token retrieved, expiration in %d seconds", t.JWTTokenInfo.ExpiresIn/time.Second)

		t.m.RLock()
	}

	token := t.JWTTokenInfo.Token

	defer func() {
		t.m.RUnlock()
	}()

	return token, nil
}

func NewAuthProvider(credentials AuthCredentials) (*authProviderT, error) {
	if !credentials.isValidCredentials() && !credentials.isBlankCredentials() {
		return nil, fmt.Errorf("invalid credentials provided")
	}

	provider := &authProviderT{
		Credentials: credentials,
	}

	return provider, nil
}
