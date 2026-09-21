package domain

import "strings"

type Site struct {
	Alias    string `json:"alias"`
	Hostname string `json:"hostname"`
	UUID     string `json:"uuid"`
	Role     string `json:"role"`
}

func Lookup(id string) (Site, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Site{}, Usage("site is required").WithHint("pass an alias, hostname, or UUID from the config")
	}
	sites := Sites()
	if len(sites) == 0 {
		return Site{}, Usage("no sites configured").WithHint("write $XDG_CONFIG_HOME/atlas/config.toml or set ATLAS_CONFIG")
	}
	for _, s := range sites {
		if strings.EqualFold(s.Alias, id) || strings.EqualFold(s.Hostname, id) || strings.EqualFold(s.UUID, id) {
			return s, nil
		}
	}
	return Site{}, Usagef("unknown site %q", id).WithHint("use an alias, hostname, or UUID from atlas site list")
}

// LookupJSM returns the jsm_customer site for hostname or alias.
func LookupJSM(id string) (Site, error) {
	site, err := Lookup(id)
	if err != nil {
		return Site{}, err
	}
	if site.Role != RoleJSMCustomer {
		return Site{}, Usage("jsm is customer REST only").WithHint("use a site with role jsm_customer; licensed Jira stays atlas jira")
	}
	return site, nil
}
