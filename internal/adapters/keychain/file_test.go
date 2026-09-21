package keychain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/app/auth"
)

func TestFileStoreHoldsLargeBlob(t *testing.T) {
	p := filepath.Join(t.TempDir(), "session.json")
	s := &FileStore{Path: p}
	big := Blob{Sites: map[string]auth.Cred{"sesamidevel": {Email: "user@example.com", Token: strings.Repeat("a", 8000)}}}
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
	if err != nil || !ok || len(got.Sites["sesamidevel"].Token) != 8000 {
		t.Fatalf("got %+v ok=%v err=%v", got, ok, err)
	}
}
