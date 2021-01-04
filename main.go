package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Opts = struct {
		Port int `long:"port" env:"DOCKERHUB_EXPORTER_PORT" required:"false" default:"8881"`

		DockerHubUsername string `long:"dockerhub-username" env:"DOCKERHUB_USERNAME" required:"false"`
		DockerHubPassword string `long:"dockerhub-password" env:"DOCKERHUB_PASSWORD" required:"false"`
	}{}

	hc = http.Client{}

	retriever = NewDockerHubRetriever(checkImageLabel, checkImageTag, tokenInfo.Get)
)

func main() {
	_, err := flags.Parse(&Opts)
	if err != nil {
		fmt.Println("cannot parse flags: ", err)
		os.Exit(1)
	}

	if Opts.DockerHubUsername != "" && Opts.DockerHubPassword != "" {
		log.Printf("Using provided '%s' credentials", Opts.DockerHubUsername)
	} else {
		log.Println("Using anonymous requests")
	}

	reg, err := RegisterCollectors()
	if err != nil {
		log.Panicf("cannot register collectors: %s", err.Error())
	}

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", Opts.Port), nil); err != nil {
		log.Panicf("cannot start http server: %s", err.Error())
	}
}
