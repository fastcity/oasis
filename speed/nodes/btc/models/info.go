package models

import "time"

type Info struct {
	Height int64
	Hash   string

	CreatedAt time.Time
	UpdatedAt time.Time
}
