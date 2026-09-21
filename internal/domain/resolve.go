package domain

import (
	"regexp"
	"strings"
)

// ResolveInput is enough context to pick exactly one of the three clouds.
// Site may be an alias, hostname, or UUID. --site is required only when
// nothing else can infer (raw JQL with no project, numeric id only).
type ResolveInput struct {
	Site    string // alias, hostname, or UUID
	Project string
	Issue   string // SDO-1, CAB-12
	Space   string // CCAB
	JQL     string
}

var (
	reIssueKey   = regexp.MustCompile(`(?i)^([A-Za-z][A-Za-z0-9]*)-(\d+)$`)
	reProjectEq  = regexp.MustCompile(`(?i)\bproject\s*=\s*['"]?([A-Za-z][A-Za-z0-9]+)['"]?`)
	reProjectIn  = regexp.MustCompile(`(?i)\bproject\s+in\s*\(([^)]*)\)`)
	reProjectTok = regexp.MustCompile(`[A-Za-z][A-Za-z0-9]+`)
)

// Resolve returns exactly one Site. Inference never crosses clouds.
func Resolve(in ResolveInput) (Site, error) {
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

	switch len(aliases) {
	case 0:
		return Site{}, Usage("cannot infer site").WithHint("pass --site sesamidevel, sesami-io, or garda")
	case 1:
		return Lookup(aliases[0])
	default:
		return Site{}, Usage("JQL names projects from two sites").WithHint("do not dual-query sesamidevel and sesami-io; split the query or pass one --site")
	}
}

// ProjectFromIssue returns the project key from SDO-1. Numeric ids do not infer.
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

func aliasForProject(p string) (string, bool) {
	switch strings.ToUpper(strings.TrimSpace(p)) {
	case "SDO", "SDP", "SES":
		return "sesamidevel", true
	case "CAB":
		return "sesami-io", true
	default:
		return "", false
	}
}

func aliasForSpace(sp string) (string, bool) {
	if strings.EqualFold(strings.TrimSpace(sp), "CCAB") {
		return "sesami-io", true
	}
	return "", false
}

func addAlias(aliases []string, alias string) []string {
	for _, a := range aliases {
		if a == alias {
			return aliases
		}
	}
	return append(aliases, alias)
}
