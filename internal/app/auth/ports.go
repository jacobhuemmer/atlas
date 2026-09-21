package auth

import (
	"context"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Cred is one site's Basic email:token. Token must never appear in CLI or MCP output.
type Cred struct {
	Email string `json:"email"`
	Token string `json:"token,omitempty"`
}

// Blob is the keyring payload: credentials keyed by site alias.
type Blob struct {
	Sites map[string]Cred `json:"sites,omitempty"`
}

type Store interface {
	Get() (Blob, bool, error)
	Put(Blob) error
	Delete() error
}

// LoginFn is a human-terminal login. MCP never calls it.
type LoginFn func(ctx context.Context, site domain.Site, email, token string) (Cred, error)

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
	if len(b.Sites) == 0 {
		return s.Delete()
	}
	return s.Put(b)
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

func Login(ctx context.Context, s Store, fn LoginFn, site domain.Site, email, token string) (domain.Session, error) {
	if fn == nil {
		return domain.Session{}, domain.Auth("login not available")
	}
	c, err := fn(ctx, site, email, token)
	if err != nil {
		return domain.Session{}, err
	}
	if err := PutSite(s, site.Alias, c); err != nil {
		return domain.Session{}, err
	}
	return Status(s)
}

func FakeLogin() LoginFn {
	return func(_ context.Context, site domain.Site, email, token string) (Cred, error) {
		if email == "" {
			email = "fake@" + site.Hostname
		}
		if token == "" {
			token = "fake-token"
		}
		return Cred{Email: email, Token: token}, nil
	}
}
