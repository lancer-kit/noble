package simplecrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"

	"github.com/lancer-kit/noble"
)

const EnvVarName = "SCR_PASS"

var defaultKeyBin = []byte{9, 190, 70, 76, 8, 173, 169, 87, 168, 51, 8, 167, 66, 188, 73, 189, 90, 66, 153, 50, 2, 185, 72, 66, 168, 68, 71, 175, 168, 189, 52, 187}
var defaultKey = base64.RawStdEncoding.EncodeToString(defaultKeyBin)

// Encrypt string by symmetrical key
func Encrypt(in string, symKey []byte) (string, error) {
	b := []byte(in)
	block, err := aes.NewCipher(symKey)
	if err != nil {
		return "", err
	}
	cipherText := make([]byte, aes.BlockSize+len(b))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(b))

	return base64.RawStdEncoding.EncodeToString(cipherText), nil
}

// Decrypt string by symmetrical key
func Decrypt(in string, symKey []byte) (string, error) {
	block, err := aes.NewCipher(symKey)
	if err != nil {
		return "", err
	}
	text, err := base64.RawStdEncoding.DecodeString(in)
	if len(text) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	return string(text), nil
}

// Reader object, implements noble.SecretStorage
type Reader struct {
	key []byte
}

func init() {
	r := Reader{}
	noble.Register("scr", r.SetKey(os.Getenv(EnvVarName)).Clone())
}

// Read implementation
func (scr *Reader) Read(s string) (string, error) {
	return Decrypt(s, scr.key)
}

// SetKey set key for
func (scr *Reader) SetKey(new string) *Reader {
	if new == "" {
		scr.key = defaultKeyBin
		return scr
	}

	var err error
	if scr.key, err = base64.RawStdEncoding.DecodeString(new); err != nil {
		panic(err)
	}
	return scr
}

// Clone returns new empty instance of Reader
func (scr *Reader) Clone() noble.SecretStorage {
	r := &Reader{}
	r.key = scr.key

	return r
}
