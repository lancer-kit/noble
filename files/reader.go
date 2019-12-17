package files

import (
	"bufio"
	"os"
	"strings"

	"github.com/lancer-kit/noble"
)

func init() {
	noble.Register("file", Reader{}.Interface())
}

//Reader object. Read data from file
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

func (r Reader) Interface() noble.SecretStorage {
	return &Reader{}
}
