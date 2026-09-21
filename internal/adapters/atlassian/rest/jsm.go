package rest

import (
	"context"
	"net/http"
	"strings"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func jsmHeaders() http.Header {
	h := make(http.Header)
	h.Set("X-Atlassian-Token", "no-check")
	h.Set("X-ExperimentalApi", "opt-in")
	return h
}

// JSM is the live customer REST client (https://<hostname>/rest/servicedeskapi).
type JSM struct {
	*Client
}

func (j JSM) Desks(ctx context.Context, hostname string) (domain.DeskList, error) {
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.DeskList{}, err
	}
	u := joinURL(j.origin(hostname), "/rest/servicedeskapi/servicedesk")
	code, body, err := j.doJSON(ctx, http.MethodGet, u, cred, jsmHeaders(), nil)
	if err != nil {
		return domain.DeskList{}, err
	}
	if code != http.StatusOK {
		return domain.DeskList{}, MapJSMStatus(code)
	}
	var items []domain.ServiceDesk
	for _, raw := range asList(mustJSON(body)["values"]) {
		m := asMap(raw)
		items = append(items, domain.ServiceDesk{
			ID:   str(m, "id"),
			Key:  str(m, "projectKey"),
			Name: str(m, "projectName"),
		})
	}
	if items == nil {
		items = []domain.ServiceDesk{}
	}
	return domain.DeskList{Site: hostname, Count: len(items), Items: items}, nil
}

func (j JSM) Types(ctx context.Context, hostname, deskID string) (domain.RequestTypeList, error) {
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.RequestTypeList{}, err
	}
	u := joinURL(j.origin(hostname), "/rest/servicedeskapi/servicedesk/"+q(deskID)+"/requesttype")
	code, body, err := j.doJSON(ctx, http.MethodGet, u, cred, jsmHeaders(), nil)
	if err != nil {
		return domain.RequestTypeList{}, err
	}
	if code != http.StatusOK {
		return domain.RequestTypeList{}, MapJSMStatus(code)
	}
	var items []domain.RequestType
	for _, raw := range asList(mustJSON(body)["values"]) {
		m := asMap(raw)
		items = append(items, domain.RequestType{ID: str(m, "id"), Name: str(m, "name"), DeskID: deskID})
	}
	if items == nil {
		items = []domain.RequestType{}
	}
	return domain.RequestTypeList{DeskID: deskID, Count: len(items), Items: items}, nil
}

func (j JSM) List(ctx context.Context, hostname, status string) (domain.RequestList, error) {
	norm, err := domain.NormalizeJSMStatus(status)
	if err != nil {
		return domain.RequestList{}, err
	}
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.RequestList{}, err
	}
	u := joinURL(j.origin(hostname), "/rest/servicedeskapi/request?requestStatus="+q(domain.JSMRequestStatusQuery(norm)))
	code, body, err := j.doJSON(ctx, http.MethodGet, u, cred, jsmHeaders(), nil)
	if err != nil {
		return domain.RequestList{}, err
	}
	if code != http.StatusOK {
		return domain.RequestList{}, MapJSMStatus(code)
	}
	var items []domain.CustomerRequest
	for _, raw := range asList(mustJSON(body)["values"]) {
		items = append(items, requestFromREST(hostname, asMap(raw)))
	}
	if items == nil {
		items = []domain.CustomerRequest{}
	}
	return domain.RequestList{Site: hostname, Status: norm, Count: len(items), Items: items}, nil
}

func (j JSM) Get(ctx context.Context, hostname, key string) (domain.CustomerRequest, error) {
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.CustomerRequest{}, err
	}
	u := joinURL(j.origin(hostname), "/rest/servicedeskapi/request/"+q(strings.ToUpper(strings.TrimSpace(key)))+"?expand=comment")
	code, body, err := j.doJSON(ctx, http.MethodGet, u, cred, jsmHeaders(), nil)
	if err != nil {
		return domain.CustomerRequest{}, err
	}
	if code != http.StatusOK {
		return domain.CustomerRequest{}, MapJSMStatus(code)
	}
	return requestFromREST(hostname, mustJSON(body)), nil
}

func (j JSM) Create(ctx context.Context, hostname string, in domain.CreateRequest, dryRun bool) (domain.CustomerRequest, error) {
	preview := domain.CustomerRequest{
		Site: hostname, DeskID: in.DeskID, TypeID: in.TypeID,
		Summary: in.Summary, Description: in.Description, Status: "OPEN",
	}
	if dryRun {
		return preview, nil
	}
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.CustomerRequest{}, err
	}
	payload := map[string]any{
		"serviceDeskId":      in.DeskID,
		"requestTypeId":      in.TypeID,
		"requestFieldValues": map[string]any{"summary": in.Summary, "description": in.Description},
	}
	u := joinURL(j.origin(hostname), "/rest/servicedeskapi/request")
	code, body, err := j.doJSON(ctx, http.MethodPost, u, cred, jsmHeaders(), payload)
	if err != nil {
		return domain.CustomerRequest{}, err
	}
	if code != http.StatusCreated && code != http.StatusOK {
		return domain.CustomerRequest{}, MapJSMStatus(code)
	}
	return requestFromREST(hostname, mustJSON(body)), nil
}

func (j JSM) Comment(ctx context.Context, hostname, key, body string, dryRun bool) error {
	if dryRun {
		return nil
	}
	cred, err := j.credForHost(hostname)
	if err != nil {
		return err
	}
	u := joinURL(j.origin(hostname), "/rest/servicedeskapi/request/"+q(strings.ToUpper(strings.TrimSpace(key)))+"/comment")
	code, _, err := j.doJSON(ctx, http.MethodPost, u, cred, jsmHeaders(), map[string]any{"body": body, "public": true})
	if err != nil {
		return err
	}
	if code != http.StatusCreated && code != http.StatusOK {
		return MapJSMStatus(code)
	}
	return nil
}

func (j JSM) Transition(ctx context.Context, hostname, key, id string, dryRun bool) (domain.CustomerRequest, error) {
	if dryRun {
		return j.Get(ctx, hostname, key)
	}
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.CustomerRequest{}, err
	}
	u := joinURL(j.origin(hostname), "/rest/servicedeskapi/request/"+q(strings.ToUpper(strings.TrimSpace(key)))+"/transition")
	code, _, err := j.doJSON(ctx, http.MethodPost, u, cred, jsmHeaders(), map[string]any{"id": id})
	if err != nil {
		return domain.CustomerRequest{}, err
	}
	if code != http.StatusNoContent && code != http.StatusOK {
		return domain.CustomerRequest{}, MapJSMStatus(code)
	}
	return j.Get(ctx, hostname, key)
}

func requestFromREST(hostname string, m map[string]any) domain.CustomerRequest {
	key := str(m, "issueKey")
	if key == "" {
		key = str(m, "issueId")
	}
	desk := str(m, "serviceDeskId")
	if desk == "" {
		desk = str(asMap(m["serviceDesk"]), "id")
	}
	st := asMap(m["currentStatus"])
	req := domain.CustomerRequest{
		Key:            key,
		Site:           hostname,
		DeskID:         desk,
		TypeID:         str(m, "requestTypeId"),
		Summary:        firstField(m, "summary"),
		Description:    firstField(m, "description"),
		Status:         str(st, "status"),
		StatusCategory: str(st, "statusCategory"),
		PortalURL:      domain.PortalURL(hostname, desk, key),
	}
	if req.Status == "" {
		req.Status = str(m, "currentStatus")
	}
	if c := asMap(m["comment"]); c != nil {
		for _, raw := range asList(c["values"]) {
			cm := asMap(raw)
			pub, _ := cm["public"].(bool)
			req.Comments = append(req.Comments, domain.Comment{Body: str(cm, "body"), Public: pub})
		}
	}
	return req
}

func firstField(m map[string]any, name string) string {
	if s := str(m, name); s != "" {
		return s
	}
	for _, raw := range asList(m["requestFieldValues"]) {
		f := asMap(raw)
		if str(f, "fieldId") == name || str(f, "label") == name {
			if s, ok := f["value"].(string); ok {
				return s
			}
		}
	}
	return ""
}

// MapJSMStatus maps JSM HTTP statuses to exit classes.
func MapJSMStatus(code int) error {
	return classify(code, "jsm", "request not found")
}
