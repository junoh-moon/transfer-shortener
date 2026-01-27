package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"transfer-shortener/domain/entity"
	"transfer-shortener/domain/repository"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("short URL not found")

type Repository struct {
	db *sql.DB
}

var _ repository.URLRepository = (*Repository)(nil)

func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			token TEXT PRIMARY KEY,
			full_url TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_created_at ON urls(created_at);
	`)
	return err
}

func (r *Repository) Save(ctx context.Context, shortURL *entity.ShortURL) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO urls (token, full_url, created_at) VALUES (?, ?, ?)",
		shortURL.Token, shortURL.FullURL, shortURL.CreatedAt.Unix(),
	)
	return err
}

func (r *Repository) FindByToken(ctx context.Context, token string) (*entity.ShortURL, error) {
	var fullURL string
	var createdAt int64

	err := r.db.QueryRowContext(ctx,
		"SELECT full_url, created_at FROM urls WHERE token = ?",
		token,
	).Scan(&fullURL, &createdAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &entity.ShortURL{
		Token:     token,
		FullURL:   fullURL,
		CreatedAt: time.Unix(createdAt, 0),
	}, nil
}
