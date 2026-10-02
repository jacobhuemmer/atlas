package domain

// JiraUser is a Jira Cloud user returned by a user search.
type JiraUser struct {
	Site        string `json:"site"`
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email,omitempty"`
	Active      bool   `json:"active"`
}

// UserSearchResult contains users from one Jira site. Project or Issue narrows
// the search to accounts assignable to that target.
type UserSearchResult struct {
	Site    string     `json:"site"`
	Query   string     `json:"query"`
	Project string     `json:"project,omitempty"`
	Issue   string     `json:"issue,omitempty"`
	Count   int        `json:"count"`
	Users   []JiraUser `json:"users"`
}
