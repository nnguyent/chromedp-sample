package utils

import (
	"crypto/rand"
	"path/filepath"

	"github.com/pkg/errors"
)

const otpChars = "1234567890"

func ListFilesInDir(dirPath string) ([]string, error) {
	pattern := dirPath + "/*.*"
	files, err := filepath.Glob(pattern)
	if err != nil {
		err = errors.Wrapf(err, "failed to list files in dir:%s", dirPath)
		return nil, err
	}

	return files, nil
}

func GenCode(length int) string {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return ""
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer)
}
