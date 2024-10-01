package token

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type TokenInfo struct {
	value      string
	validUntil time.Time
	valid      bool
}

func (t TokenInfo) expired() bool {
	return t.validUntil.Before(time.Now())
}

type TokenManager interface {
	HasToken(string) bool
	GetToken(string) (string, error)
	ResetToken(string)
	ValidateToken(string, string) (bool, error)
	InvalidateToken(string)
}

func generateToken(key string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(key+time.Now().String()), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal().Err(err).Msg("could not generate token")
	}

	hasher := md5.New()
	hasher.Write(hash)
	return hex.EncodeToString(hasher.Sum(nil))

}

type TokenManagerInMemory struct {
	tokens map[string]TokenInfo
}

func NewTokenManagerInMemory() *TokenManagerInMemory {
	return &TokenManagerInMemory{
		tokens: map[string]TokenInfo{},
	}
}

func (tm *TokenManagerInMemory) HasToken(key string) bool {
	_, exists := tm.tokens[key]
	return exists
}

func (tm *TokenManagerInMemory) GetToken(key string) (string, error) {
	token, exists := tm.tokens[key]
	if exists && !token.expired() {
		return token.value, nil
	}
	token = TokenInfo{
		value:      generateToken(key),
		validUntil: time.Now().Add(time.Minute * 10),
		valid:      true,
	}
	tm.tokens[key] = token
	log.Info().Str("token", token.value).Str("key", key).Msg("New token created")
	return token.value, nil
}

func (tm *TokenManagerInMemory) ValidateToken(key string, tokenValue string) (bool, error) {
	token, exists := tm.tokens[key]
	if !exists {
		return false, errors.New("No token for key found")
	}

	if token.expired() {
		log.Warn().Time("validUntil", token.validUntil).Time("time", time.Now())
		return false, errors.New("Token has expired")
	}

	return true, nil

}

func (tm *TokenManagerInMemory) ResetToken(key string) {
	delete(tm.tokens, key)
}

func (tm *TokenManagerInMemory) InvalidateToken(key string) {
	token, exists := tm.tokens[key]
	if exists {
		tokenClone := TokenInfo{
			value:      token.value,
			validUntil: token.validUntil,
			valid:      false,
		}
		tm.tokens[key] = tokenClone
	}
}
