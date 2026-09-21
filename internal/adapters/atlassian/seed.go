package atlassian

import "github.com/masonhuemmer/atlas/internal/domain"

const (
	hostDevel = "sesamidevel.atlassian.net"
	hostIO    = "sesami-io.atlassian.net"
)

// Seed returns in-memory issues keyed by site+key for ATLAS_FAKE and tests.
func Seed() *Memory {
	m := &Memory{
		issues:      map[string]domain.Issue{},
		transitions: map[string][]string{},
		next:        map[string]int{},
	}
	for _, iss := range []domain.Issue{
		{
			Key: "SDO-1", Site: hostDevel, BrowseURL: domain.BrowseURL(hostDevel, "SDO-1"),
			Summary: "Atlas CLI skeleton", Description: "Phase 1 binary and MCP door",
			Status: "To Do", IssueType: "Task", Priority: "Medium",
			Labels: []string{"atlas"}, Assignee: "Mason Huemmer", Reporter: "Mason Huemmer",
			Created: "2026-09-01T00:00:00Z", Updated: "2026-09-20T00:00:00Z", Project: "SDO",
		},
		{
			Key: "SDP-1", Site: hostDevel, BrowseURL: domain.BrowseURL(hostDevel, "SDP-1"),
			Summary: "Portal request", Description: "Licensed SDP on sesamidevel",
			Status: "To Do", IssueType: "Task", Priority: "Low",
			Assignee: "Mason Huemmer", Reporter: "Mason Huemmer",
			Created: "2026-09-02T00:00:00Z", Updated: "2026-09-20T00:00:00Z", Project: "SDP",
		},
		{
			Key: "SDP-2", Site: hostDevel, BrowseURL: domain.BrowseURL(hostDevel, "SDP-2"),
			Summary: "Portal follow-up", Description: "Second licensed SDP on sesamidevel",
			Status: "To Do", IssueType: "Task", Priority: "Low",
			Assignee: "Mason Huemmer", Reporter: "Mason Huemmer",
			Created: "2026-09-02T12:00:00Z", Updated: "2026-09-20T00:00:00Z", Project: "SDP",
		},
		{
			Key: "SES-1", Site: hostDevel, BrowseURL: domain.BrowseURL(hostDevel, "SES-1"),
			Summary: "Platform item", Description: "SES on sesamidevel",
			Status: "In Progress", IssueType: "Story", Priority: "High",
			Assignee: "Mason Huemmer", Reporter: "Mason Huemmer",
			Created: "2026-09-03T00:00:00Z", Updated: "2026-09-20T00:00:00Z", Project: "SES",
		},
		{
			Key: "CAB-1", Site: hostIO, BrowseURL: domain.BrowseURL(hostIO, "CAB-1"),
			Summary: "Change advisory", Description: "CAB on sesami-io",
			Status: "Open", IssueType: "Change", Priority: "High",
			Labels: []string{"cab"}, Assignee: "Mason Huemmer", Reporter: "Mason Huemmer",
			Created: "2026-09-04T00:00:00Z", Updated: "2026-09-21T00:00:00Z", Project: "CAB",
		},
	} {
		m.put(iss)
		m.transitions[memKey(iss.Site, iss.Key)] = defaultTransitions(iss.Project, iss.IssueType)
		if n := issueNumber(iss.Key); n+1 > m.next[iss.Project] {
			m.next[iss.Project] = n + 1
		}
	}
	return m
}

func (m *Memory) put(iss domain.Issue) {
	m.putLocked(iss)
}

func issueNumber(key string) int {
	i := len(key) - 1
	n := 0
	mult := 1
	for i >= 0 && key[i] >= '0' && key[i] <= '9' {
		n += int(key[i]-'0') * mult
		mult *= 10
		i--
	}
	return n
}

func memKey(hostname, key string) string {
	return hostname + "\x00" + key
}
