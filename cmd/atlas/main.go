package main

import (
	"context"
	"os"

	"github.com/masonhuemmer/atlas/internal/adapters/atlassian"
	"github.com/masonhuemmer/atlas/internal/adapters/cli"
	"github.com/masonhuemmer/atlas/internal/adapters/keychain"
	"github.com/masonhuemmer/atlas/internal/app/auth"
	"github.com/masonhuemmer/atlas/internal/config"
	"github.com/masonhuemmer/atlas/internal/domain"
)

func main() {
	cfg, _ := config.Load()
	d := cli.Deps{Config: cfg}
	if os.Getenv("ATLAS_FAKE") == "1" {
		mem := atlassian.Seed()
		d.Jira = atlassian.JiraAPI{Memory: mem}
		d.Confluence = atlassian.ConfluenceAPI{Memory: mem}
		d.Store = &keychain.FileStore{Path: keychain.FakeSessionPath()}
		d.Login = auth.FakeLogin()
	} else {
		d.Store = &keychain.Fallback{
			Primary:   keychain.Keyring{},
			Secondary: &keychain.FileStore{Path: keychain.LiveSessionPath()},
		}
		d.Login = storeLogin
	}
	os.Exit(cli.Run(os.Args, d))
}

func storeLogin(_ context.Context, _ domain.Site, email, token string) (auth.Cred, error) {
	if email == "" || token == "" {
		return auth.Cred{}, domain.Usage("email and token are required").WithHint("atlas auth login --site ALIAS --email EMAIL --token TOKEN")
	}
	return auth.Cred{Email: email, Token: token}, nil
}
