package util

import "os"

func MkdirTemp() (string, error) {
	dir, err := os.MkdirTemp("", "hf_")
	if err != nil {
		return "", err
	}
	return dir, nil
}
