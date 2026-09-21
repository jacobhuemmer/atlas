package keychain

import (
	"sync"

	"github.com/masonhuemmer/atlas/internal/app/auth"
)

type Fake struct {
	mu sync.Mutex
	v  *Blob
}

func (f *Fake) Get() (Blob, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.v == nil {
		return Blob{}, false, nil
	}
	return copyBlob(*f.v), true, nil
}

func (f *Fake) Put(b Blob) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := copyBlob(b)
	f.v = &cp
	return nil
}

func (f *Fake) Delete() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.v = nil
	return nil
}

func copyBlob(b Blob) Blob {
	out := Blob{}
	if b.Sites == nil {
		return out
	}
	out.Sites = make(map[string]auth.Cred, len(b.Sites))
	for k, v := range b.Sites {
		out.Sites[k] = v
	}
	return out
}
