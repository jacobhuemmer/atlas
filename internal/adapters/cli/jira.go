package cli

import (
	"flag"
	"strings"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func runJira(args []string, d Deps, format string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return writeHelp(d.Stdout, jiraHelp)
	}
	verb, args := args[0], args[1:]
	switch verb {
	case "get":
		return jiraGet(args, d, format)
	case "search":
		return jiraSearch(args, d, format)
	default:
		return fail(d, domain.Usagef("unknown jira verb %q", verb))
	}
}

func jiraGet(args []string, d Deps, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, jiraHelp)
	}
	fsset := flag.NewFlagSet("jira get", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	siteFlag := fsset.String("site", "", "site alias, hostname, or UUID")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	key := fsset.Arg(0)
	if strings.TrimSpace(key) == "" {
		return fail(d, domain.Usage("issue key is required").WithHint("atlas jira get SDO-1"))
	}
	site, err := domain.Resolve(domain.ResolveInput{Site: *siteFlag, Issue: key})
	if err != nil {
		return fail(d, err)
	}
	if err := refuseGardaJira(site); err != nil {
		return fail(d, err)
	}
	if d.Jira == nil {
		return fail(d, domain.Service("jira adapter not configured"))
	}
	iss, err := d.Jira.Get(ctx(), site.Hostname, key)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, iss)
}

func jiraSearch(args []string, d Deps, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, jiraHelp)
	}
	fsset := flag.NewFlagSet("jira search", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	siteFlag := fsset.String("site", "", "site alias, hostname, or UUID")
	jql := fsset.String("jql", "", "JQL (one site only)")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	site, err := domain.Resolve(domain.ResolveInput{Site: *siteFlag, JQL: *jql})
	if err != nil {
		return fail(d, err)
	}
	if err := refuseGardaJira(site); err != nil {
		return fail(d, err)
	}
	if strings.TrimSpace(*jql) == "" {
		return fail(d, domain.Usage("search requires --jql").WithHint("atlas jira search --jql 'project = CAB'"))
	}
	if d.Jira == nil {
		return fail(d, domain.Service("jira adapter not configured"))
	}
	page, err := d.Jira.Search(ctx(), site.Hostname, *jql)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, page)
}

func refuseGardaJira(site domain.Site) error {
	if site.Role != "jsm_customer" {
		return nil
	}
	return domain.Usage("jira get/search is not available on garda").WithHint("use atlas jsm, not atlas jira search")
}
