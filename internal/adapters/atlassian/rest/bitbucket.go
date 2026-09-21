package rest

import (
	"context"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Bitbucket is the live Cloud REST client (https://api.bitbucket.org/2.0).
// Phase 6 tests use the in-memory seed; this stub is unused until wired.
type Bitbucket struct{}

func (Bitbucket) Get(_ context.Context, _, _ string, _ int) (domain.PullRequest, error) {
	return domain.PullRequest{}, domain.Service("live Bitbucket REST is not wired")
}

func (Bitbucket) List(_ context.Context, _, _ string) (domain.PullRequestList, error) {
	return domain.PullRequestList{}, domain.Service("live Bitbucket REST is not wired")
}

func (Bitbucket) Create(_ context.Context, _, _ string, _ domain.CreatePullRequest, _ bool) (domain.PullRequest, error) {
	return domain.PullRequest{}, domain.Service("live Bitbucket REST is not wired")
}

func (Bitbucket) Comment(_ context.Context, _, _ string, _ int, _ string, _ bool) error {
	return domain.Service("live Bitbucket REST is not wired")
}

func (Bitbucket) Merge(_ context.Context, _, _ string, _ int, _ bool) (domain.PullRequest, error) {
	return domain.PullRequest{}, domain.Service("live Bitbucket REST is not wired")
}

func (Bitbucket) Diff(_ context.Context, _, _ string, _ int) (domain.PullRequestDiff, error) {
	return domain.PullRequestDiff{}, domain.Service("live Bitbucket REST is not wired")
}
