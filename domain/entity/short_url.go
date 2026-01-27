package entity

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"time"
)

var (
	ErrEmptyURL   = errors.New("URL cannot be empty")
	ErrInvalidURL = errors.New("invalid URL format")
)

const DefaultTokenLength = 4

type ShortURL struct {
	Token     string
	FullURL   string
	CreatedAt time.Time
}

func NewShortURL(fullURL string) (*ShortURL, error) {
	if fullURL == "" {
		return nil, ErrEmptyURL
	}

	parsed, err := url.Parse(fullURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, ErrInvalidURL
	}

	token, err := generateToken(DefaultTokenLength)
	if err != nil {
		return nil, err
	}

	return &ShortURL{
		Token:     token,
		FullURL:   fullURL,
		CreatedAt: time.Now(),
	}, nil
}

func (s *ShortURL) IsExpired(ttl time.Duration) bool {
	return time.Since(s.CreatedAt) > ttl
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(bytes)
	return token[:length], nil
}
