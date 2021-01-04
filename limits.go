package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseLimit(limitDescr string) (DockerHubLimit, error){
	//	100;w=21600
	args := strings.Split(limitDescr, ";")
	if len(args) != 2 {
		return DockerHubLimit{}, fmt.Errorf("incorrect limit description: %s", limitDescr)
	}

	limit, err := strconv.Atoi(args[0])
	if err != nil {
		return DockerHubLimit{}, fmt.Errorf("incorrect limit number: %s", args[0])
	}

	durationArgs := strings.Split(args[1], "=")
	if len(args) != 2 {
		return DockerHubLimit{}, fmt.Errorf("incorrect limit duration: %s", args[1])
	}

	dur, err := strconv.Atoi(durationArgs[1])
	if err != nil {
		return DockerHubLimit{}, fmt.Errorf("incorrect duration number: %s", durationArgs[1])
	}

	return DockerHubLimit{
		Limit: limit,
		RefreshDuration: time.Duration(dur) * time.Second,
	}, nil
}
