package main

import (
	"intcli/pkg/azula/delivery/cli"
	"intcli/pkg/azula/repository/docker"
	"intcli/pkg/azula/repository/docker/auth"
	"intcli/pkg/azula/usecase"
	"net/url"
	"os"
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
