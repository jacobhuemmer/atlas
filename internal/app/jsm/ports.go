package jsm

import (
	"context"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Store is JSM customer REST on a jsm_customer site.
// Fake memory and live REST both implement it. Never Jira search.
// Comments are public: true only. Customers cannot raiseOnBehalfOf.
type Store interface {
	Desks(ctx context.Context, hostname string) (domain.DeskList, error)
	Types(ctx context.Context, hostname, deskID string) (domain.RequestTypeList, error)
	List(ctx context.Context, hostname, status string) (domain.RequestList, error)
	Get(ctx context.Context, hostname, key string) (domain.CustomerRequest, error)
	Create(ctx context.Context, hostname string, in domain.CreateRequest, dryRun bool) (domain.CustomerRequest, error)
	Comment(ctx context.Context, hostname, key, body string, dryRun bool) error
	Transition(ctx context.Context, hostname, key, id string, dryRun bool) (domain.CustomerRequest, error)
}
