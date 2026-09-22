package keychain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/app/auth"
)

func TestFileStoreHoldsLargeBlob(t *testing.T) {
	p := filepath.Join(t.TempDir(), "session.json")
	s := &FileStore{Path: p}
	big := Blob{
		Sites:      map[string]auth.Cred{"sesamidevel": {Email: "user@example.com", Token: strings.Repeat("a", 8000)}},
		Workspaces: map[string]auth.Cred{"workspace": {Email: "user@example.com", Token: "workspace-token"}},
	}
	if err := s.Put(big); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o", st.Mode().Perm())
	}
	got, ok, err := s.Get()
	if err != nil || !ok || len(got.Sites["sesamidevel"].Token) != 8000 || got.Workspaces["workspace"].Token != "workspace-token" {
		t.Fatalf("got %+v ok=%v err=%v", got, ok, err)
	}
}

func TestFileStoreLoadsLegacySiteOnlyBlob(t *testing.T) {
	p := filepath.Join(t.TempDir(), "session.json")
	legacy := struct {
		Sites map[string]auth.Cred `json:"sites"`
	}{Sites: map[string]auth.Cred{"dev": {Email: "user@example.com", Token: "site-token"}}}
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	got, ok, err := (&FileStore{Path: p}).Get()
	if err != nil || !ok || got.Sites["dev"].Token != "site-token" || len(got.Workspaces) != 0 {
		t.Fatalf("got %+v ok=%v err=%v", got, ok, err)
	}
}
