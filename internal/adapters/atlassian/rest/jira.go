package rest

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/masonhuemmer/atlas/internal/domain"
)

const jiraFields = "summary,description,status,issuetype,priority,labels,assignee,reporter,created,updated,project,comment,issuelinks"

var jiraFieldList = []string{
	"summary", "description", "status", "issuetype", "priority", "labels",
	"assignee", "reporter", "created", "updated", "project", "comment", "issuelinks",
}

// Jira is the live Cloud REST client (https://<hostname>/rest/api/3).
type Jira struct {
	*Client
}

func (j Jira) Get(ctx context.Context, hostname, key string) (domain.Issue, error) {
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.Issue{}, err
	}
	u := joinURL(j.origin(hostname), "/rest/api/3/issue/"+q(strings.ToUpper(strings.TrimSpace(key)))+"?fields="+jiraFields)
	code, body, err := j.doJSON(ctx, http.MethodGet, u, cred, nil, nil)
	if err != nil {
		return domain.Issue{}, err
	}
	if code != http.StatusOK {
		return domain.Issue{}, MapStatus(code)
	}
	return issueFromREST(hostname, mustJSON(body)), nil
}

func (j Jira) Search(ctx context.Context, hostname, jql string) (domain.SearchResult, error) {
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.SearchResult{}, err
	}
	vals := url.Values{}
	vals.Set("jql", jql)
	vals.Set("maxResults", "50")
	vals.Set("fields", strings.Join(jiraFieldList, ","))
	getURL := joinURL(j.origin(hostname), "/rest/api/3/search/jql") + "?" + vals.Encode()
	code, body, err := j.doJSON(ctx, http.MethodGet, getURL, cred, nil, nil)
	if err != nil {
		return domain.SearchResult{}, err
	}
	parsed := mustJSON(body)
	if err := jiraSearchAPIError(code, parsed); err != nil {
		return domain.SearchResult{}, err
	}
	items := issuesFromSearch(hostname, parsed)
	if code != http.StatusOK || len(items) == 0 {
		payload := map[string]any{
			"jql":        jql,
			"maxResults": 50,
			"fields":     jiraFieldList,
		}
		postURL := joinURL(j.origin(hostname), "/rest/api/3/search/jql")
		code, body, err = j.doJSON(ctx, http.MethodPost, postURL, cred, nil, payload)
		if err != nil {
			return domain.SearchResult{}, err
		}
		parsed = mustJSON(body)
		if err := jiraSearchAPIError(code, parsed); err != nil {
			return domain.SearchResult{}, err
		}
		items = issuesFromSearch(hostname, parsed)
	}
	items = j.hydrateSearch(ctx, hostname, items)
	return domain.SearchResult{JQL: jql, Site: hostname, Count: len(items), Items: items}, nil
}

func jiraSearchAPIError(code int, m map[string]any) error {
	msg := ""
	if msgs := asList(m["errorMessages"]); len(msgs) > 0 {
		msg, _ = msgs[0].(string)
	}
	if code != http.StatusOK && code != 0 {
		err := MapStatus(code)
		if msg != "" {
			if de, ok := err.(*domain.Error); ok {
				return de.WithHint(msg)
			}
		}
		return err
	}
	if msg != "" {
		return domain.Usage(msg)
	}
	return nil
}

func issuesFromSearch(hostname string, m map[string]any) []domain.Issue {
	raw := asList(m["issues"])
	items := make([]domain.Issue, 0, len(raw))
	for _, e := range raw {
		switch t := e.(type) {
		case map[string]any:
			iss := issueFromREST(hostname, t)
			if iss.Key == "" {
				iss.Key = str(t, "key")
			}
			if iss.Key == "" {
				iss.Key = issueIDString(t["id"])
			}
			if iss.Key != "" {
				items = append(items, iss)
			}
		case string:
			if t != "" {
				items = append(items, domain.Issue{Key: t, Site: hostname, BrowseURL: domain.BrowseURL(hostname, t)})
			}
		case float64:
			id := strconv.FormatInt(int64(t), 10)
			items = append(items, domain.Issue{Key: id, Site: hostname})
		}
	}
	return items
}

func issueIDString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	default:
		return ""
	}
}

func (j Jira) hydrateSearch(ctx context.Context, hostname string, items []domain.Issue) []domain.Issue {
	out := make([]domain.Issue, 0, len(items))
	for _, iss := range items {
		if iss.Summary != "" && strings.Contains(iss.Key, "-") {
			out = append(out, iss)
			continue
		}
		got, err := j.Get(ctx, hostname, iss.Key)
		if err != nil {
			if iss.Key != "" {
				out = append(out, iss)
			}
			continue
		}
		out = append(out, got)
	}
	return out
}

func (j Jira) Create(ctx context.Context, hostname string, in domain.CreateIssue, dryRun bool) (domain.Issue, error) {
	preview := domain.Issue{
		Site: hostname, Project: in.Project, IssueType: in.IssueType,
		Summary: in.Summary, Description: in.Description, Labels: in.Labels,
		Assignee: in.Assignee, Status: "To Do",
	}
	if dryRun {
		return preview, nil
	}
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.Issue{}, err
	}
	fields := map[string]any{
		"project":   map[string]any{"key": in.Project},
		"issuetype": map[string]any{"name": in.IssueType},
		"summary":   in.Summary,
	}
	if in.Description != "" {
		fields["description"] = adfFromText(in.Description)
	}
	if len(in.Labels) > 0 {
		fields["labels"] = in.Labels
	}
	if in.Assignee != "" {
		fields["assignee"] = map[string]any{"accountId": in.Assignee}
	}
	u := joinURL(j.origin(hostname), "/rest/api/3/issue")
	code, body, err := j.doJSON(ctx, http.MethodPost, u, cred, nil, map[string]any{"fields": fields})
	if err != nil {
		return domain.Issue{}, err
	}
	if code != http.StatusCreated && code != http.StatusOK {
		return domain.Issue{}, MapStatus(code)
	}
	key := str(mustJSON(body), "key")
	if key == "" {
		return domain.Issue{}, domain.Service("jira create returned no key")
	}
	return j.Get(ctx, hostname, key)
}

func (j Jira) Edit(ctx context.Context, hostname, key string, fields map[string]any, dryRun bool) (domain.Issue, error) {
	if dryRun {
		iss, err := j.Get(ctx, hostname, key)
		if err != nil {
			return domain.Issue{}, err
		}
		return iss, nil
	}
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.Issue{}, err
	}
	u := joinURL(j.origin(hostname), "/rest/api/3/issue/"+q(strings.ToUpper(strings.TrimSpace(key))))
	code, _, err := j.doJSON(ctx, http.MethodPut, u, cred, nil, map[string]any{"fields": fields})
	if err != nil {
		return domain.Issue{}, err
	}
	if code != http.StatusNoContent && code != http.StatusOK {
		return domain.Issue{}, MapStatus(code)
	}
	return j.Get(ctx, hostname, key)
}

func (j Jira) Comment(ctx context.Context, hostname, key, body string, dryRun bool) error {
	if dryRun {
		return nil
	}
	cred, err := j.credForHost(hostname)
	if err != nil {
		return err
	}
	u := joinURL(j.origin(hostname), "/rest/api/3/issue/"+q(strings.ToUpper(strings.TrimSpace(key)))+"/comment")
	code, _, err := j.doJSON(ctx, http.MethodPost, u, cred, nil, map[string]any{"body": adfFromText(body)})
	if err != nil {
		return err
	}
	if code != http.StatusCreated && code != http.StatusOK {
		return MapStatus(code)
	}
	return nil
}

func (j Jira) Transition(ctx context.Context, hostname, key, name string, dryRun bool) (domain.Issue, error) {
	cred, err := j.credForHost(hostname)
	if err != nil {
		return domain.Issue{}, err
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	u := joinURL(j.origin(hostname), "/rest/api/3/issue/"+q(key)+"/transitions")
	code, body, err := j.doJSON(ctx, http.MethodGet, u, cred, nil, nil)
	if err != nil {
		return domain.Issue{}, err
	}
	if code != http.StatusOK {
		return domain.Issue{}, MapStatus(code)
	}
	id := ""
	for _, raw := range asList(mustJSON(body)["transitions"]) {
		t := asMap(raw)
		if strings.EqualFold(str(t, "name"), name) {
			id = str(t, "id")
			break
		}
	}
	if id == "" {
		return domain.Issue{}, domain.Usagef("unknown transition %q", name).WithHint("atlas jira transition KEY --name NAME")
	}
	if dryRun {
		return j.Get(ctx, hostname, key)
	}
	code, _, err = j.doJSON(ctx, http.MethodPost, u, cred, nil, map[string]any{"transition": map[string]any{"id": id}})
	if err != nil {
		return domain.Issue{}, err
	}
	if code != http.StatusNoContent && code != http.StatusOK {
		return domain.Issue{}, MapStatus(code)
	}
	return j.Get(ctx, hostname, key)
}

func (j Jira) Link(ctx context.Context, hostname, inward, outward, linkType string, dryRun bool) error {
	if dryRun {
		return nil
	}
	cred, err := j.credForHost(hostname)
	if err != nil {
		return err
	}
	if linkType == "" {
		linkType = domain.DefaultLinkType
	}
	u := joinURL(j.origin(hostname), "/rest/api/3/issueLink")
	payload := map[string]any{
		"type":         map[string]any{"name": linkType},
		"inwardIssue":  map[string]any{"key": strings.ToUpper(inward)},
		"outwardIssue": map[string]any{"key": strings.ToUpper(outward)},
	}
	code, _, err := j.doJSON(ctx, http.MethodPost, u, cred, nil, payload)
	if err != nil {
		return err
	}
	if code != http.StatusCreated && code != http.StatusOK && code != http.StatusNoContent {
		return MapStatus(code)
	}
	return nil
}

func issueFromREST(hostname string, m map[string]any) domain.Issue {
	key := str(m, "key")
	f := asMap(m["fields"])
	iss := domain.Issue{
		Key:         key,
		Site:        hostname,
		BrowseURL:   domain.BrowseURL(hostname, key),
		Summary:     str(f, "summary"),
		Description: textFromADF(f["description"]),
		Status:      nestedName(f, "status"),
		IssueType:   nestedName(f, "issuetype"),
		Priority:    nestedName(f, "priority"),
		Assignee:    displayUser(asMap(f["assignee"])),
		Reporter:    displayUser(asMap(f["reporter"])),
		Created:     str(f, "created"),
		Updated:     str(f, "updated"),
		Project:     str(asMap(f["project"]), "key"),
	}
	for _, raw := range asList(f["labels"]) {
		if s, ok := raw.(string); ok {
			iss.Labels = append(iss.Labels, s)
		}
	}
	if c := asMap(f["comment"]); c != nil {
		for _, raw := range asList(c["comments"]) {
			cm := asMap(raw)
			if t := textFromADF(cm["body"]); t != "" {
				iss.Comments = append(iss.Comments, t)
			}
		}
	}
	for _, raw := range asList(f["issuelinks"]) {
		l := asMap(raw)
		typ := nestedName(l, "type")
		if in := asMap(l["inwardIssue"]); in != nil {
			iss.Links = append(iss.Links, domain.IssueLink{Type: typ, Inward: str(in, "key"), Outward: key})
		}
		if out := asMap(l["outwardIssue"]); out != nil {
			iss.Links = append(iss.Links, domain.IssueLink{Type: typ, Inward: key, Outward: str(out, "key")})
		}
	}
	return iss
}

func displayUser(m map[string]any) string {
	if m == nil {
		return ""
	}
	if s := str(m, "displayName"); s != "" {
		return s
	}
	return str(m, "accountId")
}

// MapStatus maps Jira HTTP statuses to exit classes.
// 401/403 → auth, 404 → not_found, 429/5xx → service.
func MapStatus(code int) error {
	return classify(code, "jira", "issue not found")
}
