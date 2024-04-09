package storage

import "errors"

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrURLExists     = errors.New("url exists")
	ErrAliasNotFound = errors.New("alias not found")
)

type Urls struct {
	Id    uint64 `gorm:"primaryKey"`
	Alias string `gorm:""`
	Url   string `gorm:""`
}
