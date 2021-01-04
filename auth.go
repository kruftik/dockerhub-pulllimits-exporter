package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	tokenInfo = TokenInfoT{}
)

type TokenInfoT struct {
	m sync.RWMutex
	TokenInfo DockerHubJWTToken
}

type DockerHubJWTToken struct {
	Token string `json:"token"`
	ExpiresIn time.Duration `json:"expires_in"`
	IssuedAt time.Time `json:"issued_at"`

	ExpiresAt time.Time
}


func (t *TokenInfoT) refresh(image string) error {
	sURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", image)

	resp, err := hc.Get(sURL)
	if err != nil {
		return fmt.Errorf("cannot obtain JWT token: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("JWT request, respose status code = %d", resp.StatusCode)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("cannot clous resp.Body: %w", err)
		}
	}()

	dec := json.NewDecoder(resp.Body)

	t.m.Lock()
	defer func() {
		t.m.Unlock()
	}()

	err = dec.Decode(&(t.TokenInfo))
	if err != nil {
		return fmt.Errorf("cannot decode JWT token struct: %w", err)
	}

	t.TokenInfo.ExpiresIn *= time.Second

	t.TokenInfo.ExpiresAt = t.TokenInfo.IssuedAt.Add(t.TokenInfo.ExpiresIn)

	return nil
}

func (t *TokenInfoT) Get(image string) (string, error) {
	var (
		bNewTokenNeeded = false
	)

	t.m.RLock()

	if t.TokenInfo.Token == "" {
		log.Println("There is no obtained token")

		bNewTokenNeeded = true
	} else if time.Now().After(t.TokenInfo.ExpiresAt) {
		log.Println("Present token expired, refresh needed")

		bNewTokenNeeded = true
	}

	if bNewTokenNeeded {
		t.m.RUnlock()

		err := t.refresh(image)
		if err != nil {
			return "", fmt.Errorf("cannot refresh token: %w", err)
		}

		log.Printf("New token retrieved, expiration in %d seconds", t.TokenInfo.ExpiresIn / time.Second)

		t.m.RLock()
	}

	token := t.TokenInfo.Token

	defer func() {
		t.m.RUnlock()
	}()

	return token, nil
}

