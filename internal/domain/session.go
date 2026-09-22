package domain

type SiteStatus struct {
	Alias    string `json:"alias"`
	Hostname string `json:"hostname"`
	Role     string `json:"role"`
	Usable   bool   `json:"usable"`
}

type WorkspaceStatus struct {
	Slug   string `json:"slug"`
	Usable bool   `json:"usable"`
}

type Session struct {
	SignedIn      bool              `json:"signed_in"`
	SessionUsable bool              `json:"session_usable"`
	Sites         []SiteStatus      `json:"sites"`
	Workspaces    []WorkspaceStatus `json:"workspaces"`
}

func SignedOut() Session {
	table := Sites()
	sites := make([]SiteStatus, len(table))
	for i, s := range table {
		sites[i] = SiteStatus{Alias: s.Alias, Hostname: s.Hostname, Role: s.Role, Usable: false}
	}
	session := Session{Sites: sites, Workspaces: []WorkspaceStatus{}}
	if workspace := DefaultWorkspace(); workspace != "" {
		session.Workspaces = append(session.Workspaces, WorkspaceStatus{Slug: workspace})
	}
	return session
}
