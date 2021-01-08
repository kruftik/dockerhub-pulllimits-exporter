package dockerhub

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	hc = http.Client{}
)

func (l *LimitsRetriever) getLimits(token string) (LimitsList, error) {
	sURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", l.image.Label, l.image.Tag)

	req, err := http.NewRequest("HEAD", sURL, nil)
	if err != nil {
		return LimitsList{}, fmt.Errorf("cannot init HEAD request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := hc.Do(req)
	if err != nil {
		return LimitsList{}, fmt.Errorf("cannot complete HEAD request: %s", err)
	}

	if resp.StatusCode != 200 {
		return LimitsList{}, fmt.Errorf("limit parameters request, respose status code = %d", resp.StatusCode)
	}

	limitSrc := resp.Header.Get("RateLimit-Limit")
	if limitSrc == "" {
		return LimitsList{}, fmt.Errorf("cannot obtain total limit info")
	}

	limits := LimitsList{}

	limit, err := ParseLimit(limitSrc)
	if err != nil {
		return LimitsList{}, fmt.Errorf("cannot parse total limit info: %w", err)
	}

	limits.Total = limit

	limitSrc = resp.Header.Get("RateLimit-Remaining")
	if limitSrc == "" {
		return LimitsList{}, fmt.Errorf("cannot obtain remaining limit info")
	}

	limit, err = ParseLimit(limitSrc)
	if err != nil {
		return LimitsList{}, fmt.Errorf("cannot parse remaining limit info: %w", err)
	}

	limits.Remaining = limit

	log.Printf("limit - %d pulls / %d second(s), %d remained", limits.Total.Limit, limits.Total.RefreshDuration/time.Second, limits.Remaining.Limit)

	return limits, nil
}

func (l *LimitsRetriever) GetLimits() (LimitsList, error) {
	l.m.Lock()
	defer func() {
		l.m.Unlock()
	}()

	token, err := l.AuthProvider.Get(l.image.Label)
	if err != nil {
		return LimitsList{}, fmt.Errorf("cannot get token: %s", err)
	}

	return l.getLimits(token)
}

func NewLimitsRetriever(imageLabel, imageTag string, creds AuthCredentials) (LimitsRetriever, error) {
	authProvider, err := NewAuthProvider(creds)
	if err != nil {
		return LimitsRetriever{}, fmt.Errorf("cannot initialize AuthProvider: %w", err)
	}

	return LimitsRetriever{
		image: struct {
			Label string
			Tag   string
		}{Label: imageLabel, Tag: imageTag},
		AuthProvider: authProvider,
	}, nil
}
