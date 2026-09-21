package rest

import (
	"context"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Confluence is the live Cloud REST client (https://<hostname>/wiki/api/v2).
// Phase 5 tests use the in-memory seed; this stub is unused until wired.
type Confluence struct{}

func (Confluence) Get(_ context.Context, _, _ string) (domain.Page, error) {
	return domain.Page{}, domain.Service("live Confluence REST is not wired")
}

func (Confluence) Search(_ context.Context, _, _ string) (domain.PageSearchResult, error) {
	return domain.PageSearchResult{}, domain.Service("live Confluence REST is not wired")
}

func (Confluence) Create(_ context.Context, _ string, _ domain.CreatePage, _ bool) (domain.Page, error) {
	return domain.Page{}, domain.Service("live Confluence REST is not wired")
}

func (Confluence) Update(_ context.Context, _, _, _ string, _ bool) (domain.Page, error) {
	return domain.Page{}, domain.Service("live Confluence REST is not wired")
}
