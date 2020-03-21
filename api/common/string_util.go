package common

import (
	"bytes"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

func JoinErrors(errs []error) error {
	var sb bytes.Buffer
	for _, err := range errs {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(err.Error())
	}
	return errors.New(sb.String())
}

func ReadFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	return string(b), nil
}
