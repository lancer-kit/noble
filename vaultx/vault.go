package vaultx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

const (
	dataKey = "data"

	contentType       = "application/json"
	contentTypeHeader = "Content-Type"
	vaultTokenHeader  = "X-Vault-Token" // #nosec
)

type (
	// VaultCfg vault config
	VaultCfg struct {
		ServerAddress string // vault server address
		SecretPath    string
		// Timeout and Refresh Timestamp in hours
		TokenTTLHours         int64
		TokenRefreshTimeHours int64
		Token                 string
	}

	vaultStorage struct {
		client *api.Client
		config *api.Config
		logger *logrus.Entry
		stop   chan bool
	}

	storage interface {
		Get(path, key string) (string, error)
		SetLogger(*logrus.Entry)
	}

	response struct {
		Auth struct {
			Token string `json:"client_token"`
		} `json:"auth"`
	}
)

const createTokenPath = "/v1/auth/token/create"

//nolint:gochecknoglobals
var defaultConfig = VaultCfg{
	ServerAddress:         "http://127.0.0.1:8000",
	SecretPath:            "secret/data",
	TokenTTLHours:         3,
	TokenRefreshTimeHours: 1,
}

func (v *vaultStorage) SetLogger(l *logrus.Entry) {
	v.logger = l
}

func (v *vaultStorage) Get(path, key string) (string, error) {

	secret, err := v.client.Logical().ReadWithData(defaultConfig.SecretPath+path, nil)
	if err != nil {
		return "", err
	}

	if secret == nil {
		return "", nil
	}

	rawData, ok := secret.Data[dataKey]
	if !ok {
		return "", fmt.Errorf("data block not exists in key '%s'", defaultConfig.SecretPath+path)
	}

	data, ok := rawData.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("failed to parse data block: %v", rawData)
	}

	value, ok := data[key]

	if !ok {
		return "", nil
	}

	// convert interface to string
	res := ""
	switch typedValue := value.(type) {
	case string:
		res = typedValue
	case int64:
		res = fmt.Sprintf("%d", typedValue)
	case float64:
		res = fmt.Sprintf("%f", typedValue)
	default:
		res = fmt.Sprintf("%v", typedValue)
	}
	return res, nil
}

func newStorage(cfg VaultCfg) (storage, error) {
	logger := logrus.New().WithField("app_layer", "noble.nvault")
	apiConfig := api.DefaultConfig()
	apiConfig.Address = cfg.ServerAddress
	client, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}

	token, err := login(client, cfg)
	if err != nil {
		return nil, err
	}

	client.SetToken(token)

	// check connection
	_, err = client.Sys().Health()
	if err != nil {
		return nil, err
	}

	storage := &vaultStorage{
		client: client,
		config: apiConfig,
		logger: logger,
	}
	if defaultConfig.TokenRefreshTimeHours != 0 {
		go storage.refreshToken(logger)
		runtime.SetFinalizer(storage, stopRefresh)
	}
	return storage, nil
}

func stopRefresh(v *vaultStorage) {
	v.stop <- true
}

func (v *vaultStorage) refreshToken(logger *logrus.Entry) {
	ticker := time.NewTicker(time.Duration(defaultConfig.TokenRefreshTimeHours) * time.Hour)
	for {
		select {
		case <-v.stop:
			return
		case <-ticker.C:
			token, err := login(v.client, defaultConfig)
			if err != nil {
				logger.WithError(err).Error("failed to refresh vault token")
			}

			v.client.SetToken(token)
			h, err := v.client.Sys().Health()
			if err != nil {
				logger.WithError(err).WithField("cluster_name", h.ClusterName).
					WithField("cluster_id", h.ClusterID).Error("vault health check failed")
			}
		}
	}
}

func login(client *api.Client, cfg VaultCfg) (string, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"ttl":       fmt.Sprintf("%dh", cfg.TokenTTLHours),
		"renewable": true,
	})

	if err != nil {
		return "", err
	}

	u, err := url.Parse(fmt.Sprintf("%s%s", cfg.ServerAddress, createTokenPath))
	if err != nil {
		return "", err
	}

	resp, err := client.RawRequest(&api.Request{
		Method: http.MethodPost,
		URL:    u,
		Headers: http.Header{
			vaultTokenHeader:  []string{cfg.Token},
			contentTypeHeader: []string{contentType},
		},
		Body:        bytes.NewBuffer(requestBody),
		ClientToken: cfg.Token,
	})
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	result := new(response)
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.Auth.Token, nil
}
