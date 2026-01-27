package entity_test

import (
	"testing"
	"time"

	"transfer-shortener/domain/entity"
)

func TestNewShortURL_CreatesValidEntity(t *testing.T) {
	fullURL := "https://transfer.sixtyfive.me/abc12/file.txt"

	shortURL, err := entity.NewShortURL(fullURL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shortURL.FullURL != fullURL {
		t.Errorf("expected FullURL %s, got %s", fullURL, shortURL.FullURL)
	}
	if shortURL.Token == "" {
		t.Error("expected Token to be generated, got empty")
	}
	if len(shortURL.Token) != 4 {
		t.Errorf("expected Token length 4, got %d", len(shortURL.Token))
	}
	if shortURL.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNewShortURL_RejectsEmptyURL(t *testing.T) {
	_, err := entity.NewShortURL("")

	if err == nil {
		t.Error("expected error for empty URL, got nil")
	}
}

func TestNewShortURL_RejectsInvalidURL(t *testing.T) {
	_, err := entity.NewShortURL("not-a-valid-url")

	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestShortURL_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		createdAt time.Time
		ttl       time.Duration
		expected  bool
	}{
		{
			name:      "not expired when within TTL",
			createdAt: time.Now().Add(-1 * time.Hour),
			ttl:       24 * time.Hour,
			expected:  false,
		},
		{
			name:      "expired when past TTL",
			createdAt: time.Now().Add(-25 * time.Hour),
			ttl:       24 * time.Hour,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortURL := &entity.ShortURL{
				Token:     "test",
				FullURL:   "https://example.com/file.txt",
				CreatedAt: tt.createdAt,
			}

			if shortURL.IsExpired(tt.ttl) != tt.expected {
				t.Errorf("expected IsExpired=%v, got %v", tt.expected, !tt.expected)
			}
		})
	}
}
