package domain

import (
	"strings"
	"sync"
)

const (
	RoleLicensed    = "licensed"
	RoleJSMCustomer = "jsm_customer"
)

// Catalog is tenant data loaded from config. Isolation rules are not in this file.
type Catalog struct {
	Sites    []Site
	Projects map[string]string // project key → site alias
	Spaces   map[string]string // space key → site alias
	Defaults Defaults
}

// Defaults are optional create/PR/JSM values from config or env.
type Defaults struct {
	Workspace         string
	DefaultSite       string
	AssigneeAccountID string
	JSMSite           string
}

var (
	catalogMu sync.RWMutex
	catalog   Catalog
)

// SetCatalog replaces the process catalog. CLI and adapters read it via Lookup / Resolve.
func SetCatalog(c Catalog) {
	catalogMu.Lock()
	defer catalogMu.Unlock()
	catalog = cloneCatalog(c)
}

// CurrentCatalog returns a copy of the loaded catalog.
func CurrentCatalog() Catalog {
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	return cloneCatalog(catalog)
}

// Sites returns the loaded site table. Empty when no config has been applied.
func Sites() []Site {
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	out := make([]Site, len(catalog.Sites))
	copy(out, catalog.Sites)
	return out
}

// DefaultWorkspace is the Bitbucket Cloud workspace from catalog defaults.
func DefaultWorkspace() string {
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	return catalog.Defaults.Workspace
}

// DefaultAssigneeAccountID is the create assignee from catalog defaults. Empty means do not default.
func DefaultAssigneeAccountID() string {
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	return catalog.Defaults.AssigneeAccountID
}

// DefaultJSMSite is the jsm_customer alias from catalog defaults. Empty means jsm requires --site.
func DefaultJSMSite() string {
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	return catalog.Defaults.JSMSite
}

// DefaultSiteAlias is the optional fallback alias when inference cannot run.
func DefaultSiteAlias() string {
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	return catalog.Defaults.DefaultSite
}

func aliasForProject(p string) (string, bool) {
	p = strings.ToUpper(strings.TrimSpace(p))
	if p == "" {
		return "", false
	}
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	if catalog.Projects == nil {
		return "", false
	}
	a, ok := catalog.Projects[p]
	return a, ok
}

func aliasForSpace(sp string) (string, bool) {
	sp = strings.ToUpper(strings.TrimSpace(sp))
	if sp == "" {
		return "", false
	}
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	if catalog.Spaces == nil {
		return "", false
	}
	a, ok := catalog.Spaces[sp]
	return a, ok
}

func cloneCatalog(c Catalog) Catalog {
	out := Catalog{
		Sites:    append([]Site(nil), c.Sites...),
		Defaults: c.Defaults,
	}
	if c.Projects != nil {
		out.Projects = make(map[string]string, len(c.Projects))
		for k, v := range c.Projects {
			out.Projects[k] = v
		}
	}
	if c.Spaces != nil {
		out.Spaces = make(map[string]string, len(c.Spaces))
		for k, v := range c.Spaces {
			out.Spaces[k] = v
		}
	}
	return out
}
