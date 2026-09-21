package keychain

import "github.com/masonhuemmer/atlas/internal/app/auth"

type Blob = auth.Blob

type Store interface {
	Get() (Blob, bool, error)
	Put(Blob) error
	Delete() error
}
