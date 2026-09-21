package rest

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Confluence is the live Cloud REST client (https://<hostname>/wiki/api/v2).
type Confluence struct {
	*Client
}

func (c Confluence) Get(ctx context.Context, hostname, pageID string) (domain.Page, error) {
	cred, err := c.credForHost(hostname)
	if err != nil {
		return domain.Page{}, err
	}
	u := joinURL(c.origin(hostname), "/wiki/api/v2/pages/"+q(pageID)+"?body-format=storage")
	code, body, err := c.doJSON(ctx, http.MethodGet, u, cred, nil, nil)
	if err != nil {
		return domain.Page{}, err
	}
	if code != http.StatusOK {
		return domain.Page{}, MapConfluenceStatus(code)
	}
	return pageFromREST(hostname, mustJSON(body)), nil
}

func (c Confluence) Search(ctx context.Context, hostname, cql string) (domain.PageSearchResult, error) {
	cred, err := c.credForHost(hostname)
	if err != nil {
		return domain.PageSearchResult{}, err
	}
	u := joinURL(c.origin(hostname), "/wiki/rest/api/content/search?cql="+q(cql)+"&limit=25&expand=body.storage,space")
	code, body, err := c.doJSON(ctx, http.MethodGet, u, cred, nil, nil)
	if err != nil {
		return domain.PageSearchResult{}, err
	}
	if code != http.StatusOK {
		return domain.PageSearchResult{}, MapConfluenceStatus(code)
	}
	m := mustJSON(body)
	var items []domain.Page
	for _, raw := range asList(m["results"]) {
		items = append(items, pageFromV1(hostname, asMap(raw)))
	}
	if items == nil {
		items = []domain.Page{}
	}
	return domain.PageSearchResult{CQL: cql, Site: hostname, Count: len(items), Items: items}, nil
}

func (c Confluence) Create(ctx context.Context, hostname string, in domain.CreatePage, dryRun bool) (domain.Page, error) {
	preview := domain.Page{
		Site: hostname, Space: in.Space, Title: in.Title, Body: in.Body,
		ContentFormat: domain.DefaultBodyFormat, Status: "current", Version: 1,
	}
	if dryRun {
		return preview, nil
	}
	cred, err := c.credForHost(hostname)
	if err != nil {
		return domain.Page{}, err
	}
	spaceID, err := c.lookupSpaceID(ctx, hostname, in.Space)
	if err != nil {
		return domain.Page{}, err
	}
	payload := map[string]any{
		"spaceId": spaceID,
		"status":  "current",
		"title":   in.Title,
		"body": map[string]any{
			"representation": "storage",
			"value":          in.Body,
		},
	}
	u := joinURL(c.origin(hostname), "/wiki/api/v2/pages")
	code, body, err := c.doJSON(ctx, http.MethodPost, u, cred, nil, payload)
	if err != nil {
		return domain.Page{}, err
	}
	if code != http.StatusOK && code != http.StatusCreated {
		return domain.Page{}, MapConfluenceStatus(code)
	}
	p := pageFromREST(hostname, mustJSON(body))
	if p.Space == "" {
		p.Space = in.Space
	}
	p.Body = in.Body
	return p, nil
}

func (c Confluence) Update(ctx context.Context, hostname, pageID, body string, dryRun bool) (domain.Page, error) {
	cur, err := c.Get(ctx, hostname, pageID)
	if err != nil {
		return domain.Page{}, err
	}
	if dryRun {
		cur.Body = body
		cur.Version++
		return cur, nil
	}
	cred, err := c.credForHost(hostname)
	if err != nil {
		return domain.Page{}, err
	}
	payload := map[string]any{
		"id":     pageID,
		"status": "current",
		"title":  cur.Title,
		"body": map[string]any{
			"representation": "storage",
			"value":          body,
		},
		"version": map[string]any{"number": cur.Version + 1},
	}
	u := joinURL(c.origin(hostname), "/wiki/api/v2/pages/"+q(pageID))
	code, raw, err := c.doJSON(ctx, http.MethodPut, u, cred, nil, payload)
	if err != nil {
		return domain.Page{}, err
	}
	if code != http.StatusOK {
		return domain.Page{}, MapConfluenceStatus(code)
	}
	p := pageFromREST(hostname, mustJSON(raw))
	p.Body = body
	return p, nil
}

func (c Confluence) lookupSpaceID(ctx context.Context, hostname, space string) (string, error) {
	if _, err := strconv.Atoi(strings.TrimSpace(space)); err == nil {
		return strings.TrimSpace(space), nil
	}
	cred, err := c.credForHost(hostname)
	if err != nil {
		return "", err
	}
	u := joinURL(c.origin(hostname), "/wiki/api/v2/spaces?keys="+q(strings.ToUpper(space)))
	code, body, err := c.doJSON(ctx, http.MethodGet, u, cred, nil, nil)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK {
		return "", MapConfluenceStatus(code)
	}
	results := asList(mustJSON(body)["results"])
	if len(results) == 0 {
		return "", domain.Usagef("unknown space %q", space)
	}
	id := str(asMap(results[0]), "id")
	if id == "" {
		return "", domain.Usagef("unknown space %q", space)
	}
	return id, nil
}

func pageFromREST(hostname string, m map[string]any) domain.Page {
	id := str(m, "id")
	space := str(m, "spaceId")
	if space == "" {
		space = str(asMap(m["space"]), "key")
	}
	body := ""
	if b := asMap(m["body"]); b != nil {
		if st := asMap(b["storage"]); st != nil {
			body = str(st, "value")
		} else if st := asMap(b["value"]); false {
			_ = st
		} else {
			body = str(b, "value")
		}
	}
	ver := 0
	if v := asMap(m["version"]); v != nil {
		if n, ok := v["number"].(float64); ok {
			ver = int(n)
		}
	}
	p := domain.Page{
		ID: id, Site: hostname, Space: space, Title: str(m, "title"),
		Body: body, ContentFormat: domain.DefaultBodyFormat, Status: str(m, "status"),
		URL: domain.WikiPageURL(hostname, space, id), Version: ver,
	}
	if p.Status == "" {
		p.Status = "current"
	}
	return p
}

func pageFromV1(hostname string, m map[string]any) domain.Page {
	id := str(m, "id")
	space := str(asMap(m["space"]), "key")
	body := ""
	if b := asMap(m["body"]); b != nil {
		if st := asMap(b["storage"]); st != nil {
			body = str(st, "value")
		}
	}
	return domain.Page{
		ID: id, Site: hostname, Space: space, Title: str(m, "title"),
		Body: body, ContentFormat: domain.DefaultBodyFormat, Status: str(m, "status"),
		URL: domain.WikiPageURL(hostname, space, id),
	}
}

// MapConfluenceStatus maps Confluence HTTP statuses to exit classes.
func MapConfluenceStatus(code int) error {
	return classify(code, "confluence", "page not found")
}
