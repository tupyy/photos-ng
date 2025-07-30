package pg

import "github.com/jackc/pgx/v5"

type Writer struct {
	tx pgx.Tx
}
