package seal

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"time"

	"github.com/jiangjiali/vault/sdk/helper/armon/metrics"
	"github.com/jiangjiali/vault/sdk/helper/errwrap"
	"github.com/jiangjiali/vault/sdk/helper/xxuuid"
)

type Envelope struct{}

type EnvelopeInfo struct {
	Ciphertext []byte
	Key        []byte
	IV         []byte
}

func NewEnvelope() *Envelope {
	return &Envelope{}
}

func (e *Envelope) Encrypt(plaintext []byte) (*EnvelopeInfo, error) {
	defer metrics.MeasureSince([]string{"seal", "envelope", "encrypt"}, time.Now())

	// Generate DEK
	key, err := xxuuid.GenerateRandomBytes(32)
	if err != nil {
		return nil, err
	}
	iv, err := xxuuid.GenerateRandomBytes(12)
	if err != nil {
		return nil, err
	}
	aead, err := e.aeadEncrypter(key)
	if err != nil {
		return nil, err
	}
	return &EnvelopeInfo{
		Ciphertext: aead.Seal(nil, iv, plaintext, nil),
		Key:        key,
		IV:         iv,
	}, nil
}

func (e *Envelope) Decrypt(data *EnvelopeInfo) ([]byte, error) {
	defer metrics.MeasureSince([]string{"seal", "envelope", "decrypt"}, time.Now())

	aead, err := e.aeadEncrypter(data.Key)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, data.IV, data.Ciphertext, nil)
}

func (e *Envelope) aeadEncrypter(key []byte) (cipher.AEAD, error) {
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, errwrap.Wrapf("failed to create cipher: {{err}}", err)
	}

	// Create the GCM mode AEAD
	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, errors.New("failed to initialize GCM mode")
	}

	return gcm, nil
}
