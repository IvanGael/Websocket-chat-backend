package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

func xorEncrypt(text string) (string, error) {
	plaintext := []byte(text)
	key, err := generateRandomKey(len(plaintext))
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i++ {
		ciphertext[i] = plaintext[i] ^ key[i]
	}

	result := append(key, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

func xorDecrypt(encryptedText string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	if len(decoded) < 2 {
		return "", errors.New("invalid encrypted text")
	}

	keyLength := len(decoded) / 2
	key := decoded[:keyLength]
	ciphertext := decoded[keyLength:]

	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i++ {
		plaintext[i] = ciphertext[i] ^ key[i]
	}

	return string(plaintext), nil
}

func generateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
