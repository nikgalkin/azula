package registry

import (
	"net/http"

	"github.com/docker/distribution/registry/client"
)

func NewRegistry(url string, trans http.RoundTripper) (client.Registry, error) {
	return client.NewRegistry(url, trans)
}
