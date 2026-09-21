package auth

import (
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

type memStore struct {
	b  Blob
	ok bool
}

func (m *memStore) Get() (Blob, bool, error) { return m.b, m.ok, nil }
func (m *memStore) Put(b Blob) error         { m.b = b; m.ok = true; return nil }
func (m *memStore) Delete() error            { m.b = Blob{}; m.ok = false; return nil }

func TestStatusPerSiteUsable(t *testing.T) {
	s := &memStore{}
	st, err := Status(s)
	if err != nil || st.SignedIn {
		t.Fatal(err, st)
	}
	if err := PutSite(s, "sesamidevel", Cred{Email: "a@b.c", Token: "t"}); err != nil {
		t.Fatal(err)
	}
	st, err = Status(s)
	if err != nil || !st.SignedIn || !st.SessionUsable {
		t.Fatal(err, st)
	}
	var devel, garda bool
	for _, site := range st.Sites {
		if site.Alias == "sesamidevel" {
			devel = site.Usable
		}
		if site.Alias == "garda" {
			garda = site.Usable
		}
	}
	if !devel || garda {
		t.Fatal(st.Sites)
	}
	if err := LogoutSite(s, "sesamidevel"); err != nil {
		t.Fatal(err)
	}
	st, err = Status(s)
	if err != nil || st.SignedIn {
		t.Fatal(err, st)
	}
	_ = domain.Sites()
}
