package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	checkImageLabel = "ratelimitpreview/test"
	checkImageTag = "latest"
)

type DockerHubLimit struct {
	Limit int
	RefreshDuration time.Duration
}

type DockerHubLimits struct {
	Total DockerHubLimit
	Remaining DockerHubLimit
}

type DockerHubRetriever struct {
	m sync.Mutex

	image struct{
		Label string
		Tag string
	}

	tokenGetterFn func(string) (string, error)

	limits DockerHubLimits
}

func (l *DockerHubRetriever) getLimits(token string) (DockerHubLimits, error) {
	sURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", l.image.Label, l.image.Tag)

	req, err := http.NewRequest("HEAD", sURL, nil)
	if err != nil {
		return DockerHubLimits{}, fmt.Errorf("cannot init HEAD request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := hc.Do(req)
	if err != nil {
		return DockerHubLimits{}, fmt.Errorf("cannot complete HEAD request: %s", err)
	}

	if resp.StatusCode != 200 {
		return DockerHubLimits{}, fmt.Errorf("limit parameters request, respose status code = %d", resp.StatusCode)
	}

	limitSrc := resp.Header.Get("RateLimit-Limit")
	if limitSrc == "" {
		log.Panicln("cannot obtain current limit info")
	}

	limits := DockerHubLimits{}

	limit, err := ParseLimit(limitSrc)
	if err != nil {
		return DockerHubLimits{}, fmt.Errorf("cannot parse total limit info: %w", err)
	}

	limits.Total = limit

	limitSrc = resp.Header.Get("RateLimit-Remaining")
	if limitSrc == "" {
		log.Panicln("cannot obtain remaining limit info")
	}

	limit, err = ParseLimit(limitSrc)
	if err != nil {
		return DockerHubLimits{}, fmt.Errorf("cannot parse remaining limit info: %w", err)
	}

	limits.Remaining = limit

	log.Printf("DockerHub current limits: total - %d, window %d hours, remaining - %d, window %d hours", limits.Total.Limit, limits.Total.RefreshDuration / time.Hour, limits.Remaining.Limit, limits.Remaining.RefreshDuration / time.Hour)


	return limits, nil
}

func (l *DockerHubRetriever) GetLimits() (DockerHubLimits, error) {
	l.m.Lock()
	defer func() {
		l.m.Unlock()
	}()

	token, err := l.tokenGetterFn(checkImageLabel)
	if err != nil {
		return DockerHubLimits{}, fmt.Errorf("cannot get token: %s", err)
	}

	return l.getLimits(token)
}

func NewDockerHubRetriever(imageLabel, imageTag string, tokenGetterFn func(string) (string, error)) DockerHubRetriever {
	return DockerHubRetriever{
		image: struct {
			Label string
			Tag   string
		}{Label: imageLabel, Tag: imageTag},
		tokenGetterFn: tokenGetterFn,
	}
}

