package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"dockerhub-pulllimits-exporter/internal/dockerhub"
	prom "dockerhub-pulllimits-exporter/internal/prometheus"
)

var (
	Opts = struct {
		Port int `long:"port" env:"DOCKERHUB_EXPORTER_PORT" required:"false" default:"8881"`

		DockerHubUsername string `long:"dockerhub-username" env:"DOCKERHUB_USERNAME" required:"false"`
		DockerHubPassword string `long:"dockerhub-password" env:"DOCKERHUB_PASSWORD" required:"false"`
	}{}

	checkImageLabel = "ratelimitpreview/test"
	checkImageTag   = "latest"

	retriever dockerhub.LimitsRetriever
)

func main() {
	log.Printf("dockerhub-pulllimits-exporter starting")
	defer func() {
		log.Printf("dockerhub-pulllimits-exporter completed")
	}()

	_, err := flags.Parse(&Opts)
	if err != nil {
		fmt.Println("cannot parse flags: ", err)
		os.Exit(1)
	}

	authCreds := dockerhub.AuthCredentials{}

	if Opts.DockerHubUsername != "" && Opts.DockerHubPassword != "" {
		log.Printf("Using provided '%s' credentials", Opts.DockerHubUsername)

		authCreds.Login = Opts.DockerHubUsername
		authCreds.PasswordToken = Opts.DockerHubPassword
	} else {
		log.Println("Using anonymous requests")
	}

	retriever, err = dockerhub.NewLimitsRetriever(checkImageLabel, checkImageTag, authCreds)
	if err != nil {
		log.Panicf("cannot initialize limit retriever: %s", err.Error())
	}

	reg, err := prom.RegisterCollectors(&retriever)
	if err != nil {
		log.Panicf("cannot register collectors: %s", err.Error())
	}

	log.Printf("exporter started on :%d port", Opts.Port)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", Opts.Port), nil); err != nil {
		log.Panicf("cannot start http server: %s", err.Error())
	}
}
