package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"strings"
)

func GCM_encrypt(key string, plaintext string, iv []byte, additionalData []byte) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "error", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "error", err
	}
	ciphertext := aesgcm.Seal(nil, iv, []byte(plaintext), additionalData)
	stringed := hex.EncodeToString(iv) + "-" + hex.EncodeToString(ciphertext)
	return stringed, nil
}

func GCM_decrypt(key string, ct string, iv string, additionalData []byte) (string, error) {
	ciphertext, _ := hex.DecodeString(ct)
	iv_decode, _ := hex.DecodeString(iv)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "error", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "error", err
	}
	plaintext, err := aesgcm.Open(nil, iv_decode, ciphertext, additionalData)
	if err != nil {
		return "error", err
	}
	s := string(plaintext[:])
	return s, nil
}

func generateSalt(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func GenerateFromPassword(password string) string {

	salt, _ := generateSalt(8)
	nHash := pbkdf2.Key([]byte(password), salt, 4096, sha256.Size, sha256.New)
	stringed := hex.EncodeToString(salt) + "-" + hex.EncodeToString(nHash)

	return stringed
}

func CompareHashAndPassword(hashedPassword string, password string) error {

	token := strings.Split(hashedPassword, "-")
	salt, _ := hex.DecodeString(token[0])
	sHash, _ := hex.DecodeString(token[1])
	rHash := pbkdf2.Key([]byte(password), salt, 4096, sha256.Size, sha256.New)
	if bytes.Equal(sHash, rHash) {
		return nil
	}

	return errors.New("invalid password")

}
