package main

import (
	"net/url"
	"os"

	"github.com/nikgalkin/azula/pkg/azula/delivery/cli"
	"github.com/nikgalkin/azula/pkg/azula/repository/docker"
	"github.com/nikgalkin/azula/pkg/azula/repository/docker/auth"
	"github.com/nikgalkin/azula/pkg/azula/usecase"
)

func main() {
	mgr, err := genRegistryInit()
	if err != nil {
		panic(err)
	}
	dr, err := mgr.New()
	if err != nil {
		panic(err)
	}
	cli.New(usecase.New(dr)).Execute()
}

func genRegistryInit() (*docker.RegistryInit, error) {
	cfg, err := auth.LoadDefaultConfig()
	if err != nil {
		return &docker.RegistryInit{}, err
	}

	registry := os.Getenv("MAN_REGISTRY")
	if len(registry) < 1 {
		registry = "http://127.0.0.1:5000"
	}

	u, err := url.Parse(registry)
	if err != nil {
		return &docker.RegistryInit{}, err
	}
	user, pass, err := cfg.GetRegistryCredentials(u.Host)
	if err != nil {
		return &docker.RegistryInit{}, err
	}
	return &docker.RegistryInit{
		Username: user,
		Password: pass,
		URL:      registry,
	}, nil
}
