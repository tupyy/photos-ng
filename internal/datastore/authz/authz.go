package authz

import (
	"github.com/authzed/authzed-go/v1"

	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/datastore"
)

type Datastore struct {
	pool   datastore.ConnPooler
	client *authzed.Client
}

func NewAuthzDatastore(pool datastore.ConnPooler, client *authzed.Client) *Datastore {
	return &Datastore{pool: pool, client: client}
}
