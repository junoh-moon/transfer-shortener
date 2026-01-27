package usecase

import (
	"context"
	"errors"

	"transfer-shortener/domain/repository"
)

var ErrEmptyToken = errors.New("token cannot be empty")

type ResolveShortURL struct {
	repo repository.URLRepository
}

func NewResolveShortURL(repo repository.URLRepository) *ResolveShortURL {
	return &ResolveShortURL{repo: repo}
}

func (uc *ResolveShortURL) Execute(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrEmptyToken
	}

	shortURL, err := uc.repo.FindByToken(ctx, token)
	if err != nil {
		return "", err
	}

	return shortURL.FullURL, nil
}
