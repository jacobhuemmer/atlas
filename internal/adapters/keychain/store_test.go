package keychain

import (
	"testing"

	"github.com/masonhuemmer/atlas/internal/app/auth"
)

func TestFakeDoesNotUseKeychain(t *testing.T) {
	f := &Fake{}
	_, ok, err := f.Get()
	if err != nil || ok {
		t.Fatal("empty")
	}
	if err := f.Put(Blob{Sites: map[string]auth.Cred{"sesamidevel": {Email: "a", Token: "t"}}}); err != nil {
		t.Fatal(err)
	}
	b, ok, err := f.Get()
	if err != nil || !ok || b.Sites["sesamidevel"].Email != "a" {
		t.Fatalf("%+v", b)
	}
	if err := f.Delete(); err != nil {
		t.Fatal(err)
	}
	_, ok, _ = f.Get()
	if ok {
		t.Fatal("deleted")
	}
}
