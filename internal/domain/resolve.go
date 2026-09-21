package domain

import (
	"regexp"
	"strings"
)

// ResolveInput is enough context to pick exactly one catalog site.
// Site may be an alias, hostname, or UUID. --site is required only when
// nothing else can infer (raw JQL with no project, CQL with no space, numeric id only).
type ResolveInput struct {
	Site    string // alias, hostname, or UUID
	Project string
	Issue   string
	Space   string
	JQL     string
	CQL     string
}

var (
	reIssueKey   = regexp.MustCompile(`(?i)^([A-Za-z][A-Za-z0-9]*)-(\d+)$`)
	reProjectEq  = regexp.MustCompile(`(?i)\bproject\s*=\s*['"]?([A-Za-z][A-Za-z0-9]+)['"]?`)
	reProjectIn  = regexp.MustCompile(`(?i)\bproject\s+in\s*\(([^)]*)\)`)
	reProjectTok = regexp.MustCompile(`[A-Za-z][A-Za-z0-9]+`)
	reSpaceEq    = regexp.MustCompile(`(?i)\bspace\s*=\s*['"]?([A-Za-z][A-Za-z0-9]+)['"]?`)
	reSpaceIn    = regexp.MustCompile(`(?i)\bspace\s+in\s*\(([^)]*)\)`)
)

// Resolve returns exactly one Site. Inference never crosses clouds.
func Resolve(in ResolveInput) (Site, error) {
	if len(Sites()) == 0 {
		return Site{}, Usage("no sites configured").WithHint("write $XDG_CONFIG_HOME/atlas/config.toml or set ATLAS_CONFIG")
	}
	var aliases []string

	if s := strings.TrimSpace(in.Site); s != "" {
		site, err := Lookup(s)
		if err != nil {
			return Site{}, err
		}
		aliases = addAlias(aliases, site.Alias)
	}

	if p := strings.TrimSpace(in.Project); p != "" {
		if a, ok := aliasForProject(p); ok {
			aliases = addAlias(aliases, a)
		}
	}

	if k := strings.TrimSpace(in.Issue); k != "" {
		if p, ok := ProjectFromIssue(k); ok {
			if a, ok := aliasForProject(p); ok {
				aliases = addAlias(aliases, a)
			}
		}
	}

	if sp := strings.TrimSpace(in.Space); sp != "" {
		if a, ok := aliasForSpace(sp); ok {
			aliases = addAlias(aliases, a)
		}
	}

	if jql := strings.TrimSpace(in.JQL); jql != "" {
		for _, p := range JQLProjects(jql) {
			if a, ok := aliasForProject(p); ok {
				aliases = addAlias(aliases, a)
			}
		}
	}

	if cql := strings.TrimSpace(in.CQL); cql != "" {
		for _, sp := range CQLSpaces(cql) {
			if a, ok := aliasForSpace(sp); ok {
				aliases = addAlias(aliases, a)
			}
		}
	}

	switch len(aliases) {
	case 0:
		if def := DefaultSiteAlias(); def != "" {
			return Lookup(def)
		}
		return Site{}, Usage("cannot infer site").WithHint("pass --site ALIAS from atlas site list")
	case 1:
		return Lookup(aliases[0])
	default:
		return Site{}, Usage("query names projects or spaces from two sites").WithHint("do not dual-query clouds; split the query or pass one --site")
	}
}

// ProjectFromIssue returns the project key from KEY-1. Numeric ids do not infer.
func ProjectFromIssue(key string) (string, bool) {
	m := reIssueKey.FindStringSubmatch(strings.TrimSpace(key))
	if m == nil {
		return "", false
	}
	return strings.ToUpper(m[1]), true
}

// JQLProjects returns project keys named by project = KEY or project in (...).
func JQLProjects(jql string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	for _, m := range reProjectEq.FindAllStringSubmatch(jql, -1) {
		add(m[1])
	}
	for _, m := range reProjectIn.FindAllStringSubmatch(jql, -1) {
		for _, tok := range reProjectTok.FindAllString(m[1], -1) {
			add(tok)
		}
	}
	return out
}

// CQLSpaces returns space keys named by space = KEY or space in (...).
func CQLSpaces(cql string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	for _, m := range reSpaceEq.FindAllStringSubmatch(cql, -1) {
		add(m[1])
	}
	for _, m := range reSpaceIn.FindAllStringSubmatch(cql, -1) {
		for _, tok := range reProjectTok.FindAllString(m[1], -1) {
			add(tok)
		}
	}
	return out
}

func addAlias(aliases []string, alias string) []string {
	for _, a := range aliases {
		if a == alias {
			return aliases
		}
	}
	return append(aliases, alias)
}
