package usecase

import (
	"context"

	"transfer-shortener/domain/entity"
	"transfer-shortener/domain/repository"
)

type CreateShortURL struct {
	repo repository.URLRepository
}

func NewCreateShortURL(repo repository.URLRepository) *CreateShortURL {
	return &CreateShortURL{repo: repo}
}

func (uc *CreateShortURL) Execute(ctx context.Context, fullURL string) (*entity.ShortURL, error) {
	shortURL, err := entity.NewShortURL(fullURL)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, shortURL); err != nil {
		return nil, err
	}

	return shortURL, nil
}
