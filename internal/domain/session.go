package domain

type SiteStatus struct {
	Alias    string `json:"alias"`
	Hostname string `json:"hostname"`
	Role     string `json:"role"`
	Usable   bool   `json:"usable"`
}

type Session struct {
	SignedIn      bool         `json:"signed_in"`
	SessionUsable bool         `json:"session_usable"`
	Sites         []SiteStatus `json:"sites"`
}

func SignedOut() Session {
	sites := make([]SiteStatus, len(Sites))
	for i, s := range Sites {
		sites[i] = SiteStatus{Alias: s.Alias, Hostname: s.Hostname, Role: s.Role, Usable: false}
	}
	return Session{Sites: sites}
}
