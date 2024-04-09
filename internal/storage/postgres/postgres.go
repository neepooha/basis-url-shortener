package postgres

import (
	"context"
	"errors"
	"fmt"
	"url_shortener/internal/config"
	"url_shortener/internal/storage"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	const op = "storage.postgres.NewStorage"

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.User, cfg.Storage.Password, cfg.Storage.Dbname)

	db, err := pgxpool.New(context.Background(), psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt := `
	CREATE TABLE IF NOT EXISTS url(
		id SERIAL  PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias on url(alias);`

	_, err = db.Exec(context.Background(), stmt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) error {
	const op = "storage.postgres.SaveURL"

	stmt := `INSERT INTO url (url, alias) VALUES($1, $2)`
	_, err := s.db.Exec(context.Background(), stmt, urlToSave, alias)
	if err != nil {
		if IsDuplicatedKeyError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"
	stmt := `SELECT url FROM url WHERE alias = $1`
	var resURL string

	err := s.db.QueryRow(context.Background(), stmt, alias).Scan(&resURL)
	if err != nil {
		if IsNotFoundError(err) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	stmt := `DELETE FROM url WHERE alias = $1`
	res, err := s.db.Exec(context.Background(), stmt, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	affect := res.RowsAffected()
	if affect == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrAliasNotFound)
	}
	return nil
}

func IsDuplicatedKeyError(err error) bool {
	var perr *pgconn.PgError
	if errors.As(err, &perr) {
		return perr.Code == "23505" // error code of duplicate
	}
	return false
}

func IsNotFoundError(err error) bool {
	return err.Error() == "no rows in result set"
}
