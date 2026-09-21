package atlassian

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/masonhuemmer/atlas/internal/domain"
)

// Memory is the fake Jira REST seed. Issues are keyed by hostname+key.
type Memory struct {
	mu     sync.Mutex
	issues map[string]domain.Issue
}

// JiraAPI is the in-process adapter used by ATLAS_FAKE and tests.
type JiraAPI struct {
	Memory *Memory
}

func (j JiraAPI) Get(ctx context.Context, hostname, key string) (domain.Issue, error) {
	if j.Memory == nil {
		return domain.Issue{}, domain.Service("jira memory not configured")
	}
	return j.Memory.Get(ctx, hostname, key)
}

func (j JiraAPI) Search(ctx context.Context, hostname, jql string) (domain.SearchResult, error) {
	if j.Memory == nil {
		return domain.SearchResult{}, domain.Service("jira memory not configured")
	}
	return j.Memory.Search(ctx, hostname, jql)
}

func (m *Memory) Get(_ context.Context, hostname, key string) (domain.Issue, error) {
	if m == nil {
		return domain.Issue{}, domain.Service("jira memory not configured")
	}
	hostname = strings.TrimSpace(hostname)
	key = strings.ToUpper(strings.TrimSpace(key))
	if hostname == "" || key == "" {
		return domain.Issue{}, domain.Usage("hostname and issue key are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	iss, ok := m.issues[memKey(hostname, key)]
	if !ok {
		return domain.Issue{}, domain.NotFound("issue not found").WithHint("check the key and site")
	}
	return iss, nil
}

func (m *Memory) Search(_ context.Context, hostname, jql string) (domain.SearchResult, error) {
	if m == nil {
		return domain.SearchResult{}, domain.Service("jira memory not configured")
	}
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return domain.SearchResult{}, domain.Usage("hostname is required")
	}
	projects := domain.JQLProjects(jql)
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]domain.Issue, 0)
	for _, iss := range m.issues {
		if iss.Site != hostname {
			continue
		}
		if len(projects) > 0 && !containsFold(projects, iss.Project) {
			continue
		}
		items = append(items, iss)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
	return domain.SearchResult{JQL: jql, Site: hostname, Count: len(items), Items: items}, nil
}

func containsFold(keys []string, v string) bool {
	for _, k := range keys {
		if strings.EqualFold(k, v) {
			return true
		}
	}
	return false
}
