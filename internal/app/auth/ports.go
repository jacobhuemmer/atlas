package auth

import (
	"context"
	"sort"
	"strings"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Cred is one site's Basic email:token. Token must never appear in CLI or MCP output.
type Cred struct {
	Email string `json:"email"`
	Token string `json:"token,omitempty"`
}

// Blob is the keyring payload: credentials keyed by site alias or Bitbucket workspace slug.
type Blob struct {
	Sites      map[string]Cred `json:"sites,omitempty"`
	Workspaces map[string]Cred `json:"workspaces,omitempty"`
}

type Store interface {
	Get() (Blob, bool, error)
	Put(Blob) error
	Delete() error
}

// LoginFn is a human-terminal login. MCP never calls it.
type LoginFn func(ctx context.Context, email, token string) (Cred, error)

func Status(s Store) (domain.Session, error) {
	if s == nil {
		return domain.SignedOut(), nil
	}
	b, ok, err := s.Get()
	if err != nil {
		return domain.Session{}, err
	}
	if !ok {
		return domain.SignedOut(), nil
	}
	sess := domain.SignedOut()
	anyUsable := false
	for i, st := range sess.Sites {
		c, present := b.Sites[st.Alias]
		usable := present && c.Email != "" && c.Token != ""
		sess.Sites[i].Usable = usable
		if usable {
			anyUsable = true
		}
	}
	workspaceKeys := make(map[string]struct{}, len(sess.Workspaces)+len(b.Workspaces))
	for _, workspace := range sess.Workspaces {
		workspaceKeys[workspace.Slug] = struct{}{}
	}
	for workspace := range b.Workspaces {
		workspaceKeys[workspace] = struct{}{}
	}
	keys := make([]string, 0, len(workspaceKeys))
	for workspace := range workspaceKeys {
		if strings.TrimSpace(workspace) != "" {
			keys = append(keys, workspace)
		}
	}
	sort.Strings(keys)
	sess.Workspaces = make([]domain.WorkspaceStatus, 0, len(keys))
	for _, workspace := range keys {
		c, present := b.Workspaces[workspace]
		usable := present && c.Email != "" && c.Token != ""
		sess.Workspaces = append(sess.Workspaces, domain.WorkspaceStatus{Slug: workspace, Usable: usable})
		if usable {
			anyUsable = true
		}
	}
	sess.SignedIn = anyUsable
	sess.SessionUsable = anyUsable
	return sess, nil
}

func Logout(s Store) error {
	if s == nil {
		return nil
	}
	return s.Delete()
}

func LogoutSite(s Store, alias string) error {
	if s == nil {
		return nil
	}
	b, ok, err := s.Get()
	if err != nil {
		return err
	}
	if !ok || b.Sites == nil {
		return nil
	}
	delete(b.Sites, alias)
	return persistOrDelete(s, b)
}

func LogoutWorkspace(s Store, workspace string) error {
	if s == nil {
		return nil
	}
	b, ok, err := s.Get()
	if err != nil {
		return err
	}
	if !ok || b.Workspaces == nil {
		return nil
	}
	delete(b.Workspaces, strings.TrimSpace(workspace))
	return persistOrDelete(s, b)
}

func PutSite(s Store, alias string, c Cred) error {
	if s == nil {
		return domain.Auth("login not available")
	}
	b, _, err := s.Get()
	if err != nil {
		return err
	}
	if b.Sites == nil {
		b.Sites = map[string]Cred{}
	}
	b.Sites[alias] = c
	return s.Put(b)
}

func PutWorkspace(s Store, workspace string, c Cred) error {
	if s == nil {
		return domain.Auth("login not available")
	}
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return domain.Usage("workspace is required")
	}
	b, _, err := s.Get()
	if err != nil {
		return err
	}
	if b.Workspaces == nil {
		b.Workspaces = map[string]Cred{}
	}
	b.Workspaces[workspace] = c
	return s.Put(b)
}

func WorkspaceCred(s Store, workspace string) (Cred, error) {
	workspace = strings.TrimSpace(workspace)
	hint := "atlas auth login --workspace " + workspace + " --email EMAIL --token TOKEN"
	if s == nil {
		return Cred{}, domain.Auth("not signed in for workspace " + workspace).WithHint(hint)
	}
	b, ok, err := s.Get()
	if err != nil {
		return Cred{}, err
	}
	cred, present := b.Workspaces[workspace]
	if !ok || !present || cred.Email == "" || cred.Token == "" {
		return Cred{}, domain.Auth("not signed in for workspace " + workspace).WithHint(hint)
	}
	return cred, nil
}

func Login(ctx context.Context, s Store, fn LoginFn, site domain.Site, email, token string) (domain.Session, error) {
	if fn == nil {
		return domain.Session{}, domain.Auth("login not available")
	}
	c, err := fn(ctx, email, token)
	if err != nil {
		return domain.Session{}, err
	}
	if err := PutSite(s, site.Alias, c); err != nil {
		return domain.Session{}, err
	}
	return Status(s)
}

func LoginWorkspace(ctx context.Context, s Store, fn LoginFn, workspace, email, token string) (domain.Session, error) {
	if fn == nil {
		return domain.Session{}, domain.Auth("login not available")
	}
	c, err := fn(ctx, email, token)
	if err != nil {
		return domain.Session{}, err
	}
	if err := PutWorkspace(s, workspace, c); err != nil {
		return domain.Session{}, err
	}
	return Status(s)
}

func FakeLogin() LoginFn {
	return func(_ context.Context, email, token string) (Cred, error) {
		if email == "" {
			email = "fake@example.invalid"
		}
		if token == "" {
			token = "fake-token"
		}
		return Cred{Email: email, Token: token}, nil
	}
}

func persistOrDelete(s Store, b Blob) error {
	if len(b.Sites) == 0 && len(b.Workspaces) == 0 {
		return s.Delete()
	}
	return s.Put(b)
}
