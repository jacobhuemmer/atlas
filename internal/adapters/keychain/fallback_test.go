package keychain

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/app/auth"
)

type stubStore struct {
	blob    Blob
	ok      bool
	putErr  error
	deleted bool
}

func (s *stubStore) Get() (Blob, bool, error) { return s.blob, s.ok, nil }
func (s *stubStore) Put(b Blob) error {
	if s.putErr != nil {
		return s.putErr
	}
	s.blob = b
	s.ok = true
	return nil
}
func (s *stubStore) Delete() error { s.deleted = true; s.ok = false; s.blob = Blob{}; return nil }

func TestFallbackPutUsesFileWhenKeychainTooBig(t *testing.T) {
	primary := &stubStore{putErr: errors.New("data passed to Set was too big")}
	path := filepath.Join(t.TempDir(), "session.json")
	f := &Fallback{Primary: primary, Secondary: &FileStore{Path: path}}
	big := Blob{Sites: map[string]auth.Cred{"sesamidevel": {Email: "user@example.com", Token: strings.Repeat("a", 8000)}}}
	if err := f.Put(big); err != nil {
		t.Fatal(err)
	}
	got, ok, err := f.Get()
	if err != nil || !ok || len(got.Sites["sesamidevel"].Token) != 8000 {
		t.Fatalf("got ok=%v err=%v", ok, err)
	}
	st, err := os.Stat(path)
	if err != nil || st.Mode().Perm() != 0o600 {
		t.Fatalf("session file: %v", err)
	}
}

func TestFallbackPutToFileDeletesStalePrimary(t *testing.T) {
	primary := &stubStore{
		blob:   Blob{Sites: map[string]auth.Cred{"dev": {Email: "old@example.com", Token: "old-token"}}},
		ok:     true,
		putErr: errors.New("data passed to Set was too big"),
	}
	path := filepath.Join(t.TempDir(), "session.json")
	f := &Fallback{Primary: primary, Secondary: &FileStore{Path: path}}
	want := Blob{
		Sites:      map[string]auth.Cred{"dev": {Email: "site@example.com", Token: "site-token"}},
		Workspaces: map[string]auth.Cred{"workspace": {Email: "bitbucket@example.com", Token: strings.Repeat("b", 8000)}},
	}
	if err := f.Put(want); err != nil {
		t.Fatal(err)
	}
	if !primary.deleted {
		t.Fatal("stale primary was not deleted")
	}
	got, ok, err := f.Get()
	if err != nil || !ok || got.Workspaces["workspace"].Email != "bitbucket@example.com" {
		t.Fatalf("got=%+v ok=%v err=%v", got, ok, err)
	}
}

func TestFallbackPutToPrimaryDeletesStaleSecondary(t *testing.T) {
	primary := &stubStore{}
	path := filepath.Join(t.TempDir(), "session.json")
	secondary := &FileStore{Path: path}
	if err := secondary.Put(Blob{Workspaces: map[string]auth.Cred{"old": {Email: "old@example.com", Token: "old-token"}}}); err != nil {
		t.Fatal(err)
	}
	f := &Fallback{Primary: primary, Secondary: secondary}
	if err := f.Put(Blob{Sites: map[string]auth.Cred{"dev": {Email: "site@example.com", Token: "site-token"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("stale secondary still present: %v", err)
	}
	got, ok, err := f.Get()
	if err != nil || !ok || got.Sites["dev"].Email != "site@example.com" || len(got.Workspaces) != 0 {
		t.Fatalf("got=%+v ok=%v err=%v", got, ok, err)
	}
}

func TestFallbackPutReportsStaleSecondaryCleanupFailure(t *testing.T) {
	primary := &stubStore{}
	path := filepath.Join(t.TempDir(), "session.json")
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "child"), []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}
	f := &Fallback{Primary: primary, Secondary: &FileStore{Path: path}}
	err := f.Put(Blob{Sites: map[string]auth.Cred{"dev": {Email: "site@example.com", Token: "site-token"}}})
	if err == nil {
		t.Fatal("expected stale fallback cleanup error")
	}
}

func TestFallbackDeleteClearsBoth(t *testing.T) {
	primary := &stubStore{blob: Blob{Sites: map[string]auth.Cred{"sesamidevel": {Email: "k"}}}, ok: true}
	path := filepath.Join(t.TempDir(), "session.json")
	sec := &FileStore{Path: path}
	if err := sec.Put(Blob{Sites: map[string]auth.Cred{"garda": {Email: "f", Token: "t"}}}); err != nil {
		t.Fatal(err)
	}
	f := &Fallback{Primary: primary, Secondary: sec}
	if err := f.Delete(); err != nil {
		t.Fatal(err)
	}
	if !primary.deleted {
		t.Fatal("primary not deleted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("secondary still present: %v", err)
	}
}
