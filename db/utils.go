package db

import (
	"errors"

	"github.com/jackc/pgx/v4"
)

func NoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
