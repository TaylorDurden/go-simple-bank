package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

// PasetoAuthenticator is a PASETO token
type PasetoAuthenticator struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// CreateToken creates a new token for a specific username and duration
func (p PasetoAuthenticator) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}
	token, err := p.paseto.Encrypt(p.symmetricKey, payload, nil)
	return token, payload, err
}

// VerifyToken checks if the token is valid or not
func (p PasetoAuthenticator) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := p.paseto.Decrypt(token, p.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func NewPasetoAuthenticator(symmetricKey string) (Authenticator, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	auth := &PasetoAuthenticator{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return auth, nil
}
