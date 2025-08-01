package pg

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"github.com/jackc/pgx/v5"
)

type Writer struct {
	tx pgx.Tx
}

func (w *Writer) WriteAlbum(ctx context.Context, album entity.Album) error {
	return nil
}
