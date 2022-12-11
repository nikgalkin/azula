package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Config represents the on disk format of the docker CLI's config file.
type Config struct {
	AuthConfigs       map[string]AuthConfig `json:"auths"`
	CredentialsStore  string                `json:"credsStore,omitempty"`
	CredentialHelpers map[string]string     `json:"credHelpers,omitempty"`
}

// AuthConfig contains authorization information for connecting to a Registry
type AuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`

	// Email is an optional value associated with the username.
	// This field is deprecated and will be removed in a later
	// version of docker.
	Email string `json:"email,omitempty"`

	ServerAddress string `json:"serveraddress,omitempty"`

	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`

	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

var (
	ErrCredentialsNotFound         = errors.New("credentials not found in native keychain")
	ErrCredentialsMissingServerURL = errors.New("no credentials server URL")
)

// This is used by the docker CLI in casses where an oauth identity token is used.
// In that case the username is stored literally as `<token>`
// When fetching the credentials we check for this value to determine if
const tokenUsername = "<token>"

// ResolveRegistryHost can be used to transform a docker registry host name into what is used for the docker config/cred helpers
// This is useful for using with containerd authorizers.
// Naturally this only transforms docker hub URLs.
func ResolveRegistryHost(host string) string {
	switch host {
	case "index.docker.io", "docker.io", "https://index.docker.io/v1/", "registry-1.docker.io":
		return "https://index.docker.io/v1/"
	}
	return host
}

// UserHomeConfigPath returns the path to the docker config in the current user's home dir.
func UserHomeConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error looking up user home dir: %w", err)
	}

	return filepath.Join(home, ".docker", "config.json"), nil
}

// ConfigPath returns the path to the docker cli config.
//
// It will either use the DOCKER_CONFIG env var if set, or the value from `UserHomeConfigPath`
// DOCKER_CONFIG would be the dir path where `config.json` is stored, this returns the path to config.json.
func ConfigPath() (string, error) {
	if p := os.Getenv("DOCKER_CONFIG"); p != "" {
		return filepath.Join(p, "config.json"), nil
	}
	return UserHomeConfigPath()
}

// LoadDefaultConfig loads the docker cli config from the path returned from `ConfigPath`
func LoadDefaultConfig() (Config, error) {
	var cfg Config
	p, err := ConfigPath()
	if err != nil {
		return cfg, err
	}
	return cfg, FromFile(p, &cfg)
}

// FromFile loads config from the specified path into cfg
func FromFile(configPath string, cfg *Config) error {
	f, err := os.Open(configPath)
	if err != nil {
		return err
	}

	err = json.NewDecoder(f).Decode(&cfg)
	f.Close()
	return err
}

// GetRegistryCredentials gets registry credentials for the passed in registry host.
//
// This will use `LoadDefaultConfig` to read registry auth details from the config.
// If the config doesn't exist, it will attempt to load registry credentials using the default credential helper for the platform.
func GetRegistryCredentials(hostname string) (string, string, error) {
	cfg, err := LoadDefaultConfig()
	if err != nil {
		if !os.IsNotExist(err) {
			return "", "", err
		}
		return GetCredentialsFromHelper("", hostname)
	}
	return cfg.GetRegistryCredentials(hostname)
}

// GetCredentialsFromHelper attempts to lookup credentials from the passed in docker credential helper.
//
// The credential helpoer should just be the suffix name (no "docker-credential-").
// If the passed in helper program is empty this will look up the default helper for the platform.
//
// If the credentials are not found, no error is returned, only empty credentials.
//
// Hostnames should already be resolved using `ResolveRegistryAuth`
//
// If the username string is empty, the password string is an identity token.
func GetCredentialsFromHelper(helper, hostname string) (string, string, error) {
	if helper == "" {
		helper = getCredentialHelper()
	}
	if helper == "" {
		return "", "", nil
	}

	p, err := exec.LookPath("docker-credential-" + helper)
	if err != nil {
		return "", "", nil
	}

	cmd := exec.Command(p, "get")
	cmd.Stdin = strings.NewReader(hostname)

	b, err := cmd.Output()
	if err != nil {
		s := strings.TrimSpace(string(b))

		switch s {
		case ErrCredentialsNotFound.Error():
			return "", "", nil
		case ErrCredentialsMissingServerURL.Error():
			return "", "", errors.New(s)
		default:
		}

		return "", "", err
	}

	var creds struct {
		Username string
		Secret   string
	}

	if err := json.Unmarshal(b, &creds); err != nil {
		return "", "", err
	}

	// When tokenUsername is used, the output is an identity token and the username is garbage
	if creds.Username == tokenUsername {
		creds.Username = ""
	}

	return creds.Username, creds.Secret, nil
}

// getCredentialHelper gets the default credential helper name for the current platform.
func getCredentialHelper() string {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("pass"); err == nil {
			return "pass"
		}
		return "secretservice"
	case "darwin":
		return "osxkeychain"
	case "windows":
		return "wincred"
	default:
		return ""
	}
}

// GetRegistryCredentials gets credentials, if any, for the provided hostname
//
// Hostnames should already be resolved using `ResolveRegistryAuth`
//
// If the returned username string is empty, the password is an identity token.
func (c *Config) GetRegistryCredentials(hostname string) (string, string, error) {
	h, ok := c.CredentialHelpers[hostname]
	if ok {
		return GetCredentialsFromHelper(h, hostname)
	}

	if c.CredentialsStore != "" {
		return GetCredentialsFromHelper(c.CredentialsStore, hostname)
	}

	auth, ok := c.AuthConfigs[hostname]
	if !ok {
		return GetCredentialsFromHelper("", hostname)
	}

	if auth.IdentityToken != "" {
		return "", auth.IdentityToken, nil
	}

	if auth.Username != "" && auth.Password != "" {
		return auth.Username, auth.Password, nil
	}

	return DecodeBase64Auth(auth)
}

// DecodeBase64Auth decodes the legacy file-based auth storage from the docker CLI.
// It takes the "Auth" filed from AuthConfig and decodes that into a username and password.
//
// If "Auth" is empty, an empty user/pass will be returned, but not an error.
func DecodeBase64Auth(auth AuthConfig) (string, string, error) {
	if auth.Auth == "" {
		return "", "", nil
	}

	decLen := base64.StdEncoding.DecodedLen(len(auth.Auth))
	decoded := make([]byte, decLen)
	n, err := base64.StdEncoding.Decode(decoded, []byte(auth.Auth))
	if err != nil {
		return "", "", fmt.Errorf("error decoding auth from file: %w", err)
	}

	if n > decLen {
		return "", "", fmt.Errorf("decoded value is longer than expected length, expected: %d, actual: %d", decLen, n)
	}

	split := strings.SplitN(string(decoded), ":", 2)
	if len(split) != 2 {
		return "", "", errors.New("invalid auth string")
	}

	return split[0], strings.Trim(split[1], "\x00"), nil
}
