#noble.Secret 

###Simple string wrapper type for secret storage in config files.

One type with support registration extensions (like a sql driver).
Build-in supported storage type prefixes:

* raw - just string (for debug/developing) 
* env - read parameter from environment variable (see examples). read once on read from yaml/json
* dynenv - read parameter from environment variable without caching (every time when you call .Get())
####YAML config example:
```yaml
db:
  name: "sample"
  #just for development - store as is, change store type to scr, env/dynenv in prod
  user: 'raw:test:test?/test'
  #read password stored in $DB_PASS once (cache value)
  pass: env:DB_PASS

#read password stored in $DB_PASS every time
pass2: "dynenv:DB_PASS"
```

####Configure environment example:
```bash
export DB_PASS=SomeStrongPassword
```
 

####Usage example:
```go
package main
import (
  "github.com/lancer-kit/noble"
  "gopkg.in/yaml.v2"
)

type testConfig struct {
	Db struct {
        Name  string       `yaml:"name"`
		User  noble.Secret `yaml:"user"`
		Pass  noble.Secret `yaml:"pass"`
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
###Extension "simplecrypt"

Add type extension:
* scr - simple crypt value
#####Yaml config example:
````yaml
secret: scr:1Y2qKTtkeg5SmboJ970qENd54oBepinL5SF4dujQkY5Ec/J7M3bWQfiWaEPsZaXl4bPAEKoC1i29
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