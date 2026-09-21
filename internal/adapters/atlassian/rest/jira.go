package rest

import (
	"context"
	"net/http"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Jira is the live Cloud REST client (https://<hostname>/rest/api/3).
// Phase 2 tests use the in-memory seed; this stub is unused until wired.
type Jira struct{}

func (Jira) Get(_ context.Context, _, _ string) (domain.Issue, error) {
	return domain.Issue{}, domain.Service("live Jira REST is not wired")
}

func (Jira) Search(_ context.Context, _, _ string) (domain.SearchResult, error) {
	return domain.SearchResult{}, domain.Service("live Jira REST is not wired")
}

func (Jira) Create(_ context.Context, _ string, _ domain.CreateIssue, _ bool) (domain.Issue, error) {
	return domain.Issue{}, domain.Service("live Jira REST is not wired")
}

func (Jira) Edit(_ context.Context, _, _ string, _ map[string]any, _ bool) (domain.Issue, error) {
	return domain.Issue{}, domain.Service("live Jira REST is not wired")
}

func (Jira) Comment(_ context.Context, _, _, _ string, _ bool) error {
	return domain.Service("live Jira REST is not wired")
}

func (Jira) Transition(_ context.Context, _, _, _ string, _ bool) (domain.Issue, error) {
	return domain.Issue{}, domain.Service("live Jira REST is not wired")
}

func (Jira) Link(_ context.Context, _, _, _, _ string, _ bool) error {
	return domain.Service("live Jira REST is not wired")
}

// MapStatus maps Jira HTTP statuses to exit classes.
// 401/403 → auth, 404 → not_found, 429/5xx → service.
func MapStatus(code int) error {
	switch code {
	case http.StatusUnauthorized, http.StatusForbidden:
		return domain.Auth("not authorized for this Jira site")
	case http.StatusNotFound:
		return domain.NotFound("issue not found")
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return domain.Service("jira service error")
	default:
		return domain.Service("jira service error")
	}
}
