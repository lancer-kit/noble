package files

import (
	"bufio"
	"os"
	"strings"

	"github.com/lancer-kit/noble"
)

func init() {
	noble.Register("file", &Reader{})
}

// Reader object. Read data from file
type Reader struct {
	fileName string
}

// Read file
func (r *Reader) Read(fileName string) (string, error) {
	r.fileName = fileName
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}

	rdr := bufio.NewReader(f)
	res, err := rdr.ReadString([]byte("\n")[0])
	if err != nil {
		return "", err
	}

	return strings.Replace(res, "\n", "", -1), nil
}

// Clone returns new empty instance of Reader
func (r Reader) Clone() noble.SecretStorage {
	return &Reader{}
}
