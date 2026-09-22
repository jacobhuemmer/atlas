package auth

import (
	"strings"
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

func TestWorkspaceStatusAndTargetedLogoutPreserveOtherCredentials(t *testing.T) {
	s := &memStore{}
	workspace := domain.DefaultWorkspace()
	if err := PutSite(s, "sesamidevel", Cred{Email: "site@example.com", Token: "site-secret"}); err != nil {
		t.Fatal(err)
	}
	if err := PutWorkspace(s, workspace, Cred{Email: "bitbucket@example.com", Token: "workspace-secret"}); err != nil {
		t.Fatal(err)
	}

	st, err := Status(s)
	if err != nil || !st.SignedIn || !st.SessionUsable || !workspaceUsable(st, workspace) {
		t.Fatalf("status=%+v err=%v", st, err)
	}
	if err := LogoutSite(s, "sesamidevel"); err != nil {
		t.Fatal(err)
	}
	st, err = Status(s)
	if err != nil || !st.SignedIn || !workspaceUsable(st, workspace) {
		t.Fatalf("workspace lost after site logout: status=%+v err=%v", st, err)
	}
	if err := PutSite(s, "sesamidevel", Cred{Email: "site@example.com", Token: "site-secret"}); err != nil {
		t.Fatal(err)
	}
	if err := LogoutWorkspace(s, workspace); err != nil {
		t.Fatal(err)
	}
	st, err = Status(s)
	if err != nil || !st.SignedIn || workspaceUsable(st, workspace) {
		t.Fatalf("site lost after workspace logout: status=%+v err=%v", st, err)
	}
}

func TestWorkspaceCredMissingUsesWorkspaceLoginHint(t *testing.T) {
	s := &memStore{}
	if err := PutSite(s, "sesamidevel", Cred{Email: "site@example.com", Token: "site-secret"}); err != nil {
		t.Fatal(err)
	}
	_, err := WorkspaceCred(s, "other-workspace")
	if domain.ClassOf(err) != domain.ClassAuth || !strings.Contains(err.(*domain.Error).Hint, "--workspace other-workspace") {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func workspaceUsable(st domain.Session, slug string) bool {
	for _, workspace := range st.Workspaces {
		if workspace.Slug == slug {
			return workspace.Usable
		}
	}
	return false
}
