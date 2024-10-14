package tool

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

func LoadJSONFile(path string, data interface{}) error {
	filePtr, err := os.Open(path)
	if err != nil {
		return err
	}
	defer filePtr.Close()
	err = json.NewDecoder(filePtr).Decode(data)
	if err != nil {
		return err
	}
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.Wrap(err, "check path exists")
}
