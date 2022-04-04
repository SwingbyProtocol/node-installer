package keystore

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/secretbox"
)

func encryptKeyStore(secretHex string, data *SaveData) ([]byte, error) {
	if len(secretHex) != 64 {
		return nil, fmt.Errorf("expected a KSS key size of 64 hex chars but got %d", len(secretHex))
	}
	secretKeyBz, err := hex.DecodeString(secretHex)
	if err != nil {
		return nil, err
	}
	var secretKey [32]byte
	copy(secretKey[:], secretKeyBz)

	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	// this appends the encrypted cypher text to the nonce for writing to file
	encrypted := secretbox.Seal(nonce[:], marshalled, &nonce, &secretKey)
	return encrypted, nil
}

func decryptKeyStore(secretHex string, encrypted []byte) (*SaveData, error) {
	if len(encrypted) <= 24 {
		return nil, fmt.Errorf("expected the encrypted keystore to contain a nonce and data but is %d in length",
			len(encrypted))
	}
	if len(secretHex) != 64 {
		return nil, fmt.Errorf("expected a KSS key size of 64 hex chars but got %d", len(secretHex))
	}
	secretKeyBz, err := hex.DecodeString(secretHex)
	if err != nil {
		return nil, err
	}
	var secretKey [32]byte
	copy(secretKey[:], secretKeyBz)

	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &decryptNonce, &secretKey)
	if !ok || 0 == len(strings.TrimSpace(string(decrypted))) {
		return nil, errors.New("keystore decryption error")
	}

	data := SaveData{}
	if err := json.Unmarshal(decrypted, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
