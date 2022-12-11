package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/nikgalkin/azula/pkg/azula/repository/docker/registry"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/client"
	"github.com/opencontainers/go-digest"
)

type Registry struct {
	Transport http.RoundTripper
	Registry  client.Registry
	URL       string
}

type RegistryInit struct {
	Username string
	Password string
	URL      string
}

type Manager interface {
	ListReposLike(context.Context, string) ([]string, error)
	GetRepo(context.Context, string) (distribution.Repository, error)
	GetV2Descriptor(context.Context, string, string) (distribution.Descriptor, error)
}

func (init *RegistryInit) New() (Manager, error) {
	dr := Registry{}
	var err error
	dr.Transport, err = registry.NewTrans(init.Username, init.Password, init.URL)
	if err != nil {
		return nil, err
	}
	dr.Registry, err = registry.NewRegistry(init.URL, dr.Transport)
	if err != nil {
		return nil, err
	}
	dr.URL = init.URL
	return &dr, nil
}

func (r *Registry) ListReposLike(ctx context.Context, like string) ([]string, error) {
	step := 20
	entries := make([]string, step)
	var err error
	var n int
	for {
		n, err = r.Registry.Repositories(ctx, entries, "")
		if err == nil {
			entries = append(entries, make([]string, step)...)
			continue
		} else if err == io.EOF {
			break
		} else if err != nil {
			return []string{}, err
		}
	}
	if len(like) > 0 {
		n = 0
		for _, v := range entries {
			if strings.Contains(v, like) {
				entries[n] = v
				n += 1
			}
		}
	}

	return entries[:n], nil
}

func (r *Registry) GetRepo(ctx context.Context, name string) (distribution.Repository, error) {
	return registry.NewRepository(r.URL, name, r.Transport)
}

func (r *Registry) GetV2Descriptor(ctx context.Context, name, tag string) (distribution.Descriptor, error) {
	url := r.URL + "/v2/" + name + "/manifests/" + tag
	mediaType := "application/vnd.docker.distribution.manifest.v2+json"
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	req.Header.Add("Accept", mediaType)

	resp, err := r.Transport.RoundTrip(req)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 && len(resp.Header.Get("Docker-Content-Digest")) > 0 {
		return descriptorFromResponse(resp)
	} else {
		return distribution.Descriptor{}, fmt.Errorf("can't retrive description for %s:%s", name, tag)
	}

}

func descriptorFromResponse(response *http.Response) (distribution.Descriptor, error) {
	desc := distribution.Descriptor{}
	headers := response.Header

	ctHeader := headers.Get("Content-Type")
	if ctHeader == "" {
		return distribution.Descriptor{}, errors.New("missing or empty Content-Type header")
	}
	desc.MediaType = ctHeader

	digestHeader := headers.Get("Docker-Content-Digest")
	if digestHeader == "" {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return distribution.Descriptor{}, err
		}
		_, desc, err := distribution.UnmarshalManifest(ctHeader, data)
		if err != nil {
			return distribution.Descriptor{}, err
		}
		return desc, nil
	}

	dgst, err := digest.Parse(digestHeader)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	desc.Digest = dgst

	lengthHeader := headers.Get("Content-Length")
	if lengthHeader == "" {
		return distribution.Descriptor{}, errors.New("missing or empty Content-Length header")
	}
	length, err := strconv.ParseInt(lengthHeader, 10, 64)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	desc.Size = length

	return desc, nil
}
