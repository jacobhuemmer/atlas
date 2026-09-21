package domain

// DefaultAssigneeAccountID is Mason Huemmer on Sesami licensed sites.
const DefaultAssigneeAccountID = "712020:092d1246-2a54-4f40-8317-ef3dc777bf3c"

// Issue is one licensed Jira issue on a single cloud.
// Site is the hostname. BrowseURL is https://<hostname>/browse/<KEY>.
type Issue struct {
	Key         string   `json:"key"`
	Site        string   `json:"site"`
	BrowseURL   string   `json:"browse_url"`
	Summary     string   `json:"summary"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	IssueType   string   `json:"issuetype"`
	Priority    string   `json:"priority,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Assignee    string   `json:"assignee,omitempty"`
	Reporter    string   `json:"reporter,omitempty"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated,omitempty"`
	Project     string   `json:"project"`
	Comments    []string `json:"comments,omitempty"`
}

// SearchResult is one-site JQL output. Items use the default search fields.
type SearchResult struct {
	JQL   string  `json:"jql"`
	Site  string  `json:"site"`
	Count int     `json:"count"`
	Items []Issue `json:"items"`
}

// CreateIssue is the REST field set we own for atlas jira create.
// Markdown is the default description format. Assignee defaults to DefaultAssigneeAccountID.
type CreateIssue struct {
	Project     string
	IssueType   string
	Summary     string
	Description string
	Labels      []string
	Assignee    string
}

func BrowseURL(hostname, key string) string {
	return "https://" + hostname + "/browse/" + key
}
