package util

import (
	"io/ioutil"
)

func MkdirTemp() (string, error) {
	dir, err := ioutil.TempDir("","hf_")
	if err != nil {
		return "", err
	}
	return dir, nil
}