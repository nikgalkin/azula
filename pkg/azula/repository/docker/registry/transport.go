package registry

import (
	"net/http"
	"net/url"

	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
)

type regCredentialStore struct {
	username      string
	password      string
	refreshTokens map[string]string
}

func NewTrans(user, pass, url string) (http.RoundTripper, error) {
	creds := &regCredentialStore{username: user, password: pass}
	challengeManager := challenge.NewSimpleManager()
	_, err := ping(challengeManager, url+"/v2/", "")
	if err != nil {
		return nil, err
	}
	trans := transport.NewTransport(
		nil, auth.NewAuthorizer(challengeManager, auth.NewBasicHandler(creds)),
	)
	return trans, nil
}

func (tcs *regCredentialStore) Basic(*url.URL) (string, string) {
	return tcs.username, tcs.password
}

func (tcs *regCredentialStore) RefreshToken(u *url.URL, service string) string {
	return tcs.refreshTokens[service]
}

func (tcs *regCredentialStore) SetRefreshToken(u *url.URL, service string, token string) {
	if tcs.refreshTokens != nil {
		tcs.refreshTokens[service] = token
	}
}

func ping(manager challenge.Manager, endpoint, versionHeader string) ([]auth.APIVersion, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := manager.AddResponse(resp); err != nil {
		return nil, err
	}

	return auth.APIVersions(resp, versionHeader), err
}
