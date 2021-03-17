package main

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/lancer-kit/armory/crypto"
	"github.com/lancer-kit/noble/simplecrypt"
	"github.com/urfave/cli"
)

func getCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "key",
			ShortName: "k",
			Usage:     "generate new secret key to store in environment variable " + simplecrypt.EnvVarName,
			Action: func(c *cli.Context) error {
				key := genKey()
				println("New key is:")
				println(key)
				return nil
			},
		},
		{
			Name:      "encrypt",
			ShortName: "e",
			Usage:     "encrypt value by key",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "value,v",
					Usage:    "value to encrypt",
					Required: true,
				},
				cli.StringFlag{
					Name:     "key,k",
					Usage:    "key to encrypt value.",
					EnvVar:   simplecrypt.EnvVarName,
					Required: false,
				},
			},
			Action: encryptValue,
		},
	}
}

func genKey() string {
	_, tmp := crypto.GenKeyPair()
	key := []byte(tmp[:32])
	println("Key len:", len(key))
	return base64.RawStdEncoding.EncodeToString(key)
}

func encryptValue(c *cli.Context) error {
	key := c.String("key")
	if key == "" {
		println(`Required flag "key" or not set`)
	}
	binKey, err := base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		println("Key decode error:", err.Error())
		return err
	}
	println("Using key:", key)
	val := c.String("value")
	enc, err := simplecrypt.Encrypt(val, binKey)
	if err != nil {
		println("Encode error:", err.Error())
		return err
	}
	check, err := simplecrypt.Decrypt(enc, binKey)
	if err != nil {
		println("Unable to decrypt:", err.Error())
		return err
	}
	if check != val {
		println("Decrypt error")
	}
	println("Encoded value is:")
	println(enc)
	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "Command line tool for encrypt secrets"
	app.Version = "0.1.12"
	app.Commands = getCommands()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)

	}
}
