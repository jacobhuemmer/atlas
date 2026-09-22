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
	if err := f.Put(Blob{
		Sites:      map[string]auth.Cred{"sesamidevel": {Email: "a", Token: "site-token"}},
		Workspaces: map[string]auth.Cred{"workspace": {Email: "b", Token: "workspace-token"}},
	}); err != nil {
		t.Fatal(err)
	}
	b, ok, err := f.Get()
	if err != nil || !ok || b.Sites["sesamidevel"].Email != "a" || b.Workspaces["workspace"].Email != "b" {
		t.Fatalf("%+v", b)
	}
	delete(b.Sites, "sesamidevel")
	delete(b.Workspaces, "workspace")
	again, _, _ := f.Get()
	if _, ok := again.Sites["sesamidevel"]; !ok {
		t.Fatal("site map was not copied")
	}
	if _, ok := again.Workspaces["workspace"]; !ok {
		t.Fatal("workspace map was not copied")
	}
	if err := f.Delete(); err != nil {
		t.Fatal(err)
	}
	_, ok, _ = f.Get()
	if ok {
		t.Fatal("deleted")
	}
}
