package registry

import (
	"net/http"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
)

func NewRepository(registryURL, repo string, trans http.RoundTripper) (distribution.Repository, error) {
	r, err := reference.WithName(repo)
	if err != nil {
		return nil, err
	}
	return client.NewRepository(r, registryURL, trans)
}
