package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"transfer-shortener/domain/entity"
	"transfer-shortener/usecase"
)

func TestResolveShortURL_Success(t *testing.T) {
	expectedURL := "https://transfer.sixtyfive.me/abc12/file.txt"
	repo := &mockURLRepository{
		findByTokenFunc: func(ctx context.Context, token string) (*entity.ShortURL, error) {
			return &entity.ShortURL{
				Token:     token,
				FullURL:   expectedURL,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	uc := usecase.NewResolveShortURL(repo)

	result, err := uc.Execute(context.Background(), "abc1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != expectedURL {
		t.Errorf("expected %s, got %s", expectedURL, result)
	}
}

func TestResolveShortURL_NotFound(t *testing.T) {
	repo := &mockURLRepository{
		findByTokenFunc: func(ctx context.Context, token string) (*entity.ShortURL, error) {
			return nil, errors.New("not found")
		},
	}

	uc := usecase.NewResolveShortURL(repo)

	_, err := uc.Execute(context.Background(), "nonexistent")

	if err == nil {
		t.Error("expected error for non-existent token, got nil")
	}
}

func TestResolveShortURL_EmptyToken(t *testing.T) {
	repo := &mockURLRepository{}
	uc := usecase.NewResolveShortURL(repo)

	_, err := uc.Execute(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty token, got nil")
	}
}
