[![Build Status](https://travis-ci.com/lancer-kit/noble.svg?branch=master)](https://travis-ci.com/github/lancer-kit/noble)
[![GoDoc](https://godoc.org/github.com/lancer-kit/noble?status.png)](https://godoc.org/github.com/lancer-kit/noble)
[![Go Report Card](https://goreportcard.com/badge/github.com/lancer-kit/noble)](https://goreportcard.com/report/github.com/lancer-kit/noble)

# Noble. Config files secret storage.

No more secrets in github/gitlab repo's

> ## New in 1.2
> #### Support for including multiple secrets inside parameter strings. 
> #### Support for different types of secrets within one line. 
Example of yaml config line:
```yaml
db: postgres://{{env:dbname}}?user={{dynenv:USER_NAME}}&password={{vault:/data/db?password}}
```

### TOC

* [Basic functionality](#-noblesecret)
* [Extension "simplecrypt".Just crypted strings in your config](#extension-simplecrypt)
* [Extension "etcdr2".Read from **etcd** key/value API v2, ](#etcdr2)
* [Extention "files".](#files)
* [Extention "vaultx". Read stored keys from Hashicorp Vault](#vault)

# noble.Secret
-----------

### Simple string wrapper type for secret storage in config files.

One type with support registration extensions (like a sql driver). 

* raw - just string (for debug/developing)
* env - read parameter from environment variable (see examples). read once on read from yaml/json
* dynenv - read parameter from environment variable without caching (every time when you call .Get())
* vault - read key from secure storage (**hashicorp vault**)
* etcd2 - read value from selected key stored on **ETCD** by API v2  
* scr - simple crypt value
* file - read first line from text file as secret value

Build-in (does not require importing extensions) supported storage type prefixes: raw, env, dynenv
#### YAML config example:

```yaml
db:
  name: "sample"
  #just for development - store as is, change store type to scr, env/dynenv in prod
  user: 'raw:test:test?/test'
  #read password stored in DB_PASS environment variable once (cache value)
  pass: env:DB_PASS
  #some datasource as string value
  url: host=prodhost dbname=proddb user={{env:DB_USER}} password={{dynenv:DB_PASSWORD}} sslmode={{dynenv:SSL_MODE}}
#read password stored in DB_PASS every time
pass2: "dynenv:DB_PASS"
```

#### Configure environment example:

```bash
export DB_PASS=SomeStrongPassword
export DB_USER=MyDBUSer
export SSL_MODE=required
```

#### Usage example:

```go
package main

import (
	"github.com/lancer-kit/noble"
	"gopkg.in/yaml.v2"
)

type testConfig struct {
	Db struct {
		Name string       `yaml:"name"`
		User noble.Secret `yaml:"user"`
		Pass noble.Secret `yaml:"pass"`
		Url  noble.Secret `yaml:"url"`
	} `yaml:"db"`
	Pass2 noble.Secret     `yaml:"pass2"`
}

func some(user, pass, pass2 string){
    //user == "test:test?/test"
    //pass == "SomeStrongPassword"
}

func main(){
    var configData []byte
    // read config here
    var cfg testConfig
    if e := yaml.Unmarshal(configData, &cfg);e!=nil{
        panic(e)
    }
    //use config
    some(cfg.Db.User.Get(), cfg.Db.Pass.Get(), cfg.Pass2.Get())
}

```

### Extension "simplecrypt"

Add type extension:

* scr - simple crypt value

##### Yaml config example:

````yaml
secret: scr:1Y2qKTtkeg5SmboJ970qENd54oBepinL5SF4dujQkY5Ec/J7M3bWQfiWaEPsZaXl4bPAEKoC1i29
msg: "This is your key {{scr:1Y2qKTtkeg5SmboJ970qENd54oBepinL5SF4dujQkY5Ec/J7M3bWQfiWaEPsZaXl4bPAEKoC1i29}} to use"
````

where scr - extension prefix

Build and use simplecrypt/encrypter to create key and encrypt values.

````
NAME:
   encrypter - Command line tool for encrypt secrets

USAGE:
   encrypter [global options] command [command options] [arguments...]

VERSION:
   0.1.12

COMMANDS:
   key, k      generate new secret key to store in environment variable SCR_PASS
   encrypt, e  encrypt value by key
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
----------------

NAME:
   encrypter encrypt - encrypt value by key

USAGE:
   encrypter encrypt [command options] [arguments...]

OPTIONS:
   --value value, -v value  value to encrypt
   --key value, -k value    key to encrypt value. [$SCR_PASS]
   --key value, -k value    key to encrypt value. [$SCR_PASS]
````

##### Usage:

> Just import package

````go
package main

import _ "github.com/lancer-kit/noble/simplecrypt"

//.... see example above
````

### ETCDR2

### Extension for etcd key/value API v2, "etcdr2"

Add type extension:

* etcd2 - read value from selected key stored on ETCD by API v2

##### Yaml config example:

````yaml
secret: "etcd2:messages4/test"
secret2: "etcd2:test2"
secret3: "etcd2:messages4/keybox/test"
user:
  msg: "Your application ID:{{etcd2:test2}} to use with this APP!"
````

Store value example:

````bash
curl http://127.0.0.1:2379/v2/keys/messages4/test -XPUT -d value="Hello world"
curl http://127.0.0.1:2379/v2/keys/test2 -XPUT -d value="Some very secret value"
curl http://127.0.0.1:2379/v2/keys/messages4/keybox/test -XPUT -d value="One more secret value"
````

##### Usage:

> Just import package
>
> Extension will be registered automatically

````go
package main

import _ "github.com/lancer-kit/noble/etcdr2"

//.... see example above
````

### Files

### Extension "files"

Add type extension:

* file - read first line from text file as secret value

##### Yaml config example:

````yaml
secret: "file:/etc/noble/secret.cfg"
user:
  msg: "The first line is:{{file:./file.txt}}"
````

##### Usage:

> Just import package

````go
package main

import _ "github.com/lancer-kit/noble/files"

//.... see example above
````

### Vault

### Extension "vaultx"

Add type extension:

* vault - read key from secure storage (hashicorp vault)
  Key format:

`/<path>?<key>`

For example, stored by command:

```ssh
vault kv put secret/data/some-secured pass="my long password"    
```

can be read by:

```yaml
password: "vault:/data/some-secured?pass"
someURI: "https://you.site.domain/return?token={{vault:/data?token}}&redirect={{env:PAGE}}"
```

##### Yaml config example:

````yaml
secret: "vault:/data?key"
````

##### Usage:

> Just import package

````go
package main

import (
	"github.com/lancer-kit/noble/vaultx"
	"log"
)

//....
func loadConfig() {
	// initialize vault reader
	vaultx.SetServerAddress("https://vault.server.lan:2345")
	if !vaultx.SetTokenEnv("VAULT_TOKEN") {
		log.Fatal("environment var VAULT_TOKEN not set")
	}
	if err := vaultx.InitVault(nil); err != nil {
		log.Fatal(err)
	}
	// ... then load config file  
}
````

It is also possible to configure the following parameters:

* `vaultx.SetLogger(logEntry)`: set logrus entry as log source;
* `vaultx.SetServerAddress(address)`: set vault server address;
* `vaultx.SetSecretPath(path)`: set vault k/v path. Used secret/data by default;
* `vaultx.SetToken(token)`: set vault token to login
* `vaultx.SetTokenEnv(envVarName)`: set vault token to login from environment var
