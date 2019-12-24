package noble

import (
	"encoding/json"
	"os"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

const user = "test:test?/test"
const envPass = "09HHHbbaa$bbbbbbsss&&ammskjjjjdjd"

const yamlTest = `
db:
  user: 'raw:test:test?/test'
  pass: env:DB_PASS
  pass2: "dynenv:DB_PASS"
`

const jsonTest = `
{
"db":{
	"user":"raw:test:test?/test",
	"pass":"env:DB_PASS",
	"pass2":"dynenv:DB_PASS"
	}
}
`

type testConfig struct {
	Db struct {
		User  Secret `yaml:"user" json:"user"`
		Pass  Secret `yaml:"pass" json:"pass"`
		Pass2 Secret `yaml:"pass2" json:"pass2"`
	} `yaml:"db" json:"db"`
}

func TestSecretWrapper_New(t *testing.T) {
	v := Secret{}.New("raw:test")
	assert.NoError(t, v.InternalError())
	assert.Equal(t, "test", v.Get())
}

func TestSecretWrapper_UnmarshalYAML(t *testing.T) {
	tst := testConfig{}
	e := os.Setenv("DB_PASS", envPass)
	if !assert.NoError(t, e) {
		return
	}
	e = yaml.Unmarshal([]byte(yamlTest), &tst)
	if !assert.NoError(t, e) {
		return
	}
	e = os.Setenv("DB_PASS", envPass+user)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, user, tst.Db.User.Get())
	assert.Equal(t, envPass, tst.Db.Pass.Get())
	assert.NoError(t, tst.Db.Pass.InternalError())
	assert.Equal(t, envPass+user, tst.Db.Pass2.Get())
	assert.NoError(t, tst.Db.Pass2.InternalError())
}

func TestSecretWrapper_UnmarshalJSON(t *testing.T) {
	tst := testConfig{}
	e := os.Setenv("DB_PASS", envPass)
	if !assert.NoError(t, e) {
		return
	}
	e = json.Unmarshal([]byte(jsonTest), &tst)
	if !assert.NoError(t, e) {
		return
	}
	e = os.Setenv("DB_PASS", envPass+user)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, user, tst.Db.User.Get())
	assert.Equal(t, envPass, tst.Db.Pass.Get())
	assert.NoError(t, tst.Db.Pass.InternalError())
	assert.Equal(t, envPass+user, tst.Db.Pass2.Get())
	assert.NoError(t, tst.Db.Pass2.InternalError())
}
