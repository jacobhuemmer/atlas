package atlassian

import (
	"strconv"

	"github.com/masonhuemmer/atlas/internal/domain"
)

const (
	hostDevel = "sesamidevel.atlassian.net"
	hostIO    = "sesami-io.atlassian.net"
	hostGarda = domain.GardaHostname
)

const seedCCABPageID = "100"

const seedPRDiff = `diff --git a/README.md b/README.md
--- a/README.md
+++ b/README.md
@@ -1 +1,2 @@
 atlas
+phase 6
`

// Seed returns in-memory issues, pages, PRs, and Garda JSM desks for ATLAS_FAKE and tests.
func Seed() *Memory {
	m := &Memory{
		issues:         map[string]domain.Issue{},
		transitions:    map[string][]string{},
		next:           map[string]int{},
		pages:          map[string]domain.Page{},
		spaces:         map[string]string{"CCAB": "ccab-space-id"},
		nextPage:       101,
		prs:            map[string]domain.PullRequest{},
		prDiffs:        map[string]string{},
		nextPR:         map[string]int{},
		types:          map[string][]domain.RequestType{},
		requests:       map[string]domain.CustomerRequest{},
		jsmTransitions: map[string][]string{},
		nextReq:        map[string]int{},
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
	m.putPage(domain.Page{
		ID:            seedCCABPageID,
		Site:          hostIO,
		Space:         "CCAB",
		Title:         "CAB-109",
		Body:          "CCAB page for CAB-109",
		ContentFormat: domain.DefaultBodyFormat,
		Status:        "current",
		URL:           domain.WikiPageURL(hostIO, "CCAB", seedCCABPageID),
		Version:       1,
	})
	pr := domain.PullRequest{
		ID:          1,
		Workspace:   domain.DefaultWorkspace,
		Repo:        "atlas",
		Title:       "Phase 6 Bitbucket PRs",
		Description: "Seed PR for atlas pr get",
		Source:      "feature/pr-cli",
		Target:      domain.DefaultTargetBranch,
		State:       "OPEN",
		URL:         domain.PRURL(domain.DefaultWorkspace, "atlas", 1),
	}
	m.putPR(pr)
	m.prDiffs[prKey(pr.Workspace, pr.Repo, pr.ID)] = seedPRDiff
	m.nextPR[prRepoKey(pr.Workspace, pr.Repo)] = 2
	m.desks = []domain.ServiceDesk{
		{ID: "3", Key: "EOS", Name: "Engineering Operational Service"},
		{ID: "12", Key: "ITSEC", Name: "IT Security"},
		{ID: "2586", Key: "COSC", Name: "CloudOps Service Center"},
	}
	m.types["3"] = []domain.RequestType{{ID: "40", Name: "Incident", DeskID: "3"}}
	m.types["12"] = []domain.RequestType{{ID: "50", Name: "Access request", DeskID: "12"}}
	m.types["2586"] = []domain.RequestType{{ID: "60", Name: "Change", DeskID: "2586"}}
	req := domain.CustomerRequest{
		Key:            "EOS-1",
		Site:           hostGarda,
		DeskID:         "3",
		TypeID:         "40",
		Summary:        "Seed Garda request",
		Description:    "Customer REST only; never Jira search",
		Status:         "Waiting for support",
		StatusCategory: domain.JSMCategoryOpen,
		PortalURL:      domain.PortalURL(hostGarda, "3", "EOS-1"),
	}
	m.putRequest(req)
	m.jsmTransitions[memKey(hostGarda, "EOS-1")] = []string{"21"}
	m.nextReq["EOS"] = 2
	m.nextReq["ITSEC"] = 1
	m.nextReq["COSC"] = 1
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

func prRepoKey(workspace, repo string) string {
	return workspace + "\x00" + repo
}

func prKey(workspace, repo string, id int) string {
	return workspace + "\x00" + repo + "\x00" + strconv.Itoa(id)
}
