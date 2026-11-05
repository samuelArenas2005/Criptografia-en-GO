package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// Keystore almacena la clave y el nonce en formato hexadecimal.
type Keystore struct {
	Key   string `json:"key"`
	Nonce string `json:"nonce"`
}

// EncryptLog cifra el archivo en reposo utilizando AES-256-GCM y almacena clave/nonce.
func EncryptLog(plainPath, encPath, keystorePath string) error {
	plaintext, err := os.ReadFile(plainPath)
	if err != nil {
		return err
	}

	key := make([]byte, 32)
	rand.Read(key)

	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	if err := os.WriteFile(encPath, ciphertext, 0644); err != nil {
		return err
	}

	ks, _ := json.MarshalIndent(Keystore{
		Key:   hex.EncodeToString(key),
		Nonce: hex.EncodeToString(nonce),
	}, "", "  ")

	return os.WriteFile(keystorePath, append(ks, '\n'), 0644)
}

// DecryptLog descifra el archivo binario utilizando la clave y nonce almacenados.
func DecryptLog(encPath, keystorePath, decPath string) error {
	ksData, err := os.ReadFile(keystorePath)
	if err != nil {
		return err
	}

	var ks Keystore
	_ = json.Unmarshal(ksData, &ks)

	key, _ := hex.DecodeString(ks.Key)
	nonce, _ := hex.DecodeString(ks.Nonce)

	ciphertext, err := os.ReadFile(encPath)
	if err != nil {
		return err
	}

	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("error de autenticación: archivo manipulado o clave incorrecta")
	}

	return os.WriteFile(decPath, plaintext, 0644)
}
