package simplecrypt

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/lancer-kit/noble"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	testKey   = "QlpUSUNRWVRTSE1QN1pTVkg3VEpMWkVBUDQ3TFdYVjc"
	testValue = "This is the secret key to store in config"
	testYaml  = `
secret: scr:1Y2qKTtkeg5SmboJ970qENd54oBepinL5SF4dujQkY5Ec/J7M3bWQfiWaEPsZaXl4bPAEKoC1i29
key: dynenv:SCR_PASS
`
	testJSON = `
{
"secret": "scr:1Y2qKTtkeg5SmboJ970qENd54oBepinL5SF4dujQkY5Ec/J7M3bWQfiWaEPsZaXl4bPAEKoC1i29",
"key": "dynenv:SCR_PASS"
}
`
)

type testCfg struct {
	Secret noble.Secret `yaml:"secret" json:"secret"`
	Key    noble.Secret `yaml:"key" json:"key"`
}

func TestReader_Read(t *testing.T) {
	r := Reader{}
	r.SetKey("")
	assert.Equal(t, defaultKeyBin, r.key)
	es, e := Encrypt(testValue, r.key)
	println("Encrypted:", es)
	assert.NoError(t, e)
	rs, e := r.Read(es)
	assert.NoError(t, e)
	assert.Equal(t, testValue, rs)
	r.SetKey(testKey)
	rs, e = r.Read(es)
	assert.Condition(t, func() (success bool) {
		return rs != testValue

	})
	es, e = Encrypt(testValue, r.key)
	assert.NoError(t, e)
	println("Encrypted:", es)
	rs, e = r.Read(es)
	assert.NoError(t, e)
	assert.Equal(t, testValue, rs)
}

func TestUnmarshalYAML(t *testing.T) {
	var x testCfg
	var r Reader
	os.Setenv(EnvVarName, testKey)
	r.SetKey(testKey)
	noble.Register("scr", r.Interface())
	e := yaml.Unmarshal([]byte(testYaml), &x)
	assert.NoError(t, e)
	v := x.Secret.Get()
	assert.NoError(t, x.Secret.InternalError())
	assert.Equal(t, testValue, v)
	assert.Equal(t, testKey, x.Key.Get())
}
func TestUnmarshalJSON(t *testing.T) {
	var x testCfg
	var r Reader
	os.Setenv(EnvVarName, testKey)
	r.SetKey(testKey)
	noble.Register("scr", r.Interface())
	e := json.Unmarshal([]byte(testJSON), &x)
	assert.NoError(t, e)
	x.Secret.Get()
	assert.Equal(t, testValue, x.Secret.Get())
	assert.Equal(t, testKey, x.Key.Get())
}
