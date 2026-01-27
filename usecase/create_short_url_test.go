package usecase_test

import (
	"context"
	"errors"
	"testing"

	"transfer-shortener/domain/entity"
	"transfer-shortener/usecase"
)

type mockURLRepository struct {
	saveFunc       func(ctx context.Context, shortURL *entity.ShortURL) error
	findByTokenFunc func(ctx context.Context, token string) (*entity.ShortURL, error)
}

func (m *mockURLRepository) Save(ctx context.Context, shortURL *entity.ShortURL) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, shortURL)
	}
	return nil
}

func (m *mockURLRepository) FindByToken(ctx context.Context, token string) (*entity.ShortURL, error) {
	if m.findByTokenFunc != nil {
		return m.findByTokenFunc(ctx, token)
	}
	return nil, errors.New("not found")
}

func TestCreateShortURL_Success(t *testing.T) {
	var savedURL *entity.ShortURL
	repo := &mockURLRepository{
		saveFunc: func(ctx context.Context, shortURL *entity.ShortURL) error {
			savedURL = shortURL
			return nil
		},
	}

	uc := usecase.NewCreateShortURL(repo)
	fullURL := "https://transfer.sixtyfive.me/abc12/file.txt"

	result, err := uc.Execute(context.Background(), fullURL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.FullURL != fullURL {
		t.Errorf("expected FullURL %s, got %s", fullURL, result.FullURL)
	}
	if savedURL == nil {
		t.Error("expected Save to be called")
	}
}

func TestCreateShortURL_InvalidURL(t *testing.T) {
	repo := &mockURLRepository{}
	uc := usecase.NewCreateShortURL(repo)

	_, err := uc.Execute(context.Background(), "invalid-url")

	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestCreateShortURL_RepositoryError(t *testing.T) {
	repo := &mockURLRepository{
		saveFunc: func(ctx context.Context, shortURL *entity.ShortURL) error {
			return errors.New("database error")
		},
	}
	uc := usecase.NewCreateShortURL(repo)

	_, err := uc.Execute(context.Background(), "https://example.com/file.txt")

	if err == nil {
		t.Error("expected error when repository fails, got nil")
	}
}
