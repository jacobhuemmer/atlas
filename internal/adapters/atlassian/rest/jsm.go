package rest

import (
	"context"
	"net/http"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// JSM is the live customer REST client (https://<hostname>/rest/servicedeskapi).
// Credentials are the jsm_customer keyring slot. Headers:
//
//	Authorization: Basic
//	X-Atlassian-Token: no-check
//	X-ExperimentalApi: opt-in
//
// Phase 7 tests use the in-memory seed; this stub is unused until wired.
type JSM struct{}

func (JSM) Desks(_ context.Context, _ string) (domain.DeskList, error) {
	return domain.DeskList{}, domain.Service("live JSM REST is not wired")
}

func (JSM) Types(_ context.Context, _, _ string) (domain.RequestTypeList, error) {
	return domain.RequestTypeList{}, domain.Service("live JSM REST is not wired")
}

func (JSM) List(_ context.Context, _, _ string) (domain.RequestList, error) {
	return domain.RequestList{}, domain.Service("live JSM REST is not wired")
}

func (JSM) Get(_ context.Context, _, _ string) (domain.CustomerRequest, error) {
	return domain.CustomerRequest{}, domain.Service("live JSM REST is not wired")
}

func (JSM) Create(_ context.Context, _ string, _ domain.CreateRequest, _ bool) (domain.CustomerRequest, error) {
	return domain.CustomerRequest{}, domain.Service("live JSM REST is not wired")
}

func (JSM) Comment(_ context.Context, _, _, _ string, _ bool) error {
	return domain.Service("live JSM REST is not wired")
}

func (JSM) Transition(_ context.Context, _, _, _ string, _ bool) (domain.CustomerRequest, error) {
	return domain.CustomerRequest{}, domain.Service("live JSM REST is not wired")
}

// MapJSMStatus maps JSM HTTP statuses to exit classes.
// 401/403 → auth, 404 → not_found, 429/5xx → service.
func MapJSMStatus(code int) error {
	switch code {
	case http.StatusUnauthorized, http.StatusForbidden:
		return domain.Auth("not authorized for this JSM site")
	case http.StatusNotFound:
		return domain.NotFound("request not found")
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return domain.Service("jsm service error")
	default:
		return domain.Service("jsm service error")
	}
}
