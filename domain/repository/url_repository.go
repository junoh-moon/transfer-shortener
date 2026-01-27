package repository

import (
	"context"

	"transfer-shortener/domain/entity"
)

type URLRepository interface {
	Save(ctx context.Context, shortURL *entity.ShortURL) error
	FindByToken(ctx context.Context, token string) (*entity.ShortURL, error)
}
