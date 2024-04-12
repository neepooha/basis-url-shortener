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
	return &Storage{db: db}, nil
}

func (s *Storage) CloseStorage() {
	s.db.Close()
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) error {
	const op = "storage.postgres.SaveURL"

	stmt := `INSERT INTO urls (url, alias) VALUES($1, $2)`
	_, err := s.db.Exec(ctx, stmt, urlToSave, alias)
	if err != nil {
		if IsDuplicatedKeyError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	stmt := `SELECT url FROM urls WHERE alias = $1`
	var resURL string
	err := s.db.QueryRow(ctx, stmt, alias).Scan(&resURL)
	if err != nil {
		if IsNotFoundError(err) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) error {
	const op = "storage.postgres.DeleteURL"

	stmt := `DELETE FROM urls WHERE alias = $1`
	res, err := s.db.Exec(ctx, stmt, alias)
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
