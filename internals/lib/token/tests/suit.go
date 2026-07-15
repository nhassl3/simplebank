package tests

import "crypto/rand"

func GenerateRandomBytes(length int) ([]byte, error) {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func GenerateRandomString(length int) (string, error) {
	key, err := GenerateRandomBytes(length)
	if err == nil {
		return string(key), nil
	}
	return "", err
}
