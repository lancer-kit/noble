package etcdr2

import (
	"github.com/lancer-kit/armory/api/httpx"
	"github.com/lancer-kit/noble"
	"github.com/pkg/errors"
)

// EtcdConnectionString default connection string
var EtcdConnectionString = "http://127.0.0.1:2379" //nolint:gochecknoglobals

//nolint:gochecknoinits
func init() {
	noble.Register("etcd2", &KeyReader{})
}

// KeyReader type implements noble.SecretStorage
type KeyReader struct {
}

type v2Message struct {
	Node struct {
		Value string `json:"value"`
	} `json:"node"`
}

// Read key value from etcd API v2
func (r *KeyReader) Read(key string) (string, error) {
	rsp, err := httpx.GetJSON(EtcdConnectionString+"/v2/keys/"+key, nil)
	if err != nil {
		return "", err
	}
	if rsp.StatusCode != 200 {
		return "", errors.Errorf("invalid etcd api v2 status code: %d", rsp.StatusCode)
	}
	var msg v2Message
	if err := httpx.ParseJSONResult(rsp, &msg); err != nil {
		return "", err
	}
	return msg.Node.Value, nil
}

// Clone returns new empty instance of KeyReader
func (r *KeyReader) Clone() noble.SecretStorage {
	return &KeyReader{}
}
