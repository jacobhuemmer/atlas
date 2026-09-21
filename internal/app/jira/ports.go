package jira

import (
	"context"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Store is licensed Jira on one hostname. Fake memory and live REST both implement it.
type Store interface {
	Get(ctx context.Context, hostname, key string) (domain.Issue, error)
	Search(ctx context.Context, hostname, jql string) (domain.SearchResult, error)
}
