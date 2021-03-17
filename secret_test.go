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
  db_url: "sqlite://{{dynenv:DB_NAME}}?enableSome={{env:SEC}}"
`
const dbUrl = "sqlite://some?enableSome=sec"

const jsonTest = `
{
"db":{
	"user":"raw:test:test?/test",
	"pass":"env:DB_PASS",
	"pass2":"dynenv:DB_PASS",
	"db_url":"sqlite://{{dynenv:DB_NAME}}?enableSome={{env:SEC}}"
	}
}
`

type testConfig struct {
	Db struct {
		User  Secret `yaml:"user" json:"user"`
		Pass  Secret `yaml:"pass" json:"pass"`
		Pass2 Secret `yaml:"pass2" json:"pass2"`
		DbUrl Secret `yaml:"db_url" json:"db_url"`
	} `yaml:"db" json:"db"`
}

func TestSecretWrapper_New(t *testing.T) {
	v := Secret{}.New("raw:test")
	assert.NoError(t, v.InternalError())
	assert.Equal(t, "test", v.Get())
}

func TestSecretWrapper_UnmarshalYAML(t *testing.T) {
	tst := testConfig{}
	if !assert.NoError(t, os.Setenv("DB_PASS", envPass)) {
		return
	}
	assert.NoError(t, os.Setenv("DB_NAME", "some"))
	assert.NoError(t, os.Setenv("SEC", "sec"))
	e := yaml.Unmarshal([]byte(yamlTest), &tst)
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
	assert.Equal(t, dbUrl, tst.Db.DbUrl.Get())
	assert.NoError(t, tst.Db.Pass2.InternalError())
	assert.NoError(t, os.Unsetenv("DB_PASS"))
	assert.NoError(t, os.Unsetenv("DB_NAME"))
	assert.NoError(t, os.Unsetenv("SEC"))
}

func TestSecretWrapper_UnmarshalJSON(t *testing.T) {
	tst := testConfig{}
	assert.NoError(t, os.Setenv("DB_NAME", "some"))
	assert.NoError(t, os.Setenv("SEC", "sec"))
	e := json.Unmarshal([]byte(jsonTest), &tst)
	if !assert.NoError(t, e) {
		return
	}
	assert.NoError(t, os.Setenv("DB_PASS", envPass))
	assert.Equal(t, user, tst.Db.User.Get())
	assert.Equal(t, envPass, tst.Db.Pass.Get())
	assert.NoError(t, tst.Db.Pass.InternalError())
	assert.NoError(t, os.Setenv("DB_PASS", envPass+user))
	assert.Equal(t, envPass+user, tst.Db.Pass2.Get())
	assert.NoError(t, tst.Db.Pass2.InternalError())
	assert.Equal(t, dbUrl, tst.Db.DbUrl.Get())
	assert.NoError(t, tst.Db.DbUrl.InternalError())
	assert.NoError(t, tst.Db.DbUrl.ParseError())
	assert.NoError(t, os.Unsetenv("DB_PASS"))
	assert.NoError(t, os.Unsetenv("DB_NAME"))
	assert.NoError(t, os.Unsetenv("SEC"))
}

func TestSecret_Error(t *testing.T) {
	s := Secret{}.New("some-test-data{{error here{{ }")
	assert.Error(t, s.ParseError())
	t.Logf("parse error: %s", s.Error())
	assert.NotEmpty(t, s.Error())
	s = Secret{}.New("some-test-data")
	assert.Error(t, s.ParseError())
	t.Logf("parse error: %s", s.Error())
	assert.Equal(t, "some-test-data", s.Get())
	assert.Error(t, s.InternalError())
	assert.NotEmpty(t, s.Error())
	s = Secret{}.New("{{raw:some-test-data}}")
	assert.NoError(t, s.ParseError())
	assert.NoError(t, s.InternalError())
	assert.Empty(t, s.Error())
	assert.Equal(t, "some-test-data", s.Get())
	t.Logf("data is:%s", s.Get())
	s = Secret{}.New("here is:\t{{raw:some-test-data}} new value of \n{{raw:test}} {.more}")
	assert.NoError(t, s.ParseError())
	assert.NoError(t, s.InternalError())
	assert.Empty(t, s.Error())
	t.Logf("data is:%s", s.Get())
	assert.Equal(t, "here is:\tsome-test-data new value of \ntest {.more}", s.Get())
	s = Secret{}.New("{{env:some-test-data}}")
	assert.Error(t, s.ParseError())
	assert.Error(t, s.InternalError())
}

func TestSecret_NoError(t *testing.T) {

}

func Test_secret_new(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name       string
		args       args
		wantError  bool
		parseError bool
	}{
		{
			name: "raw",
			args: args{s: "raw:some"},
		},
		{
			name: "env",
			args: args{s: "env:PATH"},
		},
		{
			name:      "error read",
			args:      args{s: "env:SOME_NOT_EXISTING"},
			wantError: true,
		},
		{
			name:       "error parse",
			args:       args{s: "SOME_NOT_EXISTING"},
			parseError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sec := secret{}.new(tt.args.s)
			_ = sec.Get()
			if tt.wantError {
				assert.Error(t, sec.internal)
			} else {
				assert.NoError(t, sec.internal)
			}
			if tt.parseError {
				assert.Error(t, sec.parseError)
			} else {
				assert.NoError(t, sec.parseError)
			}
		})
	}
}
