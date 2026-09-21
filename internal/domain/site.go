package domain

import "strings"

type Site struct {
	Alias    string `json:"alias"`
	Hostname string `json:"hostname"`
	UUID     string `json:"uuid"`
	Role     string `json:"role"`
}

var Sites = []Site{
	{"sesamidevel", "sesamidevel.atlassian.net", "dd528466-a332-4f63-9cee-0033fe01f875", "licensed"},
	{"sesami-io", "sesami-io.atlassian.net", "8986e79a-6c2a-4e0d-bb65-4a57e0f68db9", "licensed"},
	{"garda", "gardaworld.atlassian.net", "62907a08-fd61-4a15-b642-f79d03e70b92", "jsm_customer"},
}

func Lookup(id string) (Site, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Site{}, Usage("site is required").WithHint("pass an alias, hostname, or UUID")
	}
	for _, s := range Sites {
		if strings.EqualFold(s.Alias, id) || strings.EqualFold(s.Hostname, id) || strings.EqualFold(s.UUID, id) {
			return s, nil
		}
	}
	return Site{}, Usagef("unknown site %q", id).WithHint("use sesamidevel, sesami-io, or garda")
}
